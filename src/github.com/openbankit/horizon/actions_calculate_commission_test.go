package horizon

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/test"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type RequestBuilder struct {
	From   string
	To     string
	Asset  details.Asset
	Amount string
}

func (b *RequestBuilder) Build() string {
	return fmt.Sprintf("/commission/calculate?amount=%s&from=%s&to=%s&asset_type=%s&asset_code=%s&asset_issuer=%s",
		b.Amount, b.From, b.To, b.Asset.Type, b.Asset.Code, b.Asset.Issuer)
}

func TestActionsCalculateCommission(t *testing.T) {
	app := NewTestApp()
	defer app.Close()
	rh := NewRequestHelper(app)
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	Convey("Calculate commission Actions:", t, func() {
		from, err := keypair.Random()
		So(err, ShouldBeNil)
		request := RequestBuilder{
			From:   from.Address(),
			To:     from.Address(),
			Amount: "101.01",
			Asset: details.Asset{
				Type:   assets.MustString(xdr.AssetTypeAssetTypeCreditAlphanum4),
				Code:   "EUR",
				Issuer: from.Address(),
			},
		}
		Convey("Invalid from", func() {
			request.From = "1"
			w := rh.Get(request.Build(), test.RequestHelperNoop)
			So(w.Code, ShouldEqual, 400)
			So(w.Body, ShouldBeProblem, problem.BadRequest, "from")
		})
		Convey("Invalid to", func() {
			request.To = "random_string"
			w := rh.Get(request.Build(), test.RequestHelperNoop)
			So(w.Code, ShouldEqual, 400)
			So(w.Body, ShouldBeProblem, problem.BadRequest, "to")
		})
		Convey("Invalid amount", func() {
			request.Amount = "-1.01"
			w := rh.Get(request.Build(), test.RequestHelperNoop)
			So(w.Code, ShouldEqual, 400)
			So(w.Body, ShouldBeProblem, problem.BadRequest, "amount")
		})
		Convey("Invalid asset_issuer", func() {
			request.Asset.Issuer = "-1.01"
			w := rh.Get(request.Build(), test.RequestHelperNoop)
			So(w.Code, ShouldEqual, 400)
			So(w.Body, ShouldBeProblem, problem.BadRequest, "asset_issuer")
		})
	})
}
