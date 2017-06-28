package admin

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActionsSetTraits(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	historyQ := &history.Q{tt.HorizonRepo()}
	account := test.NewTestConfig().BankMasterKey

	Convey("Invalid params", t, func() {
		Convey("Invalid account", func() {
			action := NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id": "invalid_id",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "account_id")
		})
		Convey("Invalid block_incoming_payments", func() {
			action := NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":              account,
				"block_incoming_payments": "not_bool",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "block_incoming_payments")
		})
		Convey("Invalid block_outcoming_payments", func() {
			action := NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":               account,
				"block_outcoming_payments": "not_bool",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "block_outcoming_payments")
		})
	})
	Convey("Set traits", t, func() {
		Convey("account does not exist", func() {
			newAccount, err := keypair.Random()
			assert.Nil(t, err)
			action := NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":               newAccount.Address(),
				"block_outcoming_payments": "not_bool",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "block_outcoming_payments")
		})
		Convey("happy path", func() {
			// create new trait
			var storedAcc history.Account
			err := historyQ.AccountByAddress(&storedAcc, account)
			So(err, ShouldBeNil)
			So(storedAcc.BlockIncomingPayments, ShouldBeFalse)
			So(storedAcc.BlockOutcomingPayments, ShouldBeFalse)
			storedAcc.BlockIncomingPayments = true
			action := NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":              account,
				"block_incoming_payments": "true",
			}, historyQ))
			checkTraitsAction(action, storedAcc, historyQ)
			// update
			storedAcc.BlockOutcomingPayments = true
			action = NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":               account,
				"block_outcoming_payments": "true",
			}, historyQ))
			checkTraitsAction(action, storedAcc, historyQ)
			// remove
			storedAcc.BlockOutcomingPayments = false
			storedAcc.BlockIncomingPayments = false
			action = NewSetTraitsAction(NewAdminAction(map[string]interface{}{
				"account_id":               account,
				"block_incoming_payments":  "false",
				"block_outcoming_payments": "false",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldBeNil)
			action.Apply()
			So(action.Err, ShouldBeNil)
			checkTraitsAction(action, storedAcc, historyQ)
		})
	})
}

func checkTraitsAction(action *SetTraitsAction,expected history.Account, historyQ *history.Q) {
	action.Validate()
	So(action.Err, ShouldBeNil)
	action.Apply()
	So(action.Err, ShouldBeNil)
	var actual history.Account
	err := historyQ.AccountByAddress(&actual, expected.Address)
	So(err, ShouldBeNil)
	So(actual.TotalOrderID.ID, ShouldEqual, expected.TotalOrderID.ID)
	So(actual.BlockIncomingPayments, ShouldEqual, expected.BlockIncomingPayments)
	So(actual.BlockOutcomingPayments, ShouldEqual, expected.BlockOutcomingPayments)
}
