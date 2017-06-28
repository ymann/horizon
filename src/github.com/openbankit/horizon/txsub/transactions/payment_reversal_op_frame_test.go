package transactions

import (
	"github.com/openbankit/go-base/build"
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/core"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/guregu/null"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"testing"
	"time"
)

func TestPaymentReversalOpFrame(t *testing.T) {
	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	historyQ := history.QMock{}
	coreQ := core.QMock{}
	config := test.NewTestConfig()

	root := test.BankMasterSeed()

	manager := NewManager(&coreQ, &historyQ, nil, &config, &cache.SharedCache{
		AccountHistoryCache: cache.NewHistoryAccount(&historyQ),
	})

	paymentSenderKP, err := keypair.Random()
	assert.Nil(t, err)

	paymentID := rand.Int63()
	paymentAmount := "125.78"
	commissionAmount := "22.34"
	assetCode := "USD"
	paymentReversal := build.PaymentReversal(build.CreditAmount{
		Code:   assetCode,
		Issuer: root.Address(),
		Amount: paymentAmount,
	}, build.CommissionAmount{
		Amount: commissionAmount,
	}, build.PaymentID{
		ID: paymentID,
	}, build.PaymentSender{
		AddressOrSeed: paymentSenderKP.Address(),
	})

	tx := build.Transaction(paymentReversal, build.Sequence{1}, build.SourceAccount{root.Address()})
	txE := NewTransactionFrame(&EnvelopeInfo{
		Tx: tx.Sign(root.Seed()).E,
	})

	validOperation := txE.Tx.Tx.Operations[0]

	now := time.Now()

	historyQ.On("AccountByAddress", root.Address()).Return(history.Account{
		Address: root.Address(),
	}, nil)
	Convey("Negative amount", t, func() {
		operation := validOperation
		paymentReversalOp := *operation.Body.PaymentReversalOp
		operation.Body.PaymentReversalOp = &paymentReversalOp
		operation.Body.PaymentReversalOp.Amount = xdr.Int64(-100)
		opFrame := NewOperationFrame(&operation, txE, 1, now)
		isValid, err := opFrame.CheckValid(manager)
		So(err, ShouldBeNil)
		So(isValid, ShouldBeFalse)
		So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, xdr.PaymentReversalResultCodePaymentReversalMalformed)
	})
	Convey("Negative commission amount", t, func() {
		operation := validOperation
		paymentReversalOp := *operation.Body.PaymentReversalOp
		operation.Body.PaymentReversalOp = &paymentReversalOp
		operation.Body.PaymentReversalOp.CommissionAmount = xdr.Int64(-10)
		opFrame := NewOperationFrame(&operation, txE, 1, now)
		isValid, err := opFrame.CheckValid(manager)
		So(err, ShouldBeNil)
		So(isValid, ShouldBeFalse)
		So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, xdr.PaymentReversalResultCodePaymentReversalMalformed)
	})
	Convey("Given valid payment reversal op", t, func() {
		log.Error("Given valid payment reversal op")
		operation := validOperation
		opFrame := NewOperationFrame(&operation, txE, 1, now)
		Convey("Failed to get payment", func() {
			expectedError := errors.New("Failed to get payment from db")
			historyQ.On("OperationByID", mock.Anything, paymentID).Return(expectedError).Once()
			isValid, err := opFrame.CheckValid(manager)
			So(err, ShouldNotBeNil)
			So(expectedError.Error(), ShouldEqual, err.Error())
			So(isValid, ShouldBeFalse)
		})
		Convey("Payment does not exists", func() {
			historyQ.On("OperationByID", mock.Anything, paymentID).Return(sql.ErrNoRows).Once()
			isValid, err := opFrame.CheckValid(manager)
			So(err, ShouldBeNil)
			So(isValid, ShouldBeFalse)
			So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, xdr.PaymentReversalResultCodePaymentReversalPaymentDoesNotExists)
		})
		Convey("Operation with same ID, but not payment", func() {
			historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
				op := args.Get(0).(*history.Operation)
				op.Type = xdr.OperationTypeAllowTrust
			}).Return(nil).Once()
			isValid, err := opFrame.CheckValid(manager)
			So(err, ShouldBeNil)
			So(isValid, ShouldBeFalse)
			So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, xdr.PaymentReversalResultCodePaymentReversalPaymentDoesNotExists)
		})
		Convey("Given valid stored payment", func() {
			validStoredPayment := history.Operation{
				Type:          xdr.OperationTypePayment,
				ClosedAt:      now,
				SourceAccount: operation.Body.PaymentReversalOp.PaymentSource.Address(),
			}

			validPaymentDetails := details.Payment{
				From:   validStoredPayment.SourceAccount,
				To:     root.Address(),
				Amount: paymentAmount,
				Asset: details.Asset{
					Type:   "credit_alphanum4",
					Code:   assetCode,
					Issuer: root.Address(),
				},
				Fee: details.Fee{
					AmountCharged: &commissionAmount,
				},
			}

			jsonDetails, err := json.Marshal(validPaymentDetails)
			assert.Nil(t, err)
			validStoredPayment.DetailsString = null.StringFrom(string(jsonDetails))
			Convey("Failed to get max reversal duration from db", func() {
				storedPayment := validStoredPayment
				historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
					op := args.Get(0).(*history.Operation)
					*op = storedPayment
				}).Return(nil).Once()
				expectedError := errors.New("Failed to get max reversal duration")
				historyQ.On("OptionsByName", history.OPTIONS_MAX_REVERSAL_DURATION).Return(nil, expectedError).Once()
				isValid, err := opFrame.CheckValid(manager)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, expectedError.Error())
				So(isValid, ShouldBeFalse)
			})
			Convey("Failed to parse options", func() {
				storedPayment := validStoredPayment
				historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
					op := args.Get(0).(*history.Operation)
					*op = storedPayment
				}).Return(nil).Once()
				option := history.Options{}
				historyQ.On("OptionsByName", history.OPTIONS_MAX_REVERSAL_DURATION).Return(&option, nil).Once()
				isValid, err := opFrame.CheckValid(manager)
				So(err, ShouldNotBeNil)
				So(isValid, ShouldBeFalse)
			})
			Convey("Valid max reversal duration", func() {
				opChecker := func(storedPayment history.Operation, expectedCode xdr.PaymentReversalResultCode) {
					historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
						op := args.Get(0).(*history.Operation)
						*op = storedPayment
					}).Return(nil).Once()
					isValid, err := opFrame.CheckValid(manager)
					So(err, ShouldBeNil)
					So(isValid, ShouldBeFalse)
					So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, expectedCode)
				}
				historyQ.On("OptionsByName", history.OPTIONS_MAX_REVERSAL_DURATION).Return(nil, nil).Once()
				Convey("Can't reverse - payment expired", func() {
					storedPayment := validStoredPayment
					storedPayment.ClosedAt = now.Add(time.Duration(-int64(MAX_REVERSE_TIME))).Add(time.Duration(-1) * time.Second)
					opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalPaymentExpired)
				})
				Convey("Invalid payment reversal source", func() {
					storedPayment := validStoredPayment
					storedPayment.SourceAccount = root.Address()
					opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidSource)
				})
				Convey("Invalid payment details", func() {
					storedPayment := validStoredPayment
					storedPayment.DetailsString = null.StringFrom("")
					historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
						op := args.Get(0).(*history.Operation)
						*op = storedPayment
					}).Return(nil).Once()
					isValid, err := opFrame.CheckValid(manager)
					So(err, ShouldNotBeNil)
					So(isValid, ShouldBeFalse)
				})
				Convey("Invalid payment source", func() {
					storedPayment := validStoredPayment
					paymentDetails := validPaymentDetails
					paymentDetails.To = paymentSenderKP.Address()
					jsonDetails, err = json.Marshal(paymentDetails)
					assert.Nil(t, err)
					storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
					opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidPaymentSender)
				})
				Convey("Invalid amount", func() {
					storedPayment := validStoredPayment
					paymentDetails := validPaymentDetails
					paymentDetails.Amount = "102"
					jsonDetails, err = json.Marshal(paymentDetails)
					assert.Nil(t, err)
					storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
					opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidAmount)
				})
				Convey("Invalid commission", func() {
					storedPayment := validStoredPayment
					paymentDetails := validPaymentDetails
					Convey("Invalid amount", func() {
						amountCharged := "102"
						paymentDetails.Fee.AmountCharged = &amountCharged
						jsonDetails, err = json.Marshal(paymentDetails)
						assert.Nil(t, err)
						storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
						opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidCommission)
					})
					Convey("Fee was not charged", func() {
						paymentDetails.Fee.AmountCharged = nil
						jsonDetails, err = json.Marshal(paymentDetails)
						assert.Nil(t, err)
						storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
						opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidCommission)
					})
				})
				Convey("Invalid asset", func() {
					storedPayment := validStoredPayment
					paymentDetails := validPaymentDetails
					Convey("Invalid asset code", func() {
						paymentDetails.Asset.Code = "EUR"
						jsonDetails, err = json.Marshal(paymentDetails)
						assert.Nil(t, err)
						storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
						opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidAsset)
					})
					Convey("Invalid asset issuer", func() {
						paymentDetails.Asset.Issuer = paymentSenderKP.Address()
						jsonDetails, err = json.Marshal(paymentDetails)
						assert.Nil(t, err)
						storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
						opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidAsset)
					})
					Convey("Invalid asset type", func() {
						paymentDetails.Asset.Type = "credit_alphanum12"
						jsonDetails, err = json.Marshal(paymentDetails)
						assert.Nil(t, err)
						storedPayment.DetailsString = null.StringFrom(string(jsonDetails))
						opChecker(storedPayment, xdr.PaymentReversalResultCodePaymentReversalInvalidAsset)
					})

				})
				Convey("Success", func() {
					storedPayment := validStoredPayment
					historyQ.On("OperationByID", mock.Anything, paymentID).Run(func(args mock.Arguments) {
						op := args.Get(0).(*history.Operation)
						*op = storedPayment
					}).Return(nil).Once()
					isValid, err := opFrame.CheckValid(manager)
					So(err, ShouldBeNil)
					So(isValid, ShouldBeTrue)
					So(opFrame.GetResult().Result.MustTr().MustPaymentReversalResult().Code, ShouldEqual, xdr.PaymentReversalResultCodePaymentReversalSuccess)
				})
			})
		})

	})
}
