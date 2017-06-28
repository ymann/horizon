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
// AssetIndexAction: pages of assets in order of creation
// AssetIndexAction renders a page of asset resources, identified by
// a normal page query, ordered by the operation id that created them.
type AssetIndexAction struct {
	Action
	PagingParams db2.PageQuery
	Records      []history.Asset
	Page         hal.Page
}

// JSON is a method for actions.JSON
func (action *AssetIndexAction) JSON() {
	action.Do(
		action.loadParams,
		action.loadRecords,
		action.loadPage,
		func() { hal.Render(action.W, action.Page) },
	)
}

// SSE is a method for actions.SSE
func (action *AssetIndexAction) SSE(stream sse.Stream) {
	action.Setup(action.loadParams)
	action.Do(
		action.loadRecords,
		func() {
			stream.SetLimit(int(action.PagingParams.Limit))
			var res resource.HistoryAsset
			for _, record := range action.Records[stream.SentCount():] {
				res.Populate(action.Ctx, record)
				stream.Send(sse.Event{ID: res.PagingToken(), Data: res})
			}
		},
	)
}

func (action *AssetIndexAction) loadParams() {
	action.ValidateCursorAsDefault()
	action.PagingParams = action.GetPageQuery()
}

func (action *AssetIndexAction) loadRecords() {
	action.Err = action.HistoryQ().Assets().Page(action.PagingParams).Select(&action.Records)
}

// LoadPage populates action.Page
func (action *AssetIndexAction) loadPage() {
	for _, record := range action.Records {
		var res resource.HistoryAsset
		res.Populate(action.Ctx, record)
		action.Page.Add(res)
	}
	action.Page.BaseURL = action.BaseURL()
	action.Page.BasePath = "/assets"
	action.Page.Limit = action.PagingParams.Limit
	action.Page.Cursor = action.PagingParams.Cursor
	action.Page.Order = action.PagingParams.Order
	action.Page.PopulateLinks()
}
