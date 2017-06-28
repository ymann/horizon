package sqx

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
	"database/sql"
	"math/rand"
)

func TestBatchUpdate(t *testing.T) {
	log.DefaultLogger.Logger.Level = log.DebugLevel
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	repo := tt.HorizonRepo()

	account := test.BankMasterSeed().Address()
	counterparty := xdr.AccountTypeAccountBank
	Convey("", t, func() {
		assetCode := "GBP"
		expected := history.CreateRandomAccountStats(account, counterparty, assetCode)
		update := BatchUpdate(BatchInsertFromInsert(repo, history.AccountStatisticsInsert), history.AccountStatisticsUpdateParams, history.AccountStatisticsUpdateWhere)
		update.Insert(&expected)

		q := &history.Q{Repo: repo}
		_, err := q.GetAccountStatistics(account, assetCode, counterparty)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, sql.ErrNoRows)

		// Update won't trigger
		newExpected := expected
		newExpected.AnnualIncome = 0
		newExpected.DailyIncome = rand.Int63()
		update.Update(&newExpected)

		_, err = q.GetAccountStatistics(account, assetCode, counterparty)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, sql.ErrNoRows)

		update.Flush()
		So(update.Err, ShouldBeNil)
		actual, err := q.GetAccountStatistics(account, assetCode, counterparty)
		So(err, ShouldBeNil)
		newExpected.UpdatedAt = actual.UpdatedAt
		assert.Equal(t, newExpected, actual)

		// new update will be flushed
		newExpected.AnnualOutcome = rand.Int63()
		err = update.Update(&newExpected)
		So(update.Err, ShouldBeNil)
		actual, err = q.GetAccountStatistics(account, assetCode, counterparty)
		So(err, ShouldBeNil)
		newExpected.UpdatedAt = actual.UpdatedAt
		assert.Equal(t, newExpected, actual)
	})
}
