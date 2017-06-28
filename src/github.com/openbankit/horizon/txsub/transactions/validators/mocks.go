package validators

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/stretchr/testify/mock"
)

type AccountTypeValidatorMock struct {
	mock.Mock
}

func (v *AccountTypeValidatorMock) VerifyAccountTypesForPayment(from, to xdr.AccountType) *results.RestrictedForAccountTypeError {
	a := v.Called(from, to)
	result := a.Get(0)
	if result == nil {
		return nil
	}
	return result.(*results.RestrictedForAccountTypeError)
}

type AssetsValidatorMock struct {
	mock.Mock
}

func (v *AssetsValidatorMock) GetValidAsset(asset xdr.Asset) (*history.Asset, error) {
	a := v.Called(asset)
	rawAsset := a.Get(0)
	if rawAsset != nil {
		asset := rawAsset.(*history.Asset)
		return asset, a.Error(1)
	}
	return nil, a.Error(1)
}
func (v *AssetsValidatorMock) IsAssetValid(asset xdr.Asset) (bool, error) {
	a := v.Called(asset)
	return a.Get(0).(bool), a.Error(1)
}
func (v *AssetsValidatorMock) IsAssetsValid(assets ...xdr.Asset) (bool, error) {
	a := v.Called(assets)
	return a.Get(0).(bool), a.Error(1)
}

type TraitsValidatorMock struct {
	mock.Mock
}

func (v *TraitsValidatorMock) CheckTraits(source, destination *history.Account) (*results.RestrictedForAccountError, error) {
	a := v.Called(source, destination)
	rawResult := a.Get(0)
	if rawResult != nil {
		result := rawResult.(*results.RestrictedForAccountError)
		return result, a.Error(1)
	}
	return nil, a.Error(1)
}
func (v *TraitsValidatorMock) CheckTraitsForAccount(account *history.Account, isSource bool) (*results.RestrictedForAccountError, error) {
	a := v.Called(account, isSource)
	rawResult := a.Get(0)
	if rawResult != nil {
		result := rawResult.(*results.RestrictedForAccountError)
		return result, a.Error(1)
	}
	return nil, a.Error(1)
}

type IncomingLimitsValidatorMock struct {
	mock.Mock
}

func (v *IncomingLimitsValidatorMock) VerifyLimits() (*results.ExceededLimitError, error) {
	a := v.Called()
	rawResult := a.Get(0)
	if rawResult != nil {
		result := rawResult.(*results.ExceededLimitError)
		return result, a.Error(1)
	}
	return nil, a.Error(1)
}

type OutgoingLimitsValidatorMock struct {
	mock.Mock
}

func (v *OutgoingLimitsValidatorMock) VerifyLimits() (*results.ExceededLimitError, error) {
	a := v.Called()
	rawResult := a.Get(0)
	if rawResult != nil {
		result := rawResult.(*results.ExceededLimitError)
		return result, a.Error(1)
	}
	return nil, a.Error(1)
}
