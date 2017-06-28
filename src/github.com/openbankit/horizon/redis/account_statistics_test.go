package redis

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestAccountStatistics(t *testing.T) {

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	types := []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountDistributionAgent,
		xdr.AccountTypeAccountSettlementAgent,
		xdr.AccountTypeAccountExchangeAgent,
		xdr.AccountTypeAccountBank,
	}

	assetCode := "USD"

	err := Init(test.RedisURL())
	assert.Nil(t, err)

	conn := NewConnectionProvider().GetConnection()
	defer conn.Close()

	accountStatsProvider := NewAccountStatisticsProvider(conn)

	Convey("Does not exist", t, func() {
		account, err := keypair.Random()
		assert.Nil(t, err)
		stats, err := accountStatsProvider.Get(account.Address(), assetCode, []xdr.AccountType{
			xdr.AccountTypeAccountAnonymousUser,
			xdr.AccountTypeAccountBank,
		})
		So(err, ShouldBeNil)
		So(stats, ShouldBeNil)
	})
	Convey("Account stats storing", t, func() {
		account, err := keypair.Random()
		assert.Nil(t, err)
		accountStats := &AccountStatistics{
			Account:            account.Address(),
			AssetCode:          assetCode,
			Balance:            rand.Int63(),
			AccountsStatistics: make(map[xdr.AccountType]history.AccountStatistics),
		}
		for _, counterparty := range types {
			accountStats.AccountsStatistics[counterparty] = history.CreateRandomAccountStats(account.Address(), counterparty, assetCode)
		}
		expireTime := time.Duration(5) * time.Second
		err = accountStatsProvider.Insert(accountStats, expireTime)
		So(err, ShouldBeNil)
		storedAccountStats, err := accountStatsProvider.Get(account.Address(), assetCode, types)
		So(err, ShouldBeNil)
		assert.Equal(t, accountStats, storedAccountStats)
		// timeout expires
		time.Sleep(expireTime + time.Duration(1)*time.Second)
		storedAccountStats, err = accountStatsProvider.Get(account.Address(), assetCode, types)
		So(err, ShouldBeNil)
		So(storedAccountStats, ShouldBeNil)
	})
}
