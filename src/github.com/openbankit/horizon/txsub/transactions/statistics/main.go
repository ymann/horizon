package statistics

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/config"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/redis"
	"errors"
	"time"
)

type ManagerInterface interface {
	// Gets statistics for account-asset pair, updates it with opAmount and returns stats
	UpdateGet(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) (result *redis.AccountStatistics, err error)

	// Cancels Op - removes it from processed ops and subtracts from stats
	CancelOp(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) error
}

type Manager struct {
	counterparties     []xdr.AccountType
	statisticsTimeOut  time.Duration
	processedOpTimeOut time.Duration
	numOfRetires       int
	log                *log.Entry

	historyQ                    history.QInterface
	connectionProvider          redis.ConnectionProviderInterface
	defaultProcessedOpProvider  redis.ProcessedOpProviderInterface
	defaultAccountStatsProvider redis.AccountStatisticsProviderInterface
}

// Creates new statistics manager. counterparties MUST BE FULL ARRAY OF COUTERPARTIES.
func NewManager(historyQ history.QInterface, counterparties []xdr.AccountType, config *config.Config) *Manager {
	return &Manager{
		historyQ:           historyQ,
		counterparties:     counterparties,
		statisticsTimeOut:  config.StatisticsTimeout,
		processedOpTimeOut: config.ProcessedOpTimeout,
		numOfRetires:       5,
		log:                log.WithField("service", "statistics_manager"),
	}
}

// statisticsTimeOut must be bigger then ledger's close time (~2 minutes is recommended)
// timeout for statistics must be greater then for processed op
func (m *Manager) SetStatisticsTimeout(timeout time.Duration) *Manager {
	m.statisticsTimeOut = timeout
	return m
}

// timeout for statistics must be greater then for processed op
func (m *Manager) SetProcessedOpTimeout(timeout time.Duration) *Manager {
	m.processedOpTimeOut = timeout
	return m
}

func (m *Manager) getConnectionProvider() redis.ConnectionProviderInterface {
	if m.connectionProvider == nil {
		m.connectionProvider = redis.NewConnectionProvider()
	}
	return m.connectionProvider
}

func (m *Manager) getProcessedOpProvider(conn redis.ConnectionInterface) redis.ProcessedOpProviderInterface {
	if m.defaultProcessedOpProvider != nil {
		return m.defaultProcessedOpProvider
	}
	return redis.NewProcessedOpProvider(conn)
}

func (m *Manager) getAccountStatsProvider(conn redis.ConnectionInterface) redis.AccountStatisticsProviderInterface {
	if m.defaultAccountStatsProvider != nil {
		return m.defaultAccountStatsProvider
	}
	return redis.NewAccountStatisticsProvider(conn)
}

func (m *Manager) CancelOp(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) error {
	for i := 0; i < m.numOfRetires; i++ {
		m.log.WithField("retry", i).Debug("CancelOp started new retry")
		var needRetry bool
		needRetry, err := m.cancelOp(paymentData, paymentDirection, now)
		if err != nil {
			if !redis.IsConnectionClosed(err) {
				return err
			}
			needRetry = true
		}

		if !needRetry {
			return nil
		}
	}

	return errors.New("Failed to cancel op")
}

// Returns true if retry needed
func (m *Manager) cancelOp(paymentData *PaymentData, direction PaymentDirection, now time.Time) (bool, error) {
	m.log.Debug("Getting new connection")
	conn := m.getConnectionProvider().GetConnection()
	defer conn.Close()

	// Check if op is still in redis
	processedOp, err := m.getProcessedOp(conn, paymentData, direction)
	if err != nil {
		return false, err
	}

	if processedOp == nil {
		// op is already canceled - remove op watch
		m.log.Debug("Op is canceled - unwatching")
		err = conn.UnWatch()
		return false, err
	}

	// Get stats
	account := paymentData.GetAccount(direction)
	accountStats, err := m.getAccountStatistics(account.Address, paymentData.Asset.Code, conn)
	if err != nil {
		m.log.WithError(err).Error("Failed to get account statistics")
		return false, err
	}

	if accountStats == nil {
		// no need to cancel
		m.log.Debug("Stats are not in redis - no need to cancel operation")
		err = conn.UnWatch()
		return false, err
	}

	if direction.IsIncoming() {
		accountStats.Balance -= paymentData.Amount
	}

	counterparty := paymentData.GetCounterparty(direction)
	for key, value := range accountStats.AccountsStatistics {
		value.ClearObsoleteStats(now)
		if key == counterparty.AccountType {
			value.Update(-paymentData.Amount, processedOp.TimeUpdated, now, direction.IsIncoming())
		}
		accountStats.AccountsStatistics[key] = value
	}

	// Update stats and del op processed
	// 4 Start multi
	m.log.Debug("Starting multi")
	err = conn.Multi()
	if err != nil {
		return false, err
	}

	// 5. Save to redis stats
	err = m.getAccountStatsProvider(conn).Insert(accountStats, m.statisticsTimeOut)
	if err != nil {
		return false, err
	}

	processedOp = redis.NewProcessedOp(paymentData.TxHash, paymentData.Index, paymentData.Amount, direction.IsIncoming(), now)
	// 6. Mark Op processed
	err = m.getProcessedOpProvider(conn).Delete(paymentData.TxHash, paymentData.Index, direction.IsIncoming())
	if err != nil {
		return false, err
	}

	// commit
	isOk, err := conn.Exec()
	if err != nil {
		return false, err
	}

	return !isOk, nil

}

func (m *Manager) UpdateGet(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) (result *redis.AccountStatistics, err error) {
	var accountStats *redis.AccountStatistics
	for i := 0; i < m.numOfRetires; i++ {
		m.log.WithField("retry", i).Debug("UpdateGet started new retry")
		var needRetry bool
		accountStats, needRetry, err = m.updateGet(paymentData, paymentDirection, now)
		if err != nil {
			m.log.WithError(err).Error("Failed to updateGet statistics")
			if !redis.IsConnectionClosed(err) {
				return nil, err
			}
			needRetry = true
		}

		if !needRetry {
			return accountStats, nil
		}
	}

	return nil, errors.New("Failed to Update and Get Account stats")
}

func (m *Manager) updateGet(paymentData *PaymentData, direction PaymentDirection, now time.Time) (*redis.AccountStatistics, bool, error) {
	m.log.Debug("Getting new connection")
	conn := m.getConnectionProvider().GetConnection()
	defer conn.Close()

	// 1. Check if op processed
	processedOp, err := m.getProcessedOp(conn, paymentData, direction)
	if err != nil {
		return nil, false, err
	}

	if processedOp != nil {
		// remove op watch
		m.log.Debug("Op is processed - unwatching")
		err := conn.UnWatch()
		if err != nil {
			return nil, false, err
		}

		return m.manageProcessedOp(conn, paymentData.GetAccount(direction).Address, paymentData.Asset, paymentData.GetAccountTrustLine(direction), now)
	}

	accountStats, err := m.getAccountStatistics(paymentData.GetAccount(direction).Address, paymentData.Asset.Code, conn)
	if err != nil {
		return nil, false, err
	}

	if accountStats == nil {
		// try get from db
		m.log.Debug("Getting stats from histroy")
		accountStats, err = m.tryGetStatisticsFromDB(paymentData.GetAccount(direction).Address, paymentData.Asset, paymentData.GetAccountTrustLine(direction), now)
		if err != nil {
			m.log.WithError(err).Error("Failed to get stats from history")
			return nil, false, err
		}
	} else {
		trustLine := paymentData.GetAccountTrustLine(direction)
		if accountStats.Balance == 0 && trustLine != nil {
			accountStats.Balance = int64(trustLine.Balance)
		}
	}
	m.log.WithField("account_stats", accountStats).Debug("Got account stats")

	// 4. Update stats and set op processed
	counterparty := paymentData.GetCounterparty(direction)
	m.updateStats(accountStats, counterparty.AccountType, direction.IsIncoming(), paymentData.Amount, now)
	// 4.1 Start multi
	m.log.Debug("Starting multi")
	err = conn.Multi()
	if err != nil {
		return nil, false, err
	}

	// 5. Save to redis stats
	err = m.getAccountStatsProvider(conn).Insert(accountStats, m.statisticsTimeOut)
	if err != nil {
		return nil, false, err
	}

	processedOp = redis.NewProcessedOp(paymentData.TxHash, paymentData.Index, paymentData.Amount, direction.IsIncoming(), now)
	// 6. Mark Op processed
	err = m.getProcessedOpProvider(conn).Insert(processedOp, m.processedOpTimeOut)
	if err != nil {
		return nil, false, err
	}

	// commit
	m.log.Debug("Exec multi")
	isOk, err := conn.Exec()
	if err != nil {
		return nil, false, err
	}

	if !isOk {
		return nil, true, nil
	}
	return accountStats, false, nil
}

func (m *Manager) getProcessedOp(conn redis.ConnectionInterface, paymentData *PaymentData, paymentDirection PaymentDirection) (*redis.ProcessedOp, error) {
	// 1. Watch op
	m.log.Debug("Setting watch for processed op key")
	opKey := redis.GetProcessedOpKey(paymentData.TxHash, paymentData.Index, paymentDirection.IsIncoming())
	err := conn.Watch(opKey)
	if err != nil {
		return nil, err
	}

	// 2. Get op
	m.log.Debug("Checking if op was processed")
	processedOpProvider := m.getProcessedOpProvider(conn)
	processedOp, err := processedOpProvider.Get(paymentData.TxHash, paymentData.Index, paymentDirection.IsIncoming())
	if err != nil {
		m.log.WithError(err).Error("Failed to get processed op")
		return nil, err
	}

	return processedOp, nil
}

func (m *Manager) getAccountStatistics(account, assetCode string, conn redis.ConnectionInterface) (*redis.AccountStatistics, error) {
	// 1. Watch stats
	m.log.Debug("Watching account stats")
	statsKey := redis.GetAccountStatisticsKey(account, assetCode)
	err := conn.Watch(statsKey)
	if err != nil {
		m.log.WithError(err).Error("Failed to watch stats key")
		return nil, err
	}
	// 2. Get stats
	m.log.Debug("Getting account stats from redis")
	accountStatsProvider := m.getAccountStatsProvider(conn)
	accountStats, err := accountStatsProvider.Get(account, assetCode, m.counterparties)
	if err != nil {
		m.log.WithError(err).Error("Failed to get stats from redis")
		return nil, err
	}

	return accountStats, nil
}

func (m *Manager) manageProcessedOp(conn redis.ConnectionInterface, account string, asset history.Asset, trustLine *core.Trustline, now time.Time) (*redis.AccountStatistics, bool, error) {
	// try get stats from redis
	accountStatsProvider := m.getAccountStatsProvider(conn)
	accountStats, err := accountStatsProvider.Get(account, asset.Code, m.counterparties)
	if err != nil {
		return nil, false, err
	}

	if accountStats == nil {
		// try get stats from history
		accountStats, err = m.tryGetStatisticsFromDB(account, asset, trustLine, now)
		if err != nil {
			return nil, false, err
		}

		err := accountStatsProvider.Insert(accountStats, m.statisticsTimeOut)
		if err != nil {
			return nil, false, err
		}
		return accountStats, false, err
	} else {
		if accountStats.Balance == 0 && trustLine != nil {
			accountStats.Balance = int64(trustLine.Balance)
		}
	}

	return accountStats, false, nil
}

func (m *Manager) updateStats(accountStats *redis.AccountStatistics, counterparty xdr.AccountType, isIncome bool, opAmount int64, now time.Time) {
	if isIncome {
		accountStats.Balance += opAmount
	}
	_, ok := accountStats.AccountsStatistics[counterparty]
	if !ok {
		accountStats.AccountsStatistics[counterparty] = history.NewAccountStatistics(accountStats.Account, accountStats.AssetCode, counterparty)
	}
	for key, value := range accountStats.AccountsStatistics {
		value.ClearObsoleteStats(now)
		if key == counterparty {
			value.Update(opAmount, now, now, isIncome)
		}
		accountStats.AccountsStatistics[key] = value
	}
}

func (m *Manager) tryGetStatisticsFromDB(account string, asset history.Asset, trustLine *core.Trustline, now time.Time) (*redis.AccountStatistics, error) {
	balance := int64(0)
	if trustLine != nil {
		balance = int64(trustLine.Balance)
	}
	accountStats := redis.NewAccountStatistics(account, asset.Code, balance, make(map[xdr.AccountType]history.AccountStatistics))
	err := m.historyQ.GetStatisticsByAccountAndAsset(accountStats.AccountsStatistics, account, asset.Code, now)
	if err != nil {
		return nil, err
	}
	return accountStats, nil
}
