package commissions

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"database/sql"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestCommission(t *testing.T) {
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	Convey("countPercentFee", t, func() {
		percentFee := xdr.Int64(1 * amount.One) // fee is 1%
		Convey("amount too small", func() {
			fee := calculatePercentFee(xdr.Int64(1), percentFee)
			assert.Equal(t, xdr.Int64(0), fee)
		})
		Convey("amount is ok", func() {
			paymentAmount := 1230 * amount.One
			fee := calculatePercentFee(xdr.Int64(paymentAmount), percentFee)
			assert.Equal(t, xdr.Int64(12.3*amount.One), fee)
		})
		Convey("fee cutted", func() {
			paymentAmount := 156
			fee := calculatePercentFee(xdr.Int64(paymentAmount), percentFee)
			assert.Equal(t, xdr.Int64(1), fee)
		})
		Convey("fee cutted not rounded", func() {
			paymentAmount := 1560
			fee := calculatePercentFee(xdr.Int64(paymentAmount), percentFee)
			assert.Equal(t, xdr.Int64(15), fee)
		})
		Convey("amount is big", func() {
			paymentAmount := math.MaxInt64
			fee := calculatePercentFee(xdr.Int64(paymentAmount), percentFee)
			assert.Equal(t, xdr.Int64(paymentAmount/100), fee)
		})
	})
	Convey("get account type", t, func() {
		account, err := keypair.Random()
		assert.Nil(t, err)
		historyQMock := &history.QMock{}
		cm := New(&cache.SharedCache{
			AccountHistoryCache: cache.NewHistoryAccount(historyQMock),
		}, historyQMock)
		Convey("source does not exist", func() {
			historyQMock.On("AccountByAddress", account.Address()).Return(history.Account{}, sql.ErrNoRows)
			_, err := cm.getAccountType(account.Address(), true)
			assert.Equal(t, sql.ErrNoRows, err)
		})
		Convey("dest does not exist", func() {
			historyQMock.On("AccountByAddress", account.Address()).Return(history.Account{}, sql.ErrNoRows)
			accType, err := cm.getAccountType(account.Address(), false)
			assert.Nil(t, err)
			assert.Equal(t, int32(xdr.AccountTypeAccountAnonymousUser), accType)
		})
		Convey("source exists", func() {
			expectedType := xdr.AccountTypeAccountExchangeAgent
			historyQMock.On("AccountByAddress", account.Address()).Return(history.Account{
				AccountType: expectedType,
			}, nil)
			accType, err := cm.getAccountType(account.Address(), true)
			assert.Nil(t, err)
			assert.Equal(t, int32(expectedType), accType)
		})
		Convey("dest exists", func() {
			expectedType := xdr.AccountTypeAccountDistributionAgent
			historyQMock.On("AccountByAddress", account.Address()).Return(history.Account{
				AccountType: expectedType,
			}, nil)
			accType, err := cm.getAccountType(account.Address(), false)
			assert.Nil(t, err)
			assert.Equal(t, int32(expectedType), accType)
		})
	})
	Convey("get smallest", t, func() {
		comms := []history.Commission{
			history.Commission{
				KeyHash:    "hash",
				KeyValue:   "{}",
				FlatFee:    int64(20000000),
				PercentFee: int64(40000000),
			},
			history.Commission{
				KeyHash:    "hash",
				KeyValue:   "{}",
				FlatFee:    int64(20000000),
				PercentFee: int64(400000000),
			},
		}
		comm := getSmallestFee(comms, xdr.Int64(1000000000))
		assert.NotNil(t, comm)
		assert.Equal(t, comms[0], *comm)
	})
}
