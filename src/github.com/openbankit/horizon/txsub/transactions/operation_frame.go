package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/txsub/results"
	"database/sql"
	"github.com/go-errors/errors"
	"time"
)

var (
	ASSET_NOT_ALLOWED     = errors.New("asset is not allowed")
	OPERATION_NOT_ALLOWED = errors.New("operation is not allowed")
)

type OperationInterface interface {
	DoCheckValid(manager *Manager) (bool, error)
	DoRollbackCachedData(manager *Manager) error
}

type OperationFrame struct {
	Op            *xdr.Operation
	ParentTxFrame *TransactionFrame
	Result        *results.OperationResult
	innerOp       OperationInterface
	SourceAccount *history.Account
	Index         int
	log           *log.Entry
	now           *time.Time
}

func NewOperationFrame(op *xdr.Operation, tx *TransactionFrame, index int, now time.Time) OperationFrame {
	return OperationFrame{
		Op:            op,
		ParentTxFrame: tx,
		Result:        &results.OperationResult{},
		Index:         index,
		log:           log.WithField("service", op.Body.Type.String()),
		SourceAccount: new(history.Account),
		now:           &now,
	}
}

func (opFrame *OperationFrame) GetInnerOp() (OperationInterface, error) {
	if opFrame.innerOp != nil {
		return opFrame.innerOp, nil
	}
	var innerOp OperationInterface
	switch opFrame.Op.Body.Type {
	case xdr.OperationTypeCreateAccount:
		innerOp = NewCreateAccountOpFrame(opFrame)
	case xdr.OperationTypePayment:
		innerOp = NewPaymentOpFrame(opFrame)
	case xdr.OperationTypePathPayment:
		innerOp = NewPathPaymentOpFrame(opFrame)
	case xdr.OperationTypeManageOffer:
		innerOp = NewManageOfferOpFrame(opFrame)
	case xdr.OperationTypeCreatePassiveOffer:
		innerOp = NewCreatePassiveOfferOpFrame(opFrame)
	case xdr.OperationTypeSetOptions:
		innerOp = NewSetOptionsOpFrame(opFrame)
	case xdr.OperationTypeChangeTrust:
		innerOp = NewChangeTrustOpFrame(opFrame)
	case xdr.OperationTypeAllowTrust:
		innerOp = NewAllowTrustOpFrame(opFrame)
	case xdr.OperationTypeAccountMerge:
		innerOp = NewAccountMergeOpFrame(opFrame)
	case xdr.OperationTypeInflation:
		innerOp = NewInflationOpFrame(opFrame)
	case xdr.OperationTypeManageData:
		innerOp = NewManageDataOpFrame(opFrame)
	case xdr.OperationTypeAdministrative:
		innerOp = NewAdministrativeOpFrame(opFrame)
	case xdr.OperationTypePaymentReversal:
		innerOp = NewPaymentReversalOpFrame(opFrame)
	case xdr.OperationTypeExternalPayment:
		innerOp = NewExternalPaymentOpFrame(opFrame)
	default:
		return nil, errors.New("unknown operation")
	}
	opFrame.innerOp = innerOp
	return opFrame.innerOp, nil
}

func (op *OperationFrame) GetResult() results.OperationResult {
	return *op.Result
}

func (opFrame *OperationFrame) CheckValid(manager *Manager) (bool, error) {
	var sourceAddress string
	if opFrame.Op.SourceAccount != nil {
		sourceAddress = opFrame.Op.SourceAccount.Address()
	} else {
		sourceAddress = opFrame.ParentTxFrame.Tx.Tx.SourceAccount.Address()
	}

	// check if source account exists
	var err error
	opFrame.SourceAccount, err = manager.AccountHistoryCache.Get(sourceAddress)
	if err != nil {
		if err == sql.ErrNoRows {
			opFrame.Result.Result = xdr.OperationResult{
				Code: xdr.OperationResultCodeOpNoAccount,
			}
			return false, nil
		}
		return false, err
	}

	opFrame.log.WithField("sourceAccount", opFrame.SourceAccount.Address).Debug("Loaded source account")
	// prepare result for op Result
	opFrame.Result.Result = xdr.OperationResult{
		Code: xdr.OperationResultCodeOpInner,
		Tr: &xdr.OperationResultTr{
			Type: opFrame.Op.Body.Type,
		},
	}

	innerOp, err := opFrame.GetInnerOp()
	if err != nil {
		return false, err
	}

	// validate
	return innerOp.DoCheckValid(manager)
}

// default implementation
func (opFrame *OperationFrame) DoRollbackCachedData(manager *Manager) error {
	return nil
}

func (opFrame *OperationFrame) RollbackCachedData(manager *Manager) error {
	innerOp, err := opFrame.GetInnerOp()
	if err != nil {
		return err
	}

	// rollback
	return innerOp.DoRollbackCachedData(manager)
}
