package redis

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/log"
	"github.com/stretchr/testify/mock"
	"time"
)

type ConnectionProviderMock struct {
	mock.Mock
}

func (p *ConnectionProviderMock) GetConnection() ConnectionInterface {
	a := p.Called()
	return a.Get(0).(ConnectionInterface)
}

type ConnectionMock struct {
	mock.Mock
}

func (m *ConnectionMock) HMSet(args ...interface{}) error {
	log.Panic("Not implemented")
	return nil
}

func (m *ConnectionMock) HGetAll(key string) (interface{}, error) {
	log.Panic("Not implemented")
	return nil, nil
}

func (m *ConnectionMock) Expire(key string, timeout time.Duration) (bool, error) {
	log.Panic("Not implemented")
	return false, nil
}

func (m *ConnectionMock) GetSet(key string, data interface{}) (interface{}, error) {
	log.Panic("Not implemented")
	return nil, nil
}

func (m *ConnectionMock) Get(key string) (interface{}, error) {
	log.Panic("Not implemented")
	return nil, nil
}

func (m *ConnectionMock) Set(key string, data interface{}) error {
	log.Panic("Not implemented")
	return nil
}

func (m *ConnectionMock) Watch(key string) error {
	a := m.Called(key)
	return a.Error(0)
}

func (m *ConnectionMock) UnWatch() error {
	a := m.Called()
	return a.Error(0)
}

func (m *ConnectionMock) Multi() error {
	a := m.Called()
	return a.Error(0)
}

func (m *ConnectionMock) Exec() (bool, error) {
	a := m.Called()
	return a.Get(0).(bool), a.Error(1)
}

func (m *ConnectionMock) Close() error {
	a := m.Called()
	return a.Error(0)
}

func (m *ConnectionMock) Delete(key string) error {
	return m.Called(key).Error(0)
}

func (m *ConnectionMock) Ping() error {
	return nil
}

type ProcessedOpProviderMock struct {
	mock.Mock
}

func (p *ProcessedOpProviderMock) Insert(processedOp *ProcessedOp, timeout time.Duration) error {
	a := p.Called(processedOp, timeout)
	return a.Error(0)
}
func (p *ProcessedOpProviderMock) Get(txHash string, opIndex int, isIncome bool) (*ProcessedOp, error) {
	a := p.Called(txHash, opIndex, isIncome)
	rawProcessedOp := a.Get(0)
	if rawProcessedOp == nil {
		return nil, a.Error(1)
	}

	return rawProcessedOp.(*ProcessedOp), a.Error(1)
}

func (p *ProcessedOpProviderMock) Delete(txHash string, opIndex int, isIncome bool) error {
	a := p.Called(txHash, opIndex, isIncome)
	return a.Error(0)
}

type AccountStatisticsProviderMock struct {
	mock.Mock
}

func (p AccountStatisticsProviderMock) Insert(stats *AccountStatistics, timeout time.Duration) error {
	return p.Called(stats, timeout).Error(0)
}
func (p AccountStatisticsProviderMock) Get(accountId, assetCode string, counterparties []xdr.AccountType) (*AccountStatistics, error) {
	a := p.Called(accountId, assetCode, counterparties)
	rawAccStats := a.Get(0)
	if rawAccStats == nil {
		return nil, a.Error(1)
	}
	return rawAccStats.(*AccountStatistics), a.Error(1)
}
