package history

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestCommissionHash(t *testing.T) {
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	issuer, err := keypair.Random()
	assert.Nil(t, err)
	EUR := details.Asset{
		Type:   assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4),
		Code:   "EUR",
		Issuer: issuer.Address(),
	}
	USD := EUR
	USD.Code = "USD"
	assert.NotEqual(t, EUR, USD)
	defaultFee := CommissionKey{}
	others := []CommissionKey{defaultFee}
	// only asset
	comKeyEUR := CommissionKey{
		Asset: EUR,
	}
	checkGreater(t, comKeyEUR, others)
	{
		comKeyUSD := CommissionKey{
			Asset: USD,
		}
		assert.Equal(t, 0, comKeyEUR.Compare(&comKeyUSD))
	}
	others = append(others, comKeyEUR)
	// only account type
	accType := int32(3)
	fromTypeKey := CommissionKey{
		FromType: &accType,
	}
	checkGreater(t, fromTypeKey, others)
	{
		accType = accType + 1
		toTypeKey := CommissionKey{
			ToType: &accType,
		}
		assert.Equal(t, 0, fromTypeKey.Compare(&toTypeKey))
	}
	others = append(others, fromTypeKey)
	// accountType and asset
	fromTypeAssetKey := fromTypeKey
	fromTypeAssetKey.Asset = EUR
	checkGreater(t, fromTypeAssetKey, others)
	others = append(others, fromTypeAssetKey)
	// only account
	accountId := issuer.Address()
	fromAcc := CommissionKey{
		From: accountId,
	}
	checkGreater(t, fromAcc, others)
	{
		toAcc := CommissionKey{
			To: accountId,
		}
		assert.Equal(t, 0, fromAcc.Compare(&toAcc))
	}
	others = append(others, fromAcc)
}

func checkGreater(t *testing.T, key CommissionKey, others []CommissionKey) {
	for _, other := range others {
		assert.Equal(t, 1, key.Compare(&other))
	}
}

func TestCommissionStore(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	q := &Q{tt.HorizonRepo()}
	err := q.DeleteCommissions()
	assert.Nil(t, err)
	Convey("not exist", t, func() {
		keys := CreateCommissionKeys("from", "to", 1, 3, details.Asset{})
		commissions, err := q.CommissionByKey(keys)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(commissions))
	})
	Convey("get", t, func() {
		var key CommissionKey
		{
			account, err := keypair.Random()
			assert.Nil(t, err)
			accountId := account.Address()
			var accountType int32
			key = CommissionKey{
				From:   accountId,
				ToType: &accountType,
			}
		}
		commission, err := NewCommission(key, 10*amount.One, 12*amount.One)
		assert.Nil(t, err)
		err = q.InsertCommission(commission)
		assert.Nil(t, err)

		keys := CreateCommissionKeys(key.From, "to", 123, *key.ToType, details.Asset{
			Type:   "random_type",
			Issuer: "random_issuer",
			Code:   "ASD",
		})
		stored, err := q.CommissionByKey(keys)
		assert.Nil(t, err)
		log.WithField("stored", stored).Debug("Got commission")
		assert.Equal(t, 1, len(stored))
		stored[0].weight = commission.weight
		assert.True(t, commission.Equals(stored[0]))
		err = q.DeleteCommissions()
		assert.Nil(t, err)
	})
	Convey("create keys", t, func() {
		keys := CreateCommissionKeys("from", "to", 1, 2, details.Asset{Type: "asset_type", Issuer: "Issuer", Code: "Code"})
		assert.Equal(t, 32, len(keys))
		for _, value := range keys {
			log.WithField("value", value).WithField("weight", value.CountWeight()).Info("got key")
		}
	})
	Convey("filter", t, func() {
		rawCommissions := []Commission{}
		filtered := filterByWeight(rawCommissions)
		assert.Equal(t, 0, len(filtered))
		rawCommissions = []Commission{
			Commission{
				weight: 2,
			},
			Commission{
				weight: 3,
			},
			Commission{
				weight: 3,
			},
		}
		filtered = filterByWeight(rawCommissions)
		assert.Equal(t, 2, len(filtered))
	})
}

func TestCommissionSelector(t *testing.T) {
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	q := &Q{test.Start(t).HorizonRepo()}
	accountIdA := getRandomAccountId(t)
	accountType := int32(4)
	asset := details.Asset{
		Issuer: getRandomAccountId(t),
		Code:   "USD",
		Type:   "random_type",
	}
	commission, err := NewCommission(CommissionKey{
		From:   accountIdA,
		ToType: &accountType,
		Asset:  asset,
	}, 10*amount.One, 12*amount.One)
	assert.Nil(t, err)

	err = q.InsertCommission(commission)
	assert.Nil(t, err)
	newAccountType := accountType + 1
	otherCommission, err := NewCommission(CommissionKey{
		From:     getRandomAccountId(t),
		To:       getRandomAccountId(t),
		ToType:   &newAccountType,
		FromType: &newAccountType,
		Asset:    asset,
	}, 10*amount.One, 12*amount.One)
	assert.Nil(t, err)
	err = q.InsertCommission(otherCommission)
	assert.Nil(t, err)
	Convey("by account", t, func() {
		var comms []Commission
		err = q.Commissions().ForAccount(accountIdA).Select(&comms)
		assert.Nil(t, err)
		log.WithField("comms", comms).WithField("len", len(comms)).Debug("Go commission by account")
		assert.Equal(t, 1, len(comms))
		assert.True(t, commission.Equals(comms[0]))
	})
	Convey("by account type", t, func() {
		var comms []Commission
		err = q.Commissions().ForAccountType(accountType).Select(&comms)
		assert.Nil(t, err)
		log.WithField("comms", comms).Debug("Go commission by account type")
		assert.Equal(t, 1, len(comms))
		assert.True(t, commission.Equals(comms[0]))
	})
	Convey("by asset", t, func() {
		var comms []Commission
		err = q.Commissions().ForAsset(asset).Select(&comms)
		assert.Nil(t, err)
		log.WithField("comms", comms).Debug("Go commission by asset")
		assert.Equal(t, 2, len(comms))
		equal := false
		for _, comm := range comms {
			if commission.Equals(comm) {
				equal = true
				break
			}
		}
		assert.True(t, equal)
	})
	Convey("by account and type", t, func() {
		var comms []Commission
		err = q.Commissions().ForAccount(accountIdA).ForAccountType(123).Select(&comms)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(comms))
		err = q.Commissions().ForAccount(accountIdA).ForAccountType(accountType).Select(&comms)
		assert.Nil(t, err)
		log.WithField("comms", comms).Debug("Go commission by account and type")
		assert.Equal(t, 1, len(comms))
		assert.True(t, commission.Equals(comms[0]))
	})
	Convey("by account and type and asset", t, func() {
		var comms []Commission
		err = q.Commissions().ForAccount(accountIdA).ForAccountType(accountType).ForAsset(details.Asset{
			Type: "new_random_type",
		}).Select(&comms)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(comms))
		err = q.Commissions().ForAccount(accountIdA).ForAccountType(accountType).ForAsset(asset).Select(&comms)
		assert.Nil(t, err)
		log.WithField("comms", comms).Debug("Go commission by account and type and asset")
		assert.Equal(t, 1, len(comms))
		assert.True(t, commission.Equals(comms[0]))
	})
	//err = q.deleteCommissions()
	assert.Nil(t, err)
	keys := CreateCommissionKeys(getRandomAccountId(t), getRandomAccountId(t), int32(1), int32(2), details.Asset{
		Type:   assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4),
		Code:   "EUR",
		Issuer: getRandomAccountId(t),
	})
	for _, key := range keys {
		comm, err := NewCommission(key, rand.Int63(), rand.Int63())
		assert.Nil(t, err)
		q.InsertCommission(comm)
	}
}

func getRandomAccountId(t *testing.T) string {
	key, err := keypair.Random()
	assert.Nil(t, err)
	return key.Address()
}
