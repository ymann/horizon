package horizon

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/accounttypes"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/redis"
	"github.com/openbankit/horizon/render/hal"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/render/sse"
	"github.com/openbankit/horizon/resource"
	"github.com/go-errors/errors"
	"time"
)

// AccountStatisticsAction detailed income/outcome statistics for single account
type AccountStatisticsAction struct {
	Action
	Address       string
	AssetCode     string
	IsCached      bool
	HistoryRecord history.Account
	Statistics    []history.AccountStatistics
	Resource      resource.AccountStatistics
}

// JSON is a method for actions.JSON
func (action *AccountStatisticsAction) JSON() {
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
func (action *AccountStatisticsAction) SSE(stream sse.Stream) {
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

func (action *AccountStatisticsAction) loadParams() {
	action.Address = action.GetString("account_id")
	action.AssetCode = action.GetString("asset_code")
	action.IsCached = action.GetBool("cached")
}

func (action *AccountStatisticsAction) loadRecord() {
	action.Err = action.HistoryQ().AccountByAddress(&action.HistoryRecord, action.Address)
	if action.Err != nil {
		return
	}

	if action.IsCached {
		action.loadFromCache()
		return
	}

	action.loadFromDB()

}

func (action *AccountStatisticsAction) loadFromDB() {
	if action.AssetCode == "" {
		action.Err = action.HistoryQ().GetStatisticsByAccount(&action.Statistics, action.Address)
		return
	}

	response := make(map[xdr.AccountType]history.AccountStatistics)
	action.Err = action.HistoryQ().GetStatisticsByAccountAndAsset(response, action.Address, action.AssetCode, time.Now())
	if action.Err == nil {
		action.mapToArray(response)
	}
}

func (action *AccountStatisticsAction) mapToArray(response map[xdr.AccountType]history.AccountStatistics) {
	action.Statistics = make([]history.AccountStatistics, len(response))
	i := 0
	for _, value := range response {
		action.Statistics[i] = value
		i++
	}
}

func (action *AccountStatisticsAction) loadFromCache() {
	if action.AssetCode == "" {
		action.SetInvalidField("asset_code", errors.New("Can not be empty"))
		return
	}

	conn := redis.NewConnectionProvider().GetConnection()
	defer conn.Close()
	stats, err := redis.NewAccountStatisticsProvider(conn).Get(action.Address, action.AssetCode, accounttype.GetAll())
	if err != nil {
		action.Err = &problem.ServerError
		return
	}

	if stats != nil {
		action.mapToArray(stats.AccountsStatistics)
	}
}

func (action *AccountStatisticsAction) loadResource() {
	action.Err = action.Resource.Populate(
		action.Ctx,
		action.Statistics,
		action.HistoryRecord,
	)
}
