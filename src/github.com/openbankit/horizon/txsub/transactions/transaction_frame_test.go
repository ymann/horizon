package transactions

import (
	"github.com/openbankit/go-base/build"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/admin"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/txsub/results"
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransactionFrame(t *testing.T) {

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	Convey("Test Transaction Frame:", t, func() {
		signer, err := keypair.Random()
		So(err, ShouldBeNil)
		Convey("Several non admin operations", func() {
			newAccount, err := keypair.Random()
			So(err, ShouldBeNil)
			payment := build.Payment(build.CreditAmount{"USD", newAccount.Address(), "50.0"},
				build.Destination{newAccount.Address()})
			createAccount := build.CreateAccount(build.Destination{newAccount.Address()})
			tx := build.Transaction(payment, createAccount, build.Sequence{1}, build.SourceAccount{signer.Address()})
			txFrame := NewTransactionFrame(&EnvelopeInfo{
				Tx: tx.Sign(signer.Seed()).E,
			})
			isValid, err := txFrame.checkTransaction()
			So(err, ShouldBeNil)
			So(isValid, ShouldBeTrue)
		})
		Convey("Admin op", func() {
			opObjData := map[string]interface{}{
				string(admin.SubjectCommission): map[string]interface{}{},
			}
			opData, err := json.Marshal(opObjData)
			So(err, ShouldBeNil)
			adminOp := build.AdministrativeOp(build.OpLongData{OpData: string(opData)})
			So(adminOp.Err, ShouldBeNil)
			tx := build.Transaction(adminOp, build.Sequence{1}, build.SourceAccount{signer.Address()})
			So(tx.Err, ShouldBeNil)
			txFrame := NewTransactionFrame(&EnvelopeInfo{
				Tx: tx.Sign(signer.Seed()).E,
			})
			isValid, err := txFrame.checkTransaction()
			So(err, ShouldBeNil)
			So(isValid, ShouldBeTrue)
		})
		Convey("Admin op & non admin", func() {
			opObjData := map[string]interface{}{
				string(admin.SubjectCommission): map[string]interface{}{},
			}
			opData, err := json.Marshal(opObjData)
			So(err, ShouldBeNil)
			adminOp := build.AdministrativeOp(build.OpLongData{OpData: string(opData)})
			So(adminOp.Err, ShouldBeNil)
			newAccount, err := keypair.Random()
			createAccount := build.CreateAccount(build.Destination{newAccount.Address()})
			tx := build.Transaction(adminOp, createAccount, build.Sequence{1}, build.SourceAccount{signer.Address()})
			So(tx.Err, ShouldBeNil)
			txFrame := NewTransactionFrame(&EnvelopeInfo{
				Tx: tx.Sign(signer.Seed()).E,
			})
			isValid, err := txFrame.checkTransaction()
			So(err, ShouldBeNil)
			So(isValid, ShouldBeFalse)
			txResult := txFrame.GetResult()
			So(txResult.TransactionErrorInfo, ShouldNotBeNil)
			assert.Equal(t, results.AdditionalErrorInfoStrError("Administrative op must be only op in tx"), *txResult.TransactionErrorInfo)
		})
	})
}
