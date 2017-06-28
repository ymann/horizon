package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/validators"
)

type ChangeTrustOpFrame struct {
	*OperationFrame
	operation xdr.ChangeTrustOp
}

func NewChangeTrustOpFrame(opFrame *OperationFrame) *ChangeTrustOpFrame {
	return &ChangeTrustOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustChangeTrustOp(),
	}
}

func (frame *ChangeTrustOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	isValid, err := validators.NewAssetsValidator(manager.HistoryQ).IsAssetValid(frame.operation.Line)
	if err != nil {
		return false, err
	}

	if !isValid {
		frame.getInnerResult().Code = xdr.ChangeTrustResultCodeChangeTrustMalformed
		frame.Result.Info = results.AdditionalErrorInfoError(ASSET_NOT_ALLOWED)
		return false, nil
	}
	frame.getInnerResult().Code = xdr.ChangeTrustResultCodeChangeTrustSuccess
	return true, nil
}

func (frame *ChangeTrustOpFrame) getInnerResult() *xdr.ChangeTrustResult {
	if frame.Result.Result.Tr.ChangeTrustResult == nil {
		frame.Result.Result.Tr.ChangeTrustResult = &xdr.ChangeTrustResult{}
	}
	return frame.Result.Result.Tr.ChangeTrustResult
}
