package transactions

import (
	"github.com/openbankit/go-base/xdr"
)

type SetOptionsOpFrame struct {
	*OperationFrame
	operation xdr.SetOptionsOp
}

func NewSetOptionsOpFrame(opFrame *OperationFrame) *SetOptionsOpFrame {
	return &SetOptionsOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustSetOptionsOp(),
	}
}

func (p *SetOptionsOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	p.getInnerResult().Code = xdr.SetOptionsResultCodeSetOptionsSuccess
	return true, nil
}

func (p *SetOptionsOpFrame) getInnerResult() *xdr.SetOptionsResult {
	if p.Result.Result.Tr.SetOptionsResult == nil {
		p.Result.Result.Tr.SetOptionsResult = &xdr.SetOptionsResult{}
	}
	return p.Result.Result.Tr.SetOptionsResult
}
