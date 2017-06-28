package admin

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"database/sql"
)

type SetTraitsAction struct {
	AdminAction
	Address  string
	BlockIn  *bool
	BlockOut *bool

	account  history.Account
}

func NewSetTraitsAction(adminAction AdminAction) *SetTraitsAction {
	return &SetTraitsAction{
		AdminAction: adminAction,
	}
}

func (action *SetTraitsAction) Validate() {
	action.loadParams()
	if action.Err != nil {
		return
	}

	err := action.HistoryQ().AccountByAddress(&action.account, action.Address)
	if err != nil {
		if err != sql.ErrNoRows {
			action.Err = &problem.ServerError
			action.Log.WithStack(err).WithError(err).Error("Failed to get account")
			return
		}
		action.Err = &problem.NotFound
		return
	}

	//Set traits
	if action.BlockIn != nil {
		action.account.BlockIncomingPayments = *action.BlockIn
	}

	if action.BlockOut != nil {
		action.account.BlockOutcomingPayments = *action.BlockOut
	}
}

func (action *SetTraitsAction) Apply() {
	if action.Err != nil {
		return
	}

	action.Err = action.HistoryQ().AccountUpdate(&action.account)
}

func (action *SetTraitsAction) loadParams() {
	action.Address = action.GetAddress("account_id")
	action.BlockIn = action.GetOptionalBool("block_incoming_payments")
	action.BlockOut = action.GetOptionalBool("block_outcoming_payments")
}
