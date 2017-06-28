package txsub

import (
	"testing"

	"github.com/openbankit/go-base/build"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	subResults "github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/sequence"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"time"
)

func TestTxsub(t *testing.T) {

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	Convey("txsub.System", t, func() {
		ctx := test.Context()
		submitter := &MockSubmitter{}
		results := &MockResultProvider{}
		sequences := &MockSequenceProvider{}

		system := &System{
			Pending:           NewDefaultSubmissionList(),
			Submitter:         submitter,
			Results:           results,
			Sequences:         sequences,
			SubmissionQueue:   sequence.NewManager(),
			NetworkPassphrase: build.TestNetwork.Passphrase,
		}

		account, err := keypair.Random()
		So(err, ShouldBeNil)

		noResults := Result{Err: subResults.ErrNoResults}
		successTx := getSuccessResult(account)

		badSeq := SubmissionResult{
			Err: subResults.ErrBadSequence,
		}

		sequences.Results = map[string]uint64{
			account.Address(): 0,
		}

		Convey("Submit", func() {
			Convey("returns the result provided by the ResultProvider", func() {
				results.Results = []Result{successTx}
				r := <-system.Submit(ctx, successTx.EnvelopeXDR)

				So(r.Err, ShouldBeNil)
				So(r.Hash, ShouldEqual, successTx.Hash)
				So(submitter.WasSubmittedTo, ShouldBeFalse)
			})

			Convey("returns the error from submission if no result is found by hash and the submitter returns an error", func() {
				submitter.R.Err = errors.New("busted for some reason")
				r := <-system.Submit(ctx, successTx.EnvelopeXDR)

				So(r.Err, ShouldNotBeNil)
				So(submitter.WasSubmittedTo, ShouldBeTrue)
				So(system.Metrics.SuccessfulSubmissionsMeter.Count(), ShouldEqual, 0)
				So(system.Metrics.FailedSubmissionsMeter.Count(), ShouldEqual, 1)
				So(system.Metrics.SubmissionTimer.Count(), ShouldEqual, 1)
			})

			Convey("if the error is bad_seq and the result at the transaction's sequence number is for the same hash, return result", func() {
				submitter.R = badSeq
				results.Results = []Result{noResults, successTx}

				r := <-system.Submit(ctx, successTx.EnvelopeXDR)

				So(r.Err, ShouldBeNil)
				So(r.Hash, ShouldEqual, successTx.Hash)
				So(submitter.WasSubmittedTo, ShouldBeTrue)
			})

			Convey("if error is bad_seq and no result is found, return error", func() {
				submitter.R = badSeq
				r := <-system.Submit(ctx, successTx.EnvelopeXDR)

				So(r.Err, ShouldNotBeNil)
				So(submitter.WasSubmittedTo, ShouldBeTrue)
			})

			Convey("if no result found and no error submitting, add to open transaction list", func() {
				_ = system.Submit(ctx, successTx.EnvelopeXDR)
				pending := system.Pending.Pending(ctx)
				So(len(pending), ShouldEqual, 1)
				So(pending[0], ShouldEqual, successTx.Hash)
				So(system.Metrics.SuccessfulSubmissionsMeter.Count(), ShouldEqual, 1)
				So(system.Metrics.FailedSubmissionsMeter.Count(), ShouldEqual, 0)
				So(system.Metrics.SubmissionTimer.Count(), ShouldEqual, 1)
			})
		})

		Convey("Tick", func() {

			Convey("no-ops if there are no open submissions", func() {
				system.Tick(ctx)
			})

			Convey("finishes any available transactions", func() {
				l := make(chan Result, 1)
				system.Pending.Add(ctx, successTx.Hash, l)
				system.Tick(ctx)
				So(len(l), ShouldEqual, 0)
				So(len(system.Pending.Pending(ctx)), ShouldEqual, 1)

				results.Results = []Result{successTx}
				system.Tick(ctx)

				So(len(l), ShouldEqual, 1)
				So(len(system.Pending.Pending(ctx)), ShouldEqual, 0)
			})

			Convey("removes old submissions that have timed out", func() {
				l := make(chan Result, 1)
				system.SubmissionTimeout = 100 * time.Millisecond
				system.Pending.Add(ctx, successTx.Hash, l)
				<-time.After(101 * time.Millisecond)
				system.Tick(ctx)

				So(len(system.Pending.Pending(ctx)), ShouldEqual, 0)
				So(len(l), ShouldEqual, 1)
				<-l
				select {
				case _, stillOpen := <-l:
					So(stillOpen, ShouldBeFalse)
				default:
					panic("could not read from listener")
				}

			})
		})

	})
}

func getSuccessResult(account *keypair.Full) Result {
	createAccount := build.CreateAccount(build.Destination{account.Address()})
	tx := build.Transaction(createAccount, build.Sequence{1}, build.SourceAccount{account.Address()}, build.Network{build.TestNetwork.Passphrase})
	hash, err := tx.HashHex()
	So(err, ShouldBeNil)
	txE := tx.Sign(account.Seed())
	envelopeXdr, err := txE.Base64()
	So(err, ShouldBeNil)
	var result xdr.TransactionResult
	result.Result = xdr.TransactionResultResult{
		Code: xdr.TransactionResultCodeTxSuccess,
		Results: &[]xdr.OperationResult{
			xdr.OperationResult{
				Code: xdr.OperationResultCodeOpInner,
				Tr: &xdr.OperationResultTr{
					Type: xdr.OperationTypeCreateAccount,
					CreateAccountResult: &xdr.CreateAccountResult{
						Code: xdr.CreateAccountResultCodeCreateAccountSuccess,
					},
				},
			},
		},
	}
	resultXdr, err := xdr.MarshalBase64(result)
	So(err, ShouldBeNil)
	return Result{
		Hash:           hash,
		LedgerSequence: 2,
		EnvelopeXDR:    envelopeXdr,
		ResultXDR:      resultXdr,
	}
}
