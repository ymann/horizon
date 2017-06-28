package transactions

import (
	"github.com/openbankit/go-base/build"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPathPaymentOpFrame(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	historyQ := &history.Q{
		tt.HorizonRepo(),
	}
	coreQ := &core.Q{
		tt.CoreRepo(),
	}
	config := test.NewTestConfig()

	manager := NewManager(coreQ, historyQ, nil, &config, &cache.SharedCache{
		AccountHistoryCache: cache.NewHistoryAccount(historyQ),
	})

	root := test.BankMasterSeed()

	newAccount, err := keypair.Random()
	assert.Nil(t, err)

	Convey("Invalid source asset", t, func() {
		payment := build.Payment(build.Destination{newAccount.Address()}, build.PayWithPath{
			Asset: build.Asset{
				Code:   "USD",
				Issuer: root.Address(),
			},
			MaxAmount: "1000000",
		})
		checkPaymentInvalidAsset(payment, root, manager)
	})
	Convey("Invalid dest asset", t, func() {
		payment := build.Payment(build.Destination{newAccount.Address()}, build.PayWithPath{
			Asset: build.Asset{
				Code:   "UAH",
				Issuer: root.Address(),
			},
			Path: []build.Asset{
				build.Asset{
					Code:   "USD",
					Issuer: root.Address(),
				},
			},
			MaxAmount: "1000000",
		})
		checkPaymentInvalidAsset(payment, root, manager)
	})
	Convey("Invalid path asset", t, func() {
		payment := build.Payment(build.Destination{newAccount.Address()}, build.PayWithPath{
			Asset: build.Asset{
				Code:   "UAH",
				Issuer: root.Address(),
			},
			Path: []build.Asset{
				build.Asset{
					Code:   "USD",
					Issuer: root.Address(),
				},
				build.Asset{
					Code:   "AUAH",
					Issuer: root.Address(),
				},
			},
			MaxAmount: "1000000",
		})
		checkPaymentInvalidAsset(payment, root, manager)
	})
}

func checkPaymentInvalidAsset(payment build.PaymentBuilder, root *keypair.Full, manager *Manager) {
	tx := build.Transaction(payment, build.Sequence{1}, build.SourceAccount{root.Address()})
	txE := NewTransactionFrame(&EnvelopeInfo{
		Tx: tx.Sign(root.Seed()).E,
	})
	opFrame := NewOperationFrame(&txE.Tx.Tx.Operations[0], txE, 1, time.Now())
	isValid, err := opFrame.CheckValid(manager)
	So(err, ShouldBeNil)
	So(isValid, ShouldBeFalse)
	So(opFrame.GetResult().Result.MustTr().MustPathPaymentResult().Code, ShouldEqual, xdr.PathPaymentResultCodePathPaymentMalformed)
	So(opFrame.GetResult().Info.GetError(), ShouldEqual, ASSET_NOT_ALLOWED.Error())
}
