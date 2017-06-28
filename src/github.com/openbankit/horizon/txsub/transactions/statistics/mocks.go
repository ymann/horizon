package statistics

import (
	"github.com/stretchr/testify/mock"
	"time"
	"github.com/openbankit/horizon/redis"
)

type ManagerMock struct {
	mock.Mock
}

func (m *ManagerMock) UpdateGet(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) (*redis.AccountStatistics, error) {
	a := m.Called(paymentData, paymentDirection, now)
	return a.Get(0).(*redis.AccountStatistics), a.Error(1)
}

func (m *ManagerMock) CancelOp(paymentData *PaymentData, paymentDirection PaymentDirection, now time.Time) error {
	return m.Called(paymentData, paymentDirection, now).Error(0)
}

