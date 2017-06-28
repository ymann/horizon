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

func TestSetOptionsOpFrame(t *testing.T) {
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

	root := test.BankMasterSeed()

	manager := NewManager(coreQ, historyQ, nil, &config, &cache.SharedCache{
		AccountHistoryCache: cache.NewHistoryAccount(historyQ),
	})

	Convey("Success", t, func() {
		setOptions := build.SetOptions()
		tx := build.Transaction(setOptions, build.Sequence{1}, build.SourceAccount{root.Address()})
		txE := NewTransactionFrame(&EnvelopeInfo{
			Tx: tx.Sign(root.Seed()).E,
		})
		opFrame := NewOperationFrame(&txE.Tx.Tx.Operations[0], txE, 1, time.Now())
		isValid, err := opFrame.CheckValid(manager)
		So(err, ShouldBeNil)
		So(isValid, ShouldBeTrue)
		So(opFrame.GetResult().Result.MustTr().MustSetOptionsResult().Code, ShouldEqual, xdr.SetOptionsResultCodeSetOptionsSuccess)
	})
}
