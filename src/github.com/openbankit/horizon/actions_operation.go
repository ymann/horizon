package horizon

import (
	"encoding/json"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/render/sse"
	"github.com/openbankit/horizon/resource"
)

// This file contains the actions:
//
// OperationIndexAction: pages of operations
// OperationShowAction: single operation by id

// OperationIndexAction renders a page of operations resources, identified by
// a normal page query and optionally filtered by an account, ledger, or
// transaction.
// Allows to select operations by before & after filters. Time must be passed as 2006-01-02T15:04:05Z
type OperationIndexAction struct {
	Action
	LedgerFilter      int32
	AccountFilter     string
	MultiAccountFilter	[]string
	TransactionFilter string
	PagingParams      db2.PageQuery
	CloseAtQuery      db2.CloseAtQuery
	Records           []history.Operation
	Page              hal.Page
}

// JSON is a method for actions.JSON
func (action *OperationIndexAction) JSON() {
	action.Do(action.loadParams, action.loadRecords, action.loadPage)
	action.Do(func() {
		hal.Render(action.W, action.Page)
	})
}

// SSE is a method for actions.SSE
func (action *OperationIndexAction) SSE(stream sse.Stream) {
	action.Setup(action.loadParams)
	action.Do(
		action.loadRecords,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			records := action.Records[stream.SentCount():]

			for _, record := range records {
				res, err := resource.NewOperation(action.Ctx, record)

				if err != nil {
					stream.Err(action.Err)
					return
				}

				stream.Send(sse.Event{
					ID:   res.PagingToken(),
					Data: res,
				})
			}
		})

}

func (action *OperationIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.AccountFilter = action.GetString("account_id")
	addressesStr := action.GetString("multi_accounts")
	if action.AccountFilter == "" && addressesStr != "" {
		var addresses []string 
		err := json.Unmarshal([]byte(addressesStr),&addresses)
		if err == nil{
			action.MultiAccountFilter = addresses
		}
	}
	action.LedgerFilter = action.GetInt32("ledger_id")
	action.TransactionFilter = action.GetString("tx_id")
	action.PagingParams = action.GetPageQuery()
	action.CloseAtQuery = action.GetCloseAtQuery()
}

func (action *OperationIndexAction) loadRecords() {
	if action.Err != nil {
		return
	}

	q := action.HistoryQ()
	ops := q.Operations()

	switch {
	case action.AccountFilter != "":
		ops.ForAccount(action.AccountFilter)
	case action.MultiAccountFilter != nil:
		ops.ForAccounts(action.MultiAccountFilter)
	case action.LedgerFilter > 0:
		ops.ForLedger(action.LedgerFilter)
	case action.TransactionFilter != "":
		ops.ForTransaction(action.TransactionFilter)
	}

	action.Err = ops.Page(action.PagingParams).ClosedAt(action.CloseAtQuery).Select(&action.Records)
}

func (action *OperationIndexAction) loadPage() {
	if action.Err != nil {
		return
	}

	for _, record := range action.Records {
		var res hal.Pageable
		res, action.Err = resource.NewOperation(action.Ctx, record)
		if action.Err != nil {
			return
		}
		action.Page.Add(res)
	}

	action.Page.BaseURL = action.BaseURL()
	action.Page.BasePath = action.Path()
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}

// OperationShowAction renders a ledger found by its sequence number.
type OperationShowAction struct {
	Action
	ID       int64
	Record   history.Operation
	Resource interface{}
}

func (action *OperationShowAction) loadParams() {
	action.ID = action.GetInt64("id")
}

func (action *OperationShowAction) loadRecord() {
	action.Err = action.HistoryQ().OperationByID(&action.Record, action.ID)
}

func (action *OperationShowAction) loadResource() {
	action.Resource, action.Err = resource.NewOperation(action.Ctx, action.Record)
}

// JSON is a method for actions.JSON
func (action *OperationShowAction) JSON() {
	action.Do(action.loadParams, action.loadRecord, action.loadResource)
	action.Do(func() {
		hal.Render(action.W, action.Resource)
	})
}
