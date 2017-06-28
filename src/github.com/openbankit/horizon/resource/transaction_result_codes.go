package resource

import (
	"github.com/openbankit/horizon/txsub/results"
	"golang.org/x/net/context"
)

// Populate fills out the details
func (res *TransactionResultCodes) Populate(ctx context.Context,
	fail *results.FailedTransactionError,
) (err error) {

	res.TransactionCode, err = fail.TransactionResultCode()
	if err != nil {
		return
	}

	res.OperationCodes, err = fail.OperationResultCodes()
	if err != nil {
		return
	}

	return
}
