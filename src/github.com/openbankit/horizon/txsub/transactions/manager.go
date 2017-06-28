package transactions

import (
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/transactions/statistics"
)

type Manager struct {
	*cache.SharedCache
	CoreQ        core.QInterface
	HistoryQ     history.QInterface
	StatsManager statistics.ManagerInterface
	Config       *config.Config
}

func NewManager(core core.QInterface, history history.QInterface, statsManager statistics.ManagerInterface, config *config.Config, sharedCache *cache.SharedCache) *Manager {
	return &Manager{
		CoreQ:        core,
		HistoryQ:     history,
		StatsManager: statsManager,
		Config:       config,
		SharedCache:  sharedCache,
	}
}
