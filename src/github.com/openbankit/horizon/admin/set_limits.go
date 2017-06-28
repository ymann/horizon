package admin

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"database/sql"
)

type SetLimitsAction struct {
	AdminAction
	Account *history.Account
	Limits history.AccountLimits
}

func NewSetLimitsAction(adminAction AdminAction) *SetLimitsAction {
	return &SetLimitsAction{
		AdminAction: adminAction,
	}
}

func (action *SetLimitsAction) Validate() {
	action.loadParams()
	if action.HasError() {
		return
	}

	// 1. Check if account exists
	action.Account = new(history.Account)
	err := action.HistoryQ().AccountByAddress(action.Account, action.Limits.Account)

	if err != nil {
		if err == sql.ErrNoRows {
			action.Err = &problem.NotFound
			return
		}
		action.Log.WithStack(err).WithError(err).Error("Failed to load account by address")
		action.Err = &problem.ServerError
		return
	}
}

func (action *SetLimitsAction) Apply() {
	if action.Err != nil {
		return
	}
	// 2. Try get limits for account
	var isNewEntry bool
	var accLimits history.AccountLimits
	err := action.HistoryQ().GetAccountLimits(&accLimits, action.Limits.Account, action.Limits.AssetCode)
	if err != nil {
		if err != sql.ErrNoRows {
			action.Log.WithStack(err).WithError(err).Error("Failed to get account limits")
			action.Err = &problem.ServerError
			return
		}
		isNewEntry = true
	}
	// 3. Validate and set limits
	accLimits = action.Limits

	// 4. Persist changes
	if isNewEntry {
		var limitedAssets map[string]bool
		limitedAssets, err = action.Account.UnmarshalLimitedAssets()
		if err != nil {
			action.Log.WithStack(err).WithError(err).Error("Failed to unmarshal limited assets")
			action.Err = &problem.ServerError
			return
		}
		limitedAssets[accLimits.AssetCode] = true
		action.Account.SetLimitedAssets(limitedAssets)
		err = action.HistoryQ().AccountUpdate(action.Account)
		if err != nil {
			action.Log.WithStack(err).WithError(err).Error("Failed to update account's limited assets")
			action.Err = &problem.ServerError
			return
		}

		err = action.HistoryQ().CreateAccountLimits(accLimits)
	} else {
		err = action.HistoryQ().UpdateAccountLimits(accLimits)
	}

	if err != nil {
		action.Log.WithStack(err).WithField("is_new", isNewEntry).WithError(err).Error("Failed to insert/update account limits")
		action.Err = &problem.ServerError
	}
}

func (action *SetLimitsAction) loadParams() {
	action.Limits.Account = action.GetAddress("account_id")
	action.Limits.AssetCode = action.GetString("asset_code")
	action.Limits.MaxOperationOut = action.GetInt64("max_operation_out")
	action.Limits.DailyMaxOut = action.GetInt64("daily_max_out")
	action.Limits.MonthlyMaxOut = action.GetInt64("monthly_max_out")
	action.Limits.MaxOperationIn = action.GetInt64("max_operation_in")
	action.Limits.DailyMaxIn = action.GetInt64("daily_max_in")
	action.Limits.MonthlyMaxIn = action.GetInt64("monthly_max_in")
}
