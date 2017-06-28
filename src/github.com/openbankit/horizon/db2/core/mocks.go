package core

import (
	"github.com/stretchr/testify/mock"
	"github.com/openbankit/go-base/xdr"
)

type SignersProviderMock struct {
	mock.Mock
}

func (m *SignersProviderMock) SignersByAddress(dest interface{}, addy string) error {
	a := m.Called(addy)
	signers := a.Get(0).([]Signer)
	destSigners := dest.(*[]Signer)
	*destSigners = signers
	return a.Error(1)
}

type QMock struct {
	mock.Mock
}

func (m *QMock) TrustlineByAddressAndAsset(dest interface{}, addy string, assetCode string, issuer string) error {
	a := m.Called(addy, assetCode, issuer)
	rawTrustLine := a.Get(0)
	if rawTrustLine == nil {
		return a.Error(1)
	}
	trustline := rawTrustLine.(Trustline)
	destTrustLine := dest.(*Trustline)
	*destTrustLine = trustline
	return a.Error(1)
}

func (m *QMock) AccountByAddress(dest interface{}, addy string) error {
	a := m.Called(addy)
	rawAccount := a.Get(0)
	if rawAccount == nil {
		return a.Error(1)
	}
	account := rawAccount.(Account)
	destAccount := dest.(*Account)
	*destAccount = account
	return a.Error(1)
}

func (m *QMock) AccountTypeByAddress(addy string) (xdr.AccountType, error) {
	a := m.Called(addy)
	return a.Get(0).(xdr.AccountType), a.Error(1)
}
