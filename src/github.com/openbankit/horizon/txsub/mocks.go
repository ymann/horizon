package txsub

import (
	"github.com/openbankit/horizon/txsub/transactions"
	"github.com/stretchr/testify/mock"
)

type TransactionValidatorMock struct {
	mock.Mock
}

func (v *TransactionValidatorMock) CheckTransaction(envelopeInfo *transactions.EnvelopeInfo) error {
	a := v.Called(envelopeInfo)
	return a.Error(0)
}
