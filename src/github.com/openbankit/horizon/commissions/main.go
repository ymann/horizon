package commissions

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"database/sql"
	"errors"
	"math"
	"math/big"
)

type CommissionsManager struct {
	SharedCache *cache.SharedCache
	HistoryQ    history.QInterface
}

func New(sharedCache *cache.SharedCache, histQ history.QInterface) *CommissionsManager {
	return &CommissionsManager{
		SharedCache: sharedCache,
		HistoryQ:            histQ,
	}
}

// sets commission for each operation
func (cm *CommissionsManager) SetCommissions(env *xdr.TransactionEnvelope) (err error) {
	if env == nil {
		return errors.New("SetCommissions: tx must not be nil")
	}
	env.OperationFees = make([]xdr.OperationFee, len(env.Tx.Operations))
	for i, op := range env.Tx.Operations {
		commission, err := cm.CalculateCommissionForOperation(env.Tx.SourceAccount, op)
		if err != nil {
			return err
		}
		env.OperationFees[i] = *commission
	}
	return
}

// calculates operation fee based on operation source or (if is not set) on tx source and operations data
func (cm *CommissionsManager) CalculateCommissionForOperation(txSource xdr.AccountId, op xdr.Operation) (*xdr.OperationFee, error) {
	opSource := txSource
	if op.SourceAccount != nil {
		opSource = *op.SourceAccount
	}
	switch op.Body.Type {
	case xdr.OperationTypePayment:
		payment := op.Body.MustPaymentOp()
		return cm.CalculateCommission(opSource, payment.Destination, payment.Amount, payment.Asset)
	case xdr.OperationTypePathPayment:
		payment := op.Body.MustPathPaymentOp()
		return cm.CalculateCommission(opSource, payment.Destination, payment.DestAmount, payment.DestAsset)
	default:
		return &xdr.OperationFee{
			Type: xdr.OperationFeeTypeOpFeeNone,
		}, nil
	}
}

// gets account's type from core db, if account does not exist and mustExists - returns error, if mustExists false - xdr.AccountTypeAccountAnonymousUser
func (cm *CommissionsManager) getAccountType(accountId string, mustExists bool) (int32, error) {
	account, err := cm.SharedCache.AccountHistoryCache.Get(accountId)
	if err != nil {
		if err == sql.ErrNoRows && !mustExists {
			return int32(xdr.AccountTypeAccountAnonymousUser), nil
		}
		log.WithField("accountId", accountId).WithError(err).Error("Failed to get account type")
		return 0, err
	}
	return int32(account.AccountType), nil
}

// returns commission with highest weight and lowest fee from db based on keys created from params
func (cm *CommissionsManager) getCommission(sourceId, destinationId xdr.AccountId, amount xdr.Int64, asset xdr.Asset) (*history.Commission, error) {
	sourceAccountType, err := cm.getAccountType(sourceId.Address(), true)
	if err != nil {
		return nil, err
	}

	destAccountType, err := cm.getAccountType(destinationId.Address(), false)
	if err != nil {
		return nil, err
	}

	baseAsset := assets.ToBaseAsset(asset)
	keys := history.CreateCommissionKeys(sourceId.Address(), destinationId.Address(), int32(sourceAccountType), int32(destAccountType), baseAsset)
	commissions, err := cm.HistoryQ.GetHighestWeightCommission(keys)
	if err != nil {
		log.WithStack(err).WithError(err).Error("Failed to GetHighestWeightCommission")
		return nil, err
	}
	log.WithField("commissions", commissions).Debug("Got filtered commissions by weight")
	histCommission := getSmallestFee(commissions, amount)

	return histCommission, nil
}

// selects smallest fee from commissions slice
func getSmallestFee(commissions []history.Commission, amount xdr.Int64) *history.Commission {
	var histCommission *history.Commission
	fee := xdr.Int64(math.MaxInt64)
	for _, comm := range commissions {
		newFee := calculatePercentFee(amount, xdr.Int64(comm.PercentFee)) + xdr.Int64(comm.FlatFee)
		if newFee <= fee {
			fee = newFee
			histCommission = new(history.Commission)
			*histCommission = comm
		}
	}
	return histCommission
}

// returns xdr.Operation fee with highest weight and lowest fee from db based on keys created from params
func (cm *CommissionsManager) CalculateCommission(source, destination xdr.AccountId, amount xdr.Int64, asset xdr.Asset) (*xdr.OperationFee, error) {
	commission, err := cm.getCommission(source, destination, amount, asset)
	if err != nil {
		log.WithStack(err).WithError(err).Error("Failed to getCommission")
		return nil, err
	}
	if commission == nil {
		return &xdr.OperationFee{
			Type: xdr.OperationFeeTypeOpFeeNone,
		}, nil
	}
	percent := xdr.Int64(commission.PercentFee)
	flatFee := xdr.Int64(commission.FlatFee)
	return &xdr.OperationFee{
		Type: xdr.OperationFeeTypeOpFeeCharged,
		Fee: &xdr.OperationFeeFee{
			Asset:          asset,
			AmountToCharge: calculatePercentFee(amount, percent) + xdr.Int64(flatFee),
			PercentFee:     &percent,
			FlatFee:        &flatFee,
		},
	}, nil
}

// calculates percentI from paymentAmountI
func calculatePercentFee(paymentAmountI, percentI xdr.Int64) xdr.Int64 {
	zero := xdr.Int64(0)
	if percentI == zero {
		return zero
	}
	// (amount/100) * percent
	paymentAmount := big.NewRat(int64(paymentAmountI), 100)
	percentR := big.NewRat(int64(percentI), amount.One)
	var result big.Rat
	result.Mul(paymentAmount, percentR)
	return xdr.Int64(result.Num().Int64() / result.Denom().Int64())
}
