package transactions

import (
	"github.com/openbankit/go-base/xdr"
)

type InflationOpFrame struct {
	*OperationFrame
}

func NewInflationOpFrame(opFrame *OperationFrame) *InflationOpFrame {
	return &InflationOpFrame{
		OperationFrame: opFrame,
	}
}

func (frame *InflationOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	frame.getInnerResult().Code = xdr.InflationResultCodeInflationSuccess
	return true, nil
}

func (frame *InflationOpFrame) getInnerResult() *xdr.InflationResult {
	if frame.Result.Result.Tr.InflationResult == nil {
		frame.Result.Result.Tr.InflationResult = &xdr.InflationResult{}
	}
	return frame.Result.Result.Tr.InflationResult
}
