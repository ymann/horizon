package horizon

import (
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/history"
	"time"
)

func initCacheFromApp(app *App) {
	app.sharedCache = initCache(app.HistoryQ())
}

func initCache(history *history.Q) *cache.SharedCache {
	return &cache.SharedCache{
		AccountHistoryCache: cache.NewHistoryAccountWithExp(history, time.Duration(2)*time.Minute, time.Duration(10)*time.Second),
	}
}

func init() {
	appInit.Add("cache", initCacheFromApp, "log", "horizon-db")
}
