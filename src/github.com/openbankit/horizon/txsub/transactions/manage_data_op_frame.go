package transactions

import (
	"github.com/openbankit/go-base/xdr"
)

type ManageDataOpFrame struct {
	*OperationFrame
	operation xdr.ManageDataOp
}

func NewManageDataOpFrame(opFrame *OperationFrame) *ManageDataOpFrame {
	return &ManageDataOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustManageDataOp(),
	}
}

func (frame *ManageDataOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	frame.getInnerResult().Code = xdr.ManageDataResultCodeManageDataSuccess
	return true, nil
}

func (frame *ManageDataOpFrame) getInnerResult() *xdr.ManageDataResult {
	if frame.Result.Result.Tr.ManageDataResult == nil {
		frame.Result.Result.Tr.ManageDataResult = &xdr.ManageDataResult{}
	}
	return frame.Result.Result.Tr.ManageDataResult
}
