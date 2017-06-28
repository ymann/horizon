package transactions

import (
	"github.com/openbankit/go-base/xdr"
)

type CreateAccountOpFrame struct {
	*OperationFrame
	operation xdr.CreateAccountOp
}

func NewCreateAccountOpFrame(opFrame *OperationFrame) *CreateAccountOpFrame {
	return &CreateAccountOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustCreateAccountOp(),
	}
}

func (frame *CreateAccountOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	frame.getInnerResult().Code = xdr.CreateAccountResultCodeCreateAccountSuccess
	return true, nil
}

func (frame *CreateAccountOpFrame) getInnerResult() *xdr.CreateAccountResult {
	if frame.Result.Result.Tr.CreateAccountResult == nil {
		frame.Result.Result.Tr.CreateAccountResult = &xdr.CreateAccountResult{}
	}
	return frame.Result.Result.Tr.CreateAccountResult
}
