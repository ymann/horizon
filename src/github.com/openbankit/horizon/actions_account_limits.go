package horizon

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/render/sse"
	"github.com/openbankit/horizon/resource"
)

// AccountLimitsAction detailed income/outcome limits for single account
type AccountLimitsAction struct {
	Action
	Address       string
	AccountLimits []history.AccountLimits
	Resource      resource.AccountLimits
}

// JSON is a method for actions.JSON
func (action *AccountLimitsAction) JSON() {
	action.Do(
		action.loadParams,
		action.loadRecord,
		action.loadResource,
		func() {
			hal.Render(action.W, action.Resource)
		},
	)
}

// SSE is a method for actions.SSE
func (action *AccountLimitsAction) SSE(stream sse.Stream) {
	// TODO: check
	action.Do(
		action.loadParams,
		action.loadRecord,
		action.loadResource,
		func() {
			stream.Send(sse.Event{Data: action.Resource})
		},
	)
}

func (action *AccountLimitsAction) loadParams() {
	action.Address = action.GetString("account_id")
}

func (action *AccountLimitsAction) loadRecord() {
	action.Err = action.HistoryQ().GetLimitsByAccount(&action.AccountLimits, action.Address)
	if action.Err != nil {
		return
	}
}

func (action *AccountLimitsAction) loadResource() {
	action.Err = action.Resource.Populate(
		action.Ctx,
		action.Address,
		action.AccountLimits,
	)
}
