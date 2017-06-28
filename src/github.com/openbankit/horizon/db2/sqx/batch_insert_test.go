package sqx

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/lann/squirrel"
	"database/sql"
	"math/rand"
)

func TestBatchInsert(t *testing.T) {
	log.DefaultLogger.Logger.Level = log.DebugLevel
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	repo := tt.HorizonRepo()
	Convey("Empty table", t, func() {
		insert := BatchInsertFromInsert(repo, squirrel.Insert(""))
		So(insert.Err.Error(), ShouldEqual, "Table must be set")
	})
	Convey("Empty columns", t, func() {
		insert := BatchInsertFromInsert(repo, squirrel.Insert("table"))
		So(insert.Err.Error(), ShouldEqual, "Invalid builder. Columns must be set")
	})
	Convey("Flush of empty insert does not creates error", t, func() {
		insert := BatchInsertFromInsert(repo, history.AccountStatisticsInsert)
		insert.Flush()
		So(insert.Err, ShouldBeNil)
	})
	Convey("Can insert and flush after flush", t, func() {
		insert := BatchInsertFromInsert(repo, history.AccountStatisticsInsert)
		insertFlushCheck(insert, "USD", repo, t)
		insertFlushCheck(insert, "EUR", repo, t)
		Convey("Can flush empty", func() {
			insert.Flush()
			So(insert.Err, ShouldBeNil)
		})
	})
	account := test.BankMasterSeed().Address()
	counterparty := xdr.AccountTypeAccountBank
	Convey("Insert of object, update and insert will trigger only one insert", t, func() {
		assetCode := "GBP"
		expected := history.CreateRandomAccountStats(account, counterparty, assetCode)
		insert := BatchInsertFromInsert(repo, history.AccountStatisticsInsert)
		insert.Insert(&expected)
		newExpected := expected
		newExpected.AnnualIncome = 0
		newExpected.DailyIncome = rand.Int63()
		insert.Insert(&newExpected)
		insert.Flush()
		So(insert.Err, ShouldBeNil)
		q := &history.Q{Repo: repo}
		actual, err := q.GetAccountStatistics(account, assetCode, counterparty)
		So(err, ShouldBeNil)
		newExpected.UpdatedAt = actual.UpdatedAt
		assert.Equal(t, newExpected, actual)
	})
	Convey("auto flush", t, func() {
		s1 := history.CreateRandomAccountStats(account, counterparty, "AAA")
		paramsSize := len(s1.GetParams())
		insert := BatchInsertFromInsert(repo, history.AccountStatisticsInsert)
		insert.BatchSize(paramsSize + 1)
		insert.Insert(&s1)
		So(insert.Err, ShouldBeNil)
		So(insert.NeedFlush(), ShouldBeTrue)
		q := &history.Q{Repo: repo}
		// Update won't trigger flush
		s1.DailyIncome = rand.Int63()
		insert.Insert(&s1)
		So(insert.Err, ShouldBeNil)
		_, err := q.GetAccountStatistics(account, s1.AssetCode, counterparty)
		So(err, ShouldNotBeNil)
		assert.Equal(t, sql.ErrNoRows, err)
		// second insert triggers flush before it
		s2 := history.CreateRandomAccountStats(account, counterparty, "AAB")
		insert.Insert(&s2)
		So(insert.Err, ShouldBeNil)
		// s1 is in flushed
		actualS1, err := q.GetAccountStatistics(account, s1.AssetCode, counterparty)
		So(err, ShouldBeNil)
		s1.UpdatedAt = actualS1.UpdatedAt
		assert.Equal(t, s1, actualS1)
		// s2 is not
		_, err = q.GetAccountStatistics(account, s2.AssetCode, counterparty)
		assert.Equal(t, sql.ErrNoRows, err)
		insert.Flush()
		So(insert.Err, ShouldBeNil)
		actualS2, err := q.GetAccountStatistics(account, s2.AssetCode, counterparty)
		s2.UpdatedAt = actualS2.UpdatedAt
		assert.Equal(t, s2, actualS2)
	})
}

func insertFlushCheck(insert *BatchInsertBuilder, assetCode string, repo *db2.Repo, t *testing.T) {
	account := test.BankMasterSeed().Address()
	counterparty := xdr.AccountTypeAccountBank
	expected := history.CreateRandomAccountStats(account, counterparty, assetCode)
	insert.Insert(&expected)
	So(insert.Err, ShouldBeNil)
	insert.Flush()
	So(insert.Err, ShouldBeNil)
	q := &history.Q{Repo: repo}
	actual, err := q.GetAccountStatistics(account, assetCode, counterparty)
	So(err, ShouldBeNil)
	expected.UpdatedAt = actual.UpdatedAt
	assert.Equal(t, expected, actual)
}
