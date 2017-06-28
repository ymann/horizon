package validators

import (
	"testing"

	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/redis"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
	"database/sql"
	"fmt"
	"github.com/guregu/null"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestIncomingLimits(t *testing.T) {
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	sourceKey, err := keypair.Random()
	assert.Nil(t, err)
	destKey, err := keypair.Random()
	assert.Nil(t, err)
	source := &history.Account{
		Address:     sourceKey.Address(),
		AccountType: xdr.AccountTypeAccountAnonymousUser,
	}
	opAsset := history.Asset{
		Code:        "UAH",
		IsAnonymous: false,
	}
	destination := &history.Account{
		Address:       destKey.Address(),
		AccountType:   xdr.AccountTypeAccountAnonymousUser,
		LimitedAssets: null.StringFrom(fmt.Sprintf("{\"%s\":true}", opAsset.Code)),
	}
	opAmount := int64(amount.One * 100)

	opData := statistics.NewOperationData(source, 0, "random_tx_hash")
	paymentData := statistics.NewPaymentData(destination, nil, opAsset, opAmount, opData)
	direction := statistics.PaymentDirectionIncoming

	accountLimits := history.AccountLimits{
		Account:         paymentData.GetAccount(direction).Address,
		AssetCode:       opAsset.Code,
		MaxOperationOut: -1,
		DailyMaxOut:     -1,
		MonthlyMaxOut:   -1,
		MaxOperationIn:  -1,
		DailyMaxIn:      -1,
		MonthlyMaxIn:    -1,
	}

	statsManager := &statistics.ManagerMock{}

	now := time.Now()
	Convey("Incoming limits test:", t, func() {
		Convey("No limits for account & asset is not anonymous", func() {
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(nil, sql.ErrNoRows)
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, config.AnonymousUserRestrictions{}, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})
		Convey("All limits are empty for account & asset is not anonymous", func() {
			histMock := history.QMock{}
			limits := accountLimits
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(limits, nil)
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, config.AnonymousUserRestrictions{}, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})
		Convey("Asset is not anonymous, exceeds op amount", func() {
			limits := accountLimits
			limits.MaxOperationIn = opAmount - 1
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(limits, nil)
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, config.AnonymousUserRestrictions{}, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			assert.Equal(t, &results.ExceededLimitError{Description: fmt.Sprintf(
				"Maximal operation amount for account (%s) exceeded: %s of %s %s",
				paymentData.GetAccount(direction).Address,
				amount.String(xdr.Int64(opAmount)),
				amount.String(xdr.Int64(limits.MaxOperationIn)),
				opAsset.Code,
			)}, result)
		})
		Convey("Asset is not anonymous, exceeds daily limit with stats", func() {
			limits := accountLimits
			limits.DailyMaxIn = 2*opAmount - 1
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(limits, nil)
			stats := &redis.AccountStatistics{
				Balance: 0,
				AccountsStatistics: map[xdr.AccountType]history.AccountStatistics{
					paymentData.GetAccount(direction).AccountType: history.AccountStatistics{
						Account:          paymentData.GetAccount(direction).Address,
						AssetCode:        opAsset.Code,
						CounterpartyType: int16(paymentData.GetCounterparty(direction).AccountType),
						DailyIncome:      opAmount + opAmount,
					},
				},
			}
			statsManager.On("UpdateGet", &paymentData, direction, now).Return(stats, nil).Once()
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, config.AnonymousUserRestrictions{}, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			assert.Equal(t, &results.ExceededLimitError{Description: fmt.Sprintf("Daily incoming payments limit for account exceeded: %s out of %s %s.",
				amount.String(xdr.Int64(opAmount+opAmount)),
				amount.String(xdr.Int64(limits.DailyMaxIn)),
				opAsset.Code,
			)}, result)
		})
		Convey("Asset is not anonymous, exceeds monthly limit with empty stats", func() {
			limits := accountLimits
			limits.MonthlyMaxIn = 2*opAmount - 1
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(limits, nil)
			stats := &redis.AccountStatistics{
				Balance: 0,
				AccountsStatistics: map[xdr.AccountType]history.AccountStatistics{
					xdr.AccountTypeAccountAnonymousUser: history.AccountStatistics{
						Account:          paymentData.GetAccount(direction).Address,
						AssetCode:        opAsset.Code,
						CounterpartyType: int16(xdr.AccountTypeAccountSettlementAgent),
						MonthlyIncome:    opAmount + opAmount,
					},
				},
			}
			statsManager.On("UpdateGet", &paymentData, direction, now).Return(stats, nil).Once()
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, config.AnonymousUserRestrictions{}, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			assert.Equal(t, &results.ExceededLimitError{Description: fmt.Sprintf("Monthly incoming payments limit for account exceeded: %s out of %s %s.",
				amount.String(xdr.Int64(opAmount+opAmount)),
				amount.String(xdr.Int64(limits.MonthlyMaxIn)),
				opAsset.Code,
			)}, result)
		})
		stats := &redis.AccountStatistics{
			Balance: 0,
			AccountsStatistics: map[xdr.AccountType]history.AccountStatistics{
				xdr.AccountTypeAccountAnonymousUser: history.AccountStatistics{
					Account:          paymentData.GetAccount(direction).Address,
					AssetCode:        opAsset.Code,
					CounterpartyType: int16(xdr.AccountTypeAccountSettlementAgent),
					DailyIncome:      opAmount,
					MonthlyIncome:    opAmount,
					AnnualIncome:     opAmount,
				},
			},
		}
		Convey("Asset is anonymous exceeds max balance", func() {
			limits := config.AnonymousUserRestrictions{
				MaxBalance: 2*opAmount - 1,
			}
			stats.Balance = 2 * opAmount
			paymentData.Asset.IsAnonymous = true
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(nil, sql.ErrNoRows)
			statsManager.On("UpdateGet", &paymentData, direction, now).Return(stats, nil).Once()
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, limits, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			assert.Equal(t, &results.ExceededLimitError{Description: fmt.Sprintf(
				"User's max balance exceeded: %s + %s out of %s UAH.",
				amount.String(xdr.Int64(stats.Balance - opAmount)),
				amount.String(xdr.Int64(opAmount)),
				amount.String(xdr.Int64(limits.MaxBalance)),
			)}, result)
		})
		Convey("Asset is anonymous exceeds max balance, but is not user", func() {
			limits := config.AnonymousUserRestrictions{
				MaxBalance: 2*opAmount - 1,
			}
			stats.Balance = opAmount
			opAsset.IsAnonymous = true
			histMock := history.QMock{}
			histMock.On("GetAccountLimits", paymentData.GetAccount(direction).Address, opAsset.Code).Return(nil, sql.ErrNoRows)
			statsManager.On("UpdateGet", &paymentData, direction, now).Return(stats, nil).Once()
			paymentData.GetAccount(direction).AccountType = xdr.AccountTypeAccountMerchant
			v := NewIncomingLimitsValidator(&paymentData, &histMock, statsManager, limits, now)
			result, err := v.VerifyLimits()
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

	})
}
