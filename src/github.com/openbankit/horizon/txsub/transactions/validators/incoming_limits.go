package validators

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/helpers"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	"fmt"
	"time"
)

type IncomingLimitsValidatorInterface interface {
	VerifyLimits() (*results.ExceededLimitError, error)
}

type IncomingLimitsValidator struct {
	limitsValidator
	statsManager statistics.ManagerInterface

	dailyIncome   *int64
	monthlyIncome *int64
	balance       *int64
}

func NewIncomingLimitsValidator(paymentData *statistics.PaymentData, historyQ history.QInterface,
	statsManager statistics.ManagerInterface, anonUserRestr config.AnonymousUserRestrictions, now time.Time) *IncomingLimitsValidator {

	limitsValidator := newLimitsValidator(statistics.PaymentDirectionIncoming, paymentData, statsManager, historyQ,
		anonUserRestr, now)
	result := &IncomingLimitsValidator{
		limitsValidator: *limitsValidator,
		statsManager:    statsManager,
	}
	result.log = log.WithField("service", "incoming_limits_validator")
	return result
}

// VerifyLimits checks incoming limits
func (v *IncomingLimitsValidator) VerifyLimits() (*results.ExceededLimitError, error) {
	// check account's limits
	result, err := v.verifyReceiverAccountLimits()
	if result != nil || err != nil {
		return result, err
	}

	return v.verifyAnonymousAssetLimits()
}

func (v *IncomingLimitsValidator) verifyReceiverAccountLimits() (*results.ExceededLimitError, error) {
	limits, err := v.GetAccountLimits()
	if err != nil || limits == nil {
		return nil, err
	}

	v.log.WithField("limits", limits).Debug("Checking limits")
	if limits.MaxOperationIn >= 0 && v.paymentData.Amount > limits.MaxOperationIn {
		description := v.opMaxAmountExceededDescription(limits.MaxOperationIn)
		return &results.ExceededLimitError{Description: description}, nil
	}

	if limits.DailyMaxIn >= 0 {
		updatedDailyIncome, err := v.getUpdatedDailyIncome()
		if err != nil {
			return nil, err
		}

		v.log.WithFields(log.F{
			"newIncome": updatedDailyIncome,
			"limit":     limits.DailyMaxIn,
		}).Debug("Checking daily income for limits")
		if updatedDailyIncome > limits.DailyMaxIn {
			description := v.limitExceededDescription("Daily", false, updatedDailyIncome, limits.DailyMaxIn)
			return &results.ExceededLimitError{Description: description}, nil
		}
	}

	if limits.MonthlyMaxIn >= 0 {
		updatedMonthlyIncome, err := v.getUpdatedMonthlyIncome()
		if err != nil {
			return nil, err
		}
		v.log.WithFields(log.F{
			"newIncome": updatedMonthlyIncome,
			"limit":     limits.MonthlyMaxIn,
		}).Debug("Checking daily income for limits")
		if updatedMonthlyIncome > limits.MonthlyMaxIn {
			description := v.limitExceededDescription("Monthly", false, updatedMonthlyIncome, limits.MonthlyMaxIn)
			return &results.ExceededLimitError{Description: description}, nil
		}
	}
	return nil, nil
}

// VerifyLimitsForReceiver checks limits  and restrictions for receiver
func (v *IncomingLimitsValidator) verifyAnonymousAssetLimits() (*results.ExceededLimitError, error) {
	if !v.paymentData.Asset.IsAnonymous || !helpers.IsUser(v.getAccount().AccountType) {
		// Nothing to be checked
		return nil, nil
	}

	updatedBalance, err := v.getUpdatedBalance()
	if err != nil {
		return nil, err
	}

	if updatedBalance > v.anonUserRest.MaxBalance {
		description := fmt.Sprintf(
			"User's max balance exceeded: %s + %s out of %s UAH.",
			amount.String(xdr.Int64(updatedBalance-v.paymentData.Amount)),
			amount.String(xdr.Int64(v.paymentData.Amount)),
			amount.String(xdr.Int64(v.anonUserRest.MaxBalance)),
		)
		return &results.ExceededLimitError{Description: description}, nil
	}
	return nil, nil
}

func (v *IncomingLimitsValidator) getUpdatedDailyIncome() (int64, error) {
	if v.dailyIncome != nil {
		return *v.dailyIncome, nil
	}
	v.dailyIncome = new(int64)
	stats, err := v.updateGetAccountStats()
	if err != nil {
		return 0, err
	}

	*v.dailyIncome = helpers.SumAccountStats(
		stats.AccountsStatistics,
		func(stats *history.AccountStatistics) int64 { return stats.DailyIncome },
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountSettlementAgent,
	)
	return *v.dailyIncome, nil
}

func (v *IncomingLimitsValidator) getUpdatedMonthlyIncome() (int64, error) {
	if v.monthlyIncome != nil {
		return *v.monthlyIncome, nil
	}
	v.monthlyIncome = new(int64)
	stats, err := v.updateGetAccountStats()
	if err != nil {
		return 0, err
	}

	*v.monthlyIncome = helpers.SumAccountStats(
		stats.AccountsStatistics,
		func(stats *history.AccountStatistics) int64 { return stats.MonthlyIncome },
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountSettlementAgent,
	)
	return *v.monthlyIncome, nil
}

func (v *IncomingLimitsValidator) getUpdatedBalance() (int64, error) {
	if v.balance != nil {
		return *v.balance, nil
	}
	stats, err := v.updateGetAccountStats()
	if err != nil {
		return 0, err
	}
	v.balance = new(int64)
	return stats.Balance, nil
}
