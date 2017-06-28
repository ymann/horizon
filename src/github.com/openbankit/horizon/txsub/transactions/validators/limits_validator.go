package validators

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/redis"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	stat "github.com/openbankit/horizon/txsub/transactions/statistics"
	"database/sql"
	"fmt"
	"time"
)

type limitsValidator struct {
	paymentData      *statistics.PaymentData
	statsManager     statistics.ManagerInterface
	anonUserRest     config.AnonymousUserRestrictions
	accountStats     *redis.AccountStatistics
	historyQ         history.QInterface
	paymentDirection stat.PaymentDirection
	log              *log.Entry
	now              time.Time
}

func newLimitsValidator(paymentDirection stat.PaymentDirection, paymentData *stat.PaymentData,
	statsManager statistics.ManagerInterface, historyQ history.QInterface, anonUserRestr config.AnonymousUserRestrictions,
	now time.Time) *limitsValidator {
	return &limitsValidator{
		paymentData:      paymentData,
		statsManager:     statsManager,
		historyQ:         historyQ,
		anonUserRest:     anonUserRestr,
		paymentDirection: paymentDirection,
		log:              log.WithField("service", "limits_validator"),
		now:              now,
	}
}

func (v *limitsValidator) isIncoming() bool {
	return v.paymentDirection == stat.PaymentDirectionIncoming
}

func (v *limitsValidator) updateGetAccountStats() (*redis.AccountStatistics, error) {
	if v.accountStats != nil {
		return v.accountStats, nil
	}
	var err error
	v.accountStats, err = v.statsManager.UpdateGet(v.paymentData, v.paymentDirection, v.now)
	if err != nil {
		return nil, err
	}
	return v.accountStats, nil
}

func (v *limitsValidator) limitExceededDescription(periodName string, isAnonymous bool, outcome, limit int64) string {
	anonymous := ""
	if isAnonymous {
		anonymous = "anonymous "
	}
	return fmt.Sprintf("%s %s payments limit for %saccount exceeded: %s out of %s %s.",
		periodName,
		v.paymentDirection,
		anonymous,
		amount.String(xdr.Int64(xdr.Int64(outcome))),
		amount.String(xdr.Int64(limit)),
		v.paymentData.Asset.Code,
	)
}

func (v *limitsValidator) opMaxAmountExceededDescription(limit int64) string {
	return fmt.Sprintf(
		"Maximal operation amount for account (%s) exceeded: %s of %s %s",
		v.getAccount().Address,
		amount.String(xdr.Int64(v.paymentData.Amount)),
		amount.String(xdr.Int64(limit)),
		v.paymentData.Asset.Code,
	)
}

func (v *limitsValidator) getAccount() *history.Account {
	return v.paymentData.GetAccount(v.paymentDirection)
}

func (v *limitsValidator) getCounterparty() *history.Account {
	return v.paymentData.GetCounterparty(v.paymentDirection)
}

func (v *limitsValidator) GetAccountLimits() (*history.AccountLimits, error) {
	account := v.getAccount()
	limitedAssets, err := account.UnmarshalLimitedAssets()
	if err != nil {
		v.log.WithError(err).Error("Failed to unmarshal limited assets")
		return nil, err
	}
	if _, contains := limitedAssets[v.paymentData.Asset.Code]; !contains {
		return nil, nil
	}

	var limits history.AccountLimits
	err = v.historyQ.GetAccountLimits(&limits, account.Address, v.paymentData.Asset.Code)
	if err != nil {
		// no limits to check for sender
		if err == sql.ErrNoRows {
			v.log.Debug("No limits found")
			return nil, nil
		}
		return nil, err
	}
	return &limits, nil
}
