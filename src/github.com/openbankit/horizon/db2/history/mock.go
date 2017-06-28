package history

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/log"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"time"
)

type QMock struct {
	mock.Mock
}

// GetAccountLimits returns limits row by account and asset.
func (m *QMock) GetAccountLimits(dest interface{}, address string, assetCode string) error {
	a := m.Called(address, assetCode)
	rawLimits := a.Get(0)
	if rawLimits != nil {
		limits := rawLimits.(AccountLimits)
		destLimits := dest.(*AccountLimits)
		*destLimits = limits
	}
	return a.Error(1)
}

// Inserts new account limits instance
func (m *QMock) CreateAccountLimits(limits AccountLimits) error {
	return m.Called(limits).Error(0)
}

// Updates account's limits
func (m *QMock) UpdateAccountLimits(limits AccountLimits) error {
	return m.Called(limits).Error(0)
}

// GetStatisticsByAccountAndAsset selects rows from `account_statistics` by address and asset code
func (m *QMock) GetStatisticsByAccountAndAsset(dest map[xdr.AccountType]AccountStatistics, addy string, assetCode string, now time.Time) error {
	a := m.Called(addy, assetCode, now)
	rawStats := a.Get(0)
	if rawStats != nil {
		for key, value := range rawStats.(map[xdr.AccountType]AccountStatistics) {
			dest[key] = value
		}
	}
	return a.Error(1)
}

func (m *QMock) Asset(dest interface{}, asset xdr.Asset) error {
	log.Debug("Asset is called")
	a := m.Called(asset)
	rawAsset := a.Get(0)
	if rawAsset != nil {
		log.Debug("Raw asset is not null")
		asset := a.Get(0).(Asset)
		destAsset := dest.(*Asset)
		*destAsset = asset
	}
	return a.Error(1)
}

// Deletes asset from db by id
func (m *QMock) DeleteAsset(id int64) (bool, error) {
	log.Panic("Not implemented")
	return false, nil
}

// updates asset
func (m *QMock) UpdateAsset(asset *Asset) (bool, error) {
	log.Panic("Not implemented")
	return false, nil
}

// inserts asset
func (m *QMock) InsertAsset(asset *Asset) (err error) {
	log.Panic("Not implemented")
	return nil
}

func (m *QMock) AccountByAddress(dest interface{}, addy string) error {
	a := m.Called(addy)
	rawAccount := a.Get(0)
	if rawAccount != nil {
		account := a.Get(0).(Account)
		destAccount := dest.(*Account)
		*destAccount = account
	}
	return a.Error(1)
}

// selects commission by id
func (m *QMock) CommissionByHash(hash string) (*Commission, error) {
	log.Panic("Not implemented")
	return nil, nil
}

// Inserts new commission
func (m *QMock) InsertCommission(commission *Commission) (err error) {
	log.Panic("Not implemented")
	return nil
}

// Deletes commission
func (m *QMock) DeleteCommission(hash string) (bool, error) {
	log.Panic("Not implemented")
	return false, nil
}

// update commission
func (m *QMock) UpdateCommission(commission *Commission) (bool, error) {
	log.Panic("Not implemented")
	return false, nil
}

func (m *QMock) AccountUpdate(account *Account) error {
	return m.Called(account).Error(0)
}

func (m *QMock) AccountIDByAddress(addy string) (int64, error) {
	a := m.Called(addy)
	return a.Get(0).(int64), a.Error(1)
}

func (m *QMock) GetAccountStatistics(address string, assetCode string, counterPartyType xdr.AccountType) (AccountStatistics, error) {
	a := m.Called(address, assetCode, counterPartyType)
	return a.Get(0).(AccountStatistics), a.Error(1)
}

func (m *QMock) AssetByParams(dest interface{}, assetType int, code string, issuer string) error {
	a := m.Called(dest, assetType, code, issuer)
	return a.Error(0)
}

func (m *QMock) GetHighestWeightCommission(keys map[string]CommissionKey) (resultingCommissions []Commission, err error) {
	a := m.Called(keys)
	return a.Get(0).([]Commission), a.Error(1)
}

func (m *QMock) OperationByID(dest interface{}, id int64) error {
	a := m.Called(dest, id)
	return a.Error(0)
}

func (m *QMock) OptionsByName(name string) (*Options, error) {
	a := m.Called(name)
	options := a.Get(0)
	err := a.Error(1)
	if options == nil {
		return nil, err
	}
	return options.(*Options), err
}
func (m *QMock) OptionsInsert(options *Options) (error) {
	a := m.Called(options)
	return a.Error(0)
}
func (m *QMock) OptionsUpdate(options *Options) (bool, error) {
	a := m.Called(options)
	return a.Bool(0), a.Error(1)
}
func (m *QMock) OptionsDelete(name string) (bool, error) {
	a := m.Called(name)
	return a.Bool(0), a.Error(1)
}

func CreateRandomAccountStats(account string, counterpartyType xdr.AccountType, asset string) AccountStatistics {
	return CreateRandomAccountStatsWithMinValue(account, counterpartyType, asset, 0)
}

func CreateRandomAccountStatsWithMinValue(account string, counterpartyType xdr.AccountType, asset string, minValue int64) AccountStatistics {
	return AccountStatistics{
		Account:          account,
		AssetCode:        asset,
		CounterpartyType: int16(counterpartyType),
		DailyIncome:      Max(rand.Int63(), minValue),
		DailyOutcome:     Max(rand.Int63(), minValue),
		WeeklyIncome:     Max(rand.Int63(), minValue),
		WeeklyOutcome:    Max(rand.Int63(), minValue),
		MonthlyIncome:    Max(rand.Int63(), minValue),
		MonthlyOutcome:   Max(rand.Int63(), minValue),
		AnnualIncome:     Max(rand.Int63(), minValue),
		AnnualOutcome:    Max(rand.Int63(), minValue),
		UpdatedAt:        time.Unix(time.Now().Unix(), 0),
	}
}

func Max(x int64, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
