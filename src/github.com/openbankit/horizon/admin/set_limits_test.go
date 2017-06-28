package admin

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestActionsSetLimits(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	historyQ := &history.Q{tt.HorizonRepo()}
	Convey("Set limits", t, func() {
		Convey("Invalid account", func() {
			action := NewSetLimitsAction(NewAdminAction(map[string]interface{}{
				"account_id": "invalid_id",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "account_id")
		})
		Convey("account does not exist", func() {
			newAccount, err := keypair.Random()
			So(err, ShouldBeNil)
			action := NewSetLimitsAction(NewAdminAction(map[string]interface{}{
				"account_id": newAccount.Address(),
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, problem.ShouldBeProblem, problem.NotFound)
		})
		Convey("happy path", func() {
			account := test.NewTestConfig().BankMasterKey
			// create new limit
			expected := history.AccountLimits{
				AssetCode:       "USD",
				Account:         account,
				MaxOperationOut: 1,
				DailyMaxOut:     2,
				MonthlyMaxOut:   3,
				MaxOperationIn:  5,
				DailyMaxIn:      7,
				MonthlyMaxIn:    11,
			}
			data := limitsToMap(expected)
			applyLimit(t, data, expected, historyQ)
			// update
			expected.DailyMaxIn = 13
			expected.MaxOperationOut = 17
			data = limitsToMap(expected)
			applyLimit(t, data, expected, historyQ)
			var storedAccount history.Account
			err := historyQ.AccountByAddress(&storedAccount, account)
			So(err, ShouldBeNil)
			limitedAssets, err := storedAccount.UnmarshalLimitedAssets()
			So(err, ShouldBeNil)
			_, ok := limitedAssets[expected.AssetCode]
			So(ok, ShouldBeTrue)
		})
	})
}

func limitsToMap(l history.AccountLimits) map[string]interface{} {
	return map[string]interface{}{
		"account_id":        l.Account,
		"asset_code":        l.AssetCode,
		"max_operation_out": strconv.Itoa(int(l.MaxOperationOut)),
		"daily_max_out":     strconv.Itoa(int(l.DailyMaxOut)),
		"monthly_max_out":   strconv.Itoa(int(l.MonthlyMaxOut)),
		"max_operation_in":  strconv.Itoa(int(l.MaxOperationIn)),
		"daily_max_in":      strconv.Itoa(int(l.DailyMaxIn)),
		"monthly_max_in":    strconv.Itoa(int(l.MonthlyMaxIn)),
	}
}

func applyLimit(t *testing.T, data map[string]interface{}, expected history.AccountLimits, historyQ *history.Q) {
	action := NewSetLimitsAction(NewAdminAction(data, historyQ))
	action.Validate()
	So(action.Err, ShouldBeNil)
	action.Apply()
	So(action.Err, ShouldBeNil)
	var limits history.AccountLimits
	err := historyQ.GetAccountLimits(&limits, data["account_id"].(string), expected.AssetCode)
	if err != nil {
		log.WithField("account_id", data["account_id"]).WithError(err).Error("failed to get account limits")
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, limits)
}
