package transactions

import (
	"github.com/openbankit/go-base/xdr"
)

type AccountMergeOpFrame struct {
	*OperationFrame
	operation xdr.AccountId
}

func NewAccountMergeOpFrame(opFrame *OperationFrame) *AccountMergeOpFrame {
	return &AccountMergeOpFrame{
		OperationFrame: opFrame,
		operation: opFrame.Op.Body.MustDestination(),
	}
}

func (frame *AccountMergeOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	frame.getInnerResult().Code = xdr.AccountMergeResultCodeAccountMergeSuccess
	return true, nil
}

func (frame *AccountMergeOpFrame) getInnerResult() *xdr.AccountMergeResult {
	if frame.Result.Result.Tr.AccountMergeResult == nil {
		frame.Result.Result.Tr.AccountMergeResult = &xdr.AccountMergeResult{}
	}
	return frame.Result.Result.Tr.AccountMergeResult
}
