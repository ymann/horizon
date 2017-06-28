package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/txsub/results"
	"time"
)

type TransactionFrame struct {
	Tx       *xdr.TransactionEnvelope
	TxHash   string
	txResult *results.RestrictedTransactionError
	log      *log.Entry
}

func NewTransactionFrame(envelopeInfo *EnvelopeInfo) *TransactionFrame {
	return &TransactionFrame{
		Tx:  envelopeInfo.Tx,
		TxHash: envelopeInfo.ContentHash,
		log: log.WithField("service", "transaction_frame"),
	}
}

func (t *TransactionFrame) CheckValid(manager *Manager) (bool, error) {
	t.log.Debug("Checking transaction")
	isTxValid, err := t.checkTransaction()
	if !isTxValid || err != nil {
		return isTxValid, err
	}

	t.log.Debug("Transaction is valid. Checking operations.")
	return t.checkOperations(manager)
}

func (t *TransactionFrame) checkTransaction() (bool, error) {
	// transaction can only have one adminOp
	if len(t.Tx.Tx.Operations) == 1 {
		return true, nil
	}

	for _, op := range t.Tx.Tx.Operations {
		if op.Body.Type == xdr.OperationTypeAdministrative {
			var err error
			t.txResult, err = results.NewRestrictedTransactionErrorTx(xdr.TransactionResultCodeTxFailed, results.AdditionalErrorInfoStrError("Administrative op must be only op in tx"))
			if err != nil {
				return false, err
			}
			return false, nil
		}
	}

	return true, nil
}

func (t *TransactionFrame) checkOperations(manager *Manager) (bool, error) {
	opFrames := make([]OperationFrame, len(t.Tx.Tx.Operations))
	isValid := true
	now := time.Now()
	for i, op := range t.Tx.Tx.Operations {
		opFrames[i] = NewOperationFrame(&op, t, i, now)
		isOpValid, err := opFrames[i].CheckValid(manager)
		// failed to validate
		if err != nil {
			t.log.WithField("operation_i", i).WithError(err).Error("Failed to validate")
			return false, err
		}

		if !isOpValid {
			t.log.WithField("operation_i", i).WithField("result", opFrames[i].GetResult()).Debug("Is not valid")
			isValid = false
		}
	}
	if !isValid {
		var err error
		t.txResult, err = t.makeFailedTxResult(opFrames)
		if err != nil {
			t.log.Error("Failed to makeFailedTxResult")
			return false, err
		}
		t.rollbackCachedData(manager, opFrames)
		return false, nil
	}
	return isValid, nil
}

func (t *TransactionFrame) rollbackCachedData(manager *Manager, opFrames []OperationFrame) {
	for i, op := range opFrames {
		err := op.RollbackCachedData(manager)
		if err != nil {
			t.log.WithField("operation_i", i).WithError(err).Error("Failed to rollback cached data")
		}
	}
}

func (t *TransactionFrame) makeFailedTxResult(opFrames []OperationFrame) (*results.RestrictedTransactionError, error) {
	operationResults := make([]results.OperationResult, len(opFrames))
	for i := range opFrames {
		operationResults[i] = opFrames[i].GetResult()
	}
	return results.NewRestrictedTransactionErrorOp(xdr.TransactionResultCodeTxFailed, operationResults)
}

// returns nil, if tx is successful
func (t *TransactionFrame) GetResult() *results.RestrictedTransactionError {
	return t.txResult
}
