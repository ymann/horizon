package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/txsub/results"
)

type ManageOfferOpFrame struct {
	*OperationFrame
	manageOffer xdr.ManageOfferOp
}

func NewManageOfferOpFrame(opFrame *OperationFrame) *ManageOfferOpFrame {
	return &ManageOfferOpFrame{
		OperationFrame: opFrame,
		manageOffer:    opFrame.Op.Body.MustManageOfferOp(),
	}
}

func (frame *ManageOfferOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	frame.getInnerResult().Code = xdr.ManageOfferResultCodeManageOfferMalformed
	frame.Result.Info = results.AdditionalErrorInfoError(OPERATION_NOT_ALLOWED)
	return false, nil
}

func (frame *ManageOfferOpFrame) getInnerResult() *xdr.ManageOfferResult {
	if frame.Result.Result.Tr.ManageOfferResult == nil {
		frame.Result.Result.Tr.ManageOfferResult = &xdr.ManageOfferResult{}
	}
	return frame.Result.Result.Tr.ManageOfferResult
}
