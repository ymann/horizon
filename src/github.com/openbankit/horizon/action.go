package horizon

import (
	"net/http"
	"net/url"

	"github.com/openbankit/horizon/actions"
	"github.com/openbankit/horizon/admin"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/httpx"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/toid"
	"github.com/zenazn/goji/web"
)

// Action is the "base type" for all actions in horizon.  It provides
// structs that embed it with access to the App struct.
//
// Additionally, this type is a trigger for go-codegen and causes
// the file at Action.tmpl to be instantiated for each struct that
// embeds Action.
type Action struct {
	actions.Base
	App *App
	Log *log.Entry

	hq              *history.Q
	cq              *core.Q
	signersProvider core.SignersProvider
	adminAction     admin.AdminActionInterface
}

func (action *Action) SetSignersProvider(signersProvider core.SignersProvider) {
	action.signersProvider = signersProvider
}

func (action *Action) SignersProvider() core.SignersProvider {
	if action.signersProvider == nil {
		if action.App.SignersProvider() != nil {
			action.signersProvider = action.App.SignersProvider()
		} else {
			action.signersProvider = action.CoreQ()
		}
	}
	return action.signersProvider
}

// CoreQ provides access to queries that access the stellar core database.
func (action *Action) CoreQ() *core.Q {
	if action.cq == nil {
		action.cq = &core.Q{Repo: action.App.CoreRepo(action.Ctx)}
	}

	return action.cq
}

// GetPagingParams modifies the base GetPagingParams method to replace
// cursors that are "now" with the last seen ledger's cursor.
func (action *Action) GetPagingParams() (cursor string, order string, limit uint64) {
	if action.Err != nil {
		return
	}

	cursor, order, limit = action.Base.GetPagingParams()

	if cursor == "now" {
		tid := toid.ID{
			LedgerSequence:   action.App.latestLedgerState.Horizon,
			TransactionOrder: toid.TransactionMask,
			OperationOrder:   toid.OperationMask,
		}
		cursor = tid.String()
	}

	return
}

// GetPageQuery is a helper that returns a new db.PageQuery struct initialized
// using the results from a call to GetPagingParams()
func (action *Action) GetPageQuery() db2.PageQuery {
	if action.Err != nil {
		return db2.PageQuery{}
	}

	r, err := db2.NewPageQuery(action.GetPagingParams())

	if err != nil {
		action.Err = err
	}

	return r
}

// GetCloseAtQuery is a helper that returns a new db.CloseAtQuery struct initialized
// using the results from a call to GetCloseAtParams()
func (action *Action) GetCloseAtQuery() db2.CloseAtQuery {
	if action.Err != nil {
		return db2.CloseAtQuery{}
	}

	after, before := action.GetCloseAtParams()
	if action.Err != nil {
		return db2.CloseAtQuery{}
	}
	var result db2.CloseAtQuery
	result, action.Err = db2.NewCloseAtQuery(after, before)
	return result
}

// HistoryQ provides access to queries that access the history portion of
// horizon's database.
func (action *Action) HistoryQ() *history.Q {
	if action.hq == nil {
		action.hq = &history.Q{Repo: action.App.HorizonRepo(action.Ctx)}
	}

	return action.hq
}

// Prepare sets the action's App field based upon the goji context
func (action *Action) Prepare(c web.C, w http.ResponseWriter, r *http.Request) {
	base := &action.Base
	base.Prepare(c, w, r)
	action.App = action.GojiCtx.Env["app"].(*App)

	if action.Ctx != nil {
		action.Log = log.Ctx(action.Ctx)
	} else {
		action.Log = log.DefaultLogger
	}
}

// ValidateCursorAsDefault ensures that the cursor parameter is valid in the way
// it is normally used, i.e. it is either the string "now" or a string of
// numerals that can be parsed as an int64.
func (action *Action) ValidateCursorAsDefault() {
	if action.Err != nil {
		return
	}

	if action.GetString(actions.ParamCursor) == "now" {
		return
	}

	action.GetInt64(actions.ParamCursor)
}

// BaseURL returns the base url for this requestion, defined as a url containing
// the Host and Scheme portions of the request uri.
func (action *Action) BaseURL() *url.URL {
	return httpx.BaseURL(action.Ctx)
}
