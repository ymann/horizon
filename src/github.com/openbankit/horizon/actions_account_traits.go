package horizon

import (
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/render/sse"
	"github.com/openbankit/horizon/resource"
)

// This file contains the actions:
//
// AccountTraitsIndexAction: pages of account's traits in order of creation of accounts
type AccountTraitsIndexAction struct {
	Action
	PagingParams db2.PageQuery
	Records      []history.Account
	Page         hal.Page
}

// JSON is a method for actions.JSON
func (action *AccountTraitsIndexAction) JSON() {
	action.Do(
		action.loadParams,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
}

// SSE is a method for actions.SSE
func (action *AccountTraitsIndexAction) SSE(stream sse.Stream) {
	action.Setup(action.loadParams)
	action.Do(
		action.loadRecords,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			var res resource.AccountTraits
			for _, record := range action.Records[stream.SentCount():] {
				res.Populate(action.Ctx, record)
				stream.Send(sse.Event{ID: record.PagingToken(), Data: res})
			}
		},
	)
}

func (action *AccountTraitsIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.PagingParams = action.GetPageQuery()
}

func (action *AccountTraitsIndexAction) loadRecords() {
	action.Err = action.HistoryQ().Accounts().Blocked().Page(action.PagingParams).Select(&action.Records)
}

// LoadPage populates action.Page
func (action *AccountTraitsIndexAction) loadPage() {
	for _, record := range action.Records {
		var res resource.AccountTraits
		res.Populate(action.Ctx, record)
		action.Page.Add(res)
	}
	action.Page.BaseURL = action.BaseURL()
	action.Page.BasePath = "/traits"
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}

// AccountTraitsAction renders a account traits found by its address.
type AccountTraitsAction struct {
	Action
	Address        string
	HistoryRecord  history.Account
	Resource       resource.AccountTraits
}

// JSON is a method for actions.JSON
func (action *AccountTraitsAction) JSON() {
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
func (action *AccountTraitsAction) SSE(stream sse.Stream) {
	action.Do(
		action.loadParams,
		action.loadRecord,
		action.loadResource,
		func() {
			stream.SetLimit(10)
			stream.Send(sse.Event{Data: action.Resource})
		},
	)
}

func (action *AccountTraitsAction) loadParams() {
	action.Address = action.GetString("account_id")
}

func (action *AccountTraitsAction) loadRecord() {
	if action.Err != nil {
		return
	}

	action.Err = action.HistoryQ().AccountByAddress(&action.HistoryRecord, action.Address)
}

func (action *AccountTraitsAction) loadResource() {
	action.Err = action.Resource.Populate(
		action.Ctx,
		action.HistoryRecord,
	)
}
