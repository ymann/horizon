package validators

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/helpers"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	"time"
)

type OutgoingLimitsValidatorInterface interface {
	VerifyLimits() (*results.ExceededLimitError, error)
}

type OutgoingLimitsValidator struct {
	limitsValidator
	dailyOutcome   *int64
	monthlyOutcome *int64
}

func NewOutgoingLimitsValidator(paymentData *statistics.PaymentData, statsManager statistics.ManagerInterface, historyQ history.QInterface, anonUserRestr config.AnonymousUserRestrictions, now time.Time) *OutgoingLimitsValidator {
	limitsValidator := newLimitsValidator(statistics.PaymentDirectionOutgoing, paymentData, statsManager, historyQ, anonUserRestr, now)
	result := &OutgoingLimitsValidator{
		limitsValidator: *limitsValidator,
	}
	result.log = log.WithField("service", "outgoing_limits_validator")
	return result
}

// VerifyLimits checks outgoing limits
func (v *OutgoingLimitsValidator) VerifyLimits() (*results.ExceededLimitError, error) {
	// check account's limits
	result, err := v.verifySenderAccountLimits()
	if result != nil || err != nil {
		return result, err
	}

	return v.verifyAnonymousAssetLimits()
}

// Checks limits for sender
func (v *OutgoingLimitsValidator) verifySenderAccountLimits() (*results.ExceededLimitError, error) {
	limits, err := v.GetAccountLimits()
	if err != nil || limits == nil {
		return nil, err
	}

	v.log.WithField("limits", limits).Debug("Checking limits")
	if limits.MaxOperationOut >= 0 && v.paymentData.Amount > limits.MaxOperationOut {
		description := v.opMaxAmountExceededDescription(limits.MaxOperationOut)
		return &results.ExceededLimitError{Description: description}, nil
	}

	if limits.DailyMaxOut >= 0 {
		updatedDailyOutcome, err := v.getUpdatedDailyOutcome()
		if err != nil {
			return nil, err
		}
		v.log.WithFields(log.F{
			"newOutcome": updatedDailyOutcome,
			"limit":      limits.DailyMaxOut,
		}).Debug("Checking daily outcome for limits")
		if updatedDailyOutcome > limits.DailyMaxOut {
			description := v.limitExceededDescription("Daily", false, updatedDailyOutcome, limits.DailyMaxOut)
			return &results.ExceededLimitError{Description: description}, nil
		}
	}

	if limits.MonthlyMaxOut >= 0 {
		updatedMonthlyOutcome, err := v.getUpdatedMonthlyOutcome()
		if err != nil {
			return nil, err
		}
		v.log.WithFields(log.F{
			"newOutcome": updatedMonthlyOutcome,
			"limit":      limits.MonthlyMaxOut,
		}).Debug("Checking daily outcome for limits")
		if updatedMonthlyOutcome > limits.MonthlyMaxOut {
			description := v.limitExceededDescription("Monthly", false, updatedMonthlyOutcome, limits.MonthlyMaxOut)
			return &results.ExceededLimitError{Description: description}, nil
		}
	}
	return nil, nil
}

// checks limits for anonymous asset
func (v *OutgoingLimitsValidator) verifyAnonymousAssetLimits() (*results.ExceededLimitError, error) {
	if !v.paymentData.Asset.IsAnonymous || !helpers.IsUser(v.getAccount().AccountType) {
		// Nothing to be checked
		return nil, nil
	}
	// check anonymous asset limits
	// daily and monthly limits are not applied for payments to merchant
	if v.getCounterparty().AccountType != xdr.AccountTypeAccountMerchant {
		// 1. Check daily outcome
		if v.anonUserRest.MaxDailyOutcome >= 0 {
			updatedDailyOutcome, err := v.getUpdatedDailyOutcome()
			if err != nil {
				return nil, err
			}

			if updatedDailyOutcome > v.anonUserRest.MaxDailyOutcome {
				description := v.limitExceededDescription("Daily", true, updatedDailyOutcome, v.anonUserRest.MaxDailyOutcome)
				return &results.ExceededLimitError{Description: description}, nil
			}
		}

		// 2. Check monthly outcome
		if v.anonUserRest.MaxMonthlyOutcome >= 0 {
			updateMonthlyOutcome, err := v.getUpdatedMonthlyOutcome()
			if err != nil {
				return nil, err
			}

			if updateMonthlyOutcome > v.anonUserRest.MaxMonthlyOutcome {
				description := v.limitExceededDescription("Monthly", true, updateMonthlyOutcome, v.anonUserRest.MaxMonthlyOutcome)
				return &results.ExceededLimitError{Description: description}, nil
			}
		}
	}

	// annualOutcome does not count for payments to settlement agent
	if v.getCounterparty().AccountType != xdr.AccountTypeAccountSettlementAgent && v.anonUserRest.MaxAnnualOutcome >= 0 {
		// 3. Check annual outcome
		stats, err := v.updateGetAccountStats()
		if err != nil {
			return nil, err
		}

		updatedAnnualOutcome := helpers.SumAccountStats(
			stats.AccountsStatistics,
			func(stats *history.AccountStatistics) int64 { return stats.AnnualOutcome },
			xdr.AccountTypeAccountAnonymousUser,
			xdr.AccountTypeAccountRegisteredUser,
			xdr.AccountTypeAccountMerchant,
		)

		if updatedAnnualOutcome > v.anonUserRest.MaxAnnualOutcome {
			description := v.limitExceededDescription("Annual", true, updatedAnnualOutcome, v.anonUserRest.MaxAnnualOutcome)
			return &results.ExceededLimitError{Description: description}, nil
		}
	}

	return nil, nil
}

func (v *OutgoingLimitsValidator) getUpdatedDailyOutcome() (int64, error) {
	if v.dailyOutcome != nil {
		return *v.dailyOutcome, nil
	}
	v.dailyOutcome = new(int64)
	stats, err := v.updateGetAccountStats()
	if err != nil {
		return 0, err
	}

	*v.dailyOutcome = helpers.SumAccountStats(
		stats.AccountsStatistics,
		func(stats *history.AccountStatistics) int64 { return stats.DailyOutcome },
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountSettlementAgent,
	)
	return *v.dailyOutcome, nil
}

func (v *OutgoingLimitsValidator) getUpdatedMonthlyOutcome() (int64, error) {
	if v.monthlyOutcome != nil {
		return *v.monthlyOutcome, nil
	}
	v.monthlyOutcome = new(int64)
	stats, err := v.updateGetAccountStats()
	if err != nil {
		return 0, err
	}

	*v.monthlyOutcome = helpers.SumAccountStats(
		stats.AccountsStatistics,
		func(stats *history.AccountStatistics) int64 { return stats.MonthlyOutcome },
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountSettlementAgent,
	)
	return *v.monthlyOutcome, nil
}
