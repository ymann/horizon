package admin

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestActionsSetCommission(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel
	historyQ := &history.Q{tt.HorizonRepo()}
	historyQ.DeleteCommissions()

	Convey("Set commission Actions:", t, func() {
		Convey("Not exists", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"delete": "true",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, problem.ShouldBeProblem, problem.NotFound)

		})
		Convey("Invalid asset", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"asset_type":   assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4),
				"asset_code":   "EUR",
				"asset_issuer": "random_issuer",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "asset_issuer")

		})
		Convey("Invalid from", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"from": "random_str",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "from")
		})
		Convey("Invalid to", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"to": "random_str",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "to")
		})
		Convey("Invalid from accountType", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"from_type": "10",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "from_type")
		})
		Convey("Invalid to accountType", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"to_type": "10",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "to_type")
		})
		Convey("Invalid flat_fee", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"flat_fee": "-10",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "flat_fee")
		})
		Convey("Invalid percent_fee", func() {
			action := NewSetCommissionAction(NewAdminAction(map[string]interface{}{
				"percent_fee": "-10",
				"flat_fee":    "0",
			}, historyQ))
			action.Validate()
			So(action.Err, ShouldNotBeNil)
			So(action.Err, ShouldBeInvalidField, "percent_fee")
		})
		Convey("valid insert", func() {
			fromKey, err := keypair.Random()
			assert.Nil(t, err)
			from := fromKey.Address()
			toKey, err := keypair.Random()
			assert.Nil(t, err)
			to := toKey.Address()
			fromType := strconv.Itoa(int(xdr.AccountTypeAccountBank))
			toType := strconv.Itoa(int(xdr.AccountTypeAccountDistributionAgent))
			issuer, err := keypair.Random()
			assert.Nil(t, err)
			assetType := assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4)
			assetCode := "EUR"
			assetIssuer := issuer.Address()
			flatFee := int64(12000000)
			percentFee := int64(11)
			data := map[string]interface{}{
				"from":         from,
				"to":           to,
				"from_type":    fromType,
				"to_type":      toType,
				"asset_type":   assetType,
				"asset_code":   assetCode,
				"asset_issuer": assetIssuer,
				"flat_fee":     strconv.FormatInt(flatFee, 10),
				"percent_fee":  strconv.FormatInt(percentFee, 10),
			}
			action := NewSetCommissionAction(NewAdminAction(data, historyQ))
			check := func(action AdminAction) {
				So(action.Err, ShouldBeNil)
				var sts []history.Commission
				err = historyQ.Commissions().ForAccount(from).Select(&sts)
				assert.Nil(t, err)
				assert.Equal(t, 1, len(sts))
				st := sts[0]
				stKey := st.GetKey()
				assert.Equal(t, from, stKey.From)
				assert.Equal(t, to, stKey.To)
				assert.Equal(t, fromType, strconv.Itoa(int(*stKey.FromType)))
				assert.Equal(t, toType, strconv.Itoa(int(*stKey.ToType)))
				assert.Equal(t, assetType, stKey.Asset.Type)
				assert.Equal(t, assetCode, stKey.Asset.Code)
				assert.Equal(t, assetIssuer, stKey.Asset.Issuer)
				assert.Equal(t, flatFee, st.FlatFee)
				assert.Equal(t, percentFee, st.PercentFee)
			}
			action.Validate()
			So(action.Err, ShouldBeNil)
			action.Apply()
			check(action.AdminAction)
			Convey("update", func() {
				flatFee = 99
				data["flat_fee"] = strconv.FormatInt(int64(flatFee), 10)
				updateAction := NewSetCommissionAction(NewAdminAction(data, historyQ))
				updateAction.Validate()
				So(action.Err, ShouldBeNil)
				updateAction.Apply()
				check(updateAction.AdminAction)
			})
			Convey("delete", func() {
				data["delete"] = "true"
				deleteAction := NewSetCommissionAction(NewAdminAction(data, historyQ))
				deleteAction.Validate()
				So(action.Err, ShouldBeNil)
				deleteAction.Apply()
				So(deleteAction.Err, ShouldBeNil)
				var sts []history.Commission
				err = historyQ.Commissions().ForAccount(from).Select(&sts)
				assert.Nil(t, err)
				assert.Equal(t, 0, len(sts))
			})
			historyQ.DeleteCommissions()
		})

	})
}
