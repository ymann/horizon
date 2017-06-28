package cache

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/helpers"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHistoryAsset(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	db := history.Q{tt.HorizonRepo()}
	c := NewHistoryAsset(&db)
	tt.Assert.Equal(0, c.Cache.ItemCount())
	config := test.NewTestConfig()

	var xdrAsset xdr.Asset
	issuer, err := helpers.ParseAccountId(config.BankMasterKey)
	assert.Nil(t, err)
	xdrAsset.SetCredit("UAH", issuer)
	Convey("Chache not found", t, func() {
		var nonAsset xdr.Asset
		nonAsset.SetCredit("AAA", issuer)
		So(c.Cache.ItemCount(), ShouldEqual, 0)
		stored, err := c.Get(nonAsset)
		assert.Nil(t, err)
		assert.Nil(t, stored)
		So(c.Cache.ItemCount(), ShouldEqual, 1)
		Convey("Cache is shared", func() {
			c := NewHistoryAsset(&db)
			So(c.Cache.ItemCount(), ShouldEqual, 1)
		})
	})
	Convey("Get assets:", t, func() {
		expected := history.Asset{
			Id:          1,
			Type:        int(xdr.AssetTypeAssetTypeCreditAlphanum4),
			Code:        "UAH",
			Issuer:      config.BankMasterKey,
			IsAnonymous: false,
		}
		storedAsset, err := c.Get(xdrAsset)
		So(err, ShouldBeNil)
		assert.Equal(t, expected, *storedAsset)
		updateExpected := expected
		updateExpected.IsAnonymous = true
		isUpdated, err := db.UpdateAsset(&updateExpected)
		So(err, ShouldBeNil)
		So(isUpdated, ShouldBeTrue)
		// db is updated but cache is not
		storedAsset, err = c.Get(xdrAsset)
		So(err, ShouldBeNil)
		assert.Equal(t, expected, *storedAsset)
	})

}
