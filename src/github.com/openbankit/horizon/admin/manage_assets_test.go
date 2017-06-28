package admin

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/test"
	"database/sql"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestActionsManageAssets(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	historyQ := &history.Q{tt.HorizonRepo()}
	Convey("Manage assets", t, func() {
		account, err := keypair.Random()
		So(err, ShouldBeNil)
		asset := details.Asset{
			Type:   assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4),
			Code:   "USD",
			Issuer: account.Address(),
		}
		assetData := make(map[string]interface{})
		assetData["asset_type"] = asset.Type
		assetData["asset_code"] = asset.Code
		assetData["asset_issuer"] = asset.Issuer
		Convey("Invalid asset", func() {
			assetData["asset_type"] = "invalid_type"
			action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "asset_type")
		})
		Convey("Invalid delete", func() {
			assetData["delete"] = "not_bool"
			action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "delete")
		})
		Convey("Invalid isAnonymous", func() {
			assetData["is_anonymous"] = "not_bool"
			action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "is_anonymous")
		})
		Convey("delete nonexistsing asset", func() {
			assetData["delete"] = "true"
			action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
			action.Validate()
			So(action.Err, problem.ShouldBeProblem, problem.NotFound)
		})
		Convey("happy path", func() {
			action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
			action.Validate()
			So(action.Err, ShouldBeNil)
			action.Apply()
			So(action.Err, ShouldBeNil)
			var storedAsset history.Asset
			err := historyQ.AssetByParams(&storedAsset, int(assets.AssetTypeMap[assetData["asset_type"].(string)]),
				assetData["asset_code"].(string), assetData["asset_issuer"].(string))
			So(err, ShouldBeNil)
			So(storedAsset.Id, ShouldNotEqual, 0)
			So(storedAsset.IsAnonymous, ShouldEqual, false)
			Convey("update", func() {
				assetData["is_anonymous"] = "true"
				action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
				action.Validate()
				So(action.Err, ShouldBeNil)
				action.Apply()
				So(action.Err, ShouldBeNil)
				var storedAsset history.Asset
				err := historyQ.AssetByParams(&storedAsset, int(assets.AssetTypeMap[assetData["asset_type"].(string)]),
					assetData["asset_code"].(string), assetData["asset_issuer"].(string))
				So(err, ShouldBeNil)
				So(storedAsset.Id, ShouldNotEqual, 0)
				So(storedAsset.IsAnonymous, ShouldEqual, true)
			})
			Convey("delete", func() {
				assetData["delete"] = "true"
				action := NewManageAssetsAction(NewAdminAction(assetData, historyQ))
				action.Validate()
				So(action.Err, ShouldBeNil)
				action.Apply()
				So(action.Err, ShouldBeNil)
				var storedAsset history.Asset
				err := historyQ.AssetByParams(&storedAsset, int(assets.AssetTypeMap[assetData["asset_type"].(string)]),
					assetData["asset_code"].(string), assetData["asset_issuer"].(string))
				So(err, ShouldEqual, sql.ErrNoRows)
			})
		})
	})
}
