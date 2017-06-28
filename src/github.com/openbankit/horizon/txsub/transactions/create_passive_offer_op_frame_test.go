package transactions

import (
	"github.com/openbankit/go-base/build"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestCreatePassiveOfferOpFrame(t *testing.T) {
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

	validAsset := build.Asset{
		Code:   "UAH",
		Issuer: root.Address(),
	}
	invalidAsset := build.Asset{
		Code:   "USD",
		Issuer: root.Address(),
	}

	Convey("Operation is not allowed", t, func() {
		createPassiveOffer := build.CreatePassiveOffer(build.Rate{
			Price:   build.Price("10"),
			Selling: invalidAsset,
			Buying:  validAsset,
		}, build.Amount("1000"))
		tx := build.Transaction(createPassiveOffer, build.Sequence{1}, build.SourceAccount{root.Address()})
		txE := NewTransactionFrame(&EnvelopeInfo{
			Tx: tx.Sign(root.Seed()).E,
		})
		opFrame := NewOperationFrame(&txE.Tx.Tx.Operations[0], txE, 1, time.Now())
		isValid, err := opFrame.CheckValid(manager)
		So(err, ShouldBeNil)
		So(isValid, ShouldBeFalse)
		So(opFrame.GetResult().Result.MustTr().MustCreatePassiveOfferResult().Code, ShouldEqual, xdr.ManageOfferResultCodeManageOfferMalformed)
		So(opFrame.GetResult().Info.GetError(), ShouldEqual, OPERATION_NOT_ALLOWED.Error())
	})
}
