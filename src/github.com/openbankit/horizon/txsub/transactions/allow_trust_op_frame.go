package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/results"
	"github.com/openbankit/horizon/txsub/transactions/validators"
)

type AllowTrustOpFrame struct {
	*OperationFrame
	operation xdr.AllowTrustOp
}

func NewAllowTrustOpFrame(opFrame *OperationFrame) *AllowTrustOpFrame {
	return &AllowTrustOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustAllowTrustOp(),
	}
}

func (frame *AllowTrustOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	isValid, err := frame.isAssetValid(manager.HistoryQ)
	if err != nil {
		return false, err
	}

	if !isValid {
		frame.getInnerResult().Code = xdr.AllowTrustResultCodeAllowTrustMalformed
		frame.Result.Info = results.AdditionalErrorInfoError(ASSET_NOT_ALLOWED)
		return false, nil
	}
	frame.getInnerResult().Code = xdr.AllowTrustResultCodeAllowTrustSuccess
	return true, nil
}

func (frame *AllowTrustOpFrame) isAssetValid(historyQ history.QInterface) (bool, error) {
	xdrAsset := xdr.Asset{
		Type: frame.operation.Asset.Type,
	}
	issuer := frame.ParentTxFrame.Tx.Tx.SourceAccount
	if frame.Op.SourceAccount != nil {
		issuer = *frame.Op.SourceAccount
	}
	switch frame.operation.Asset.Type {
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		if frame.operation.Asset.AssetCode4 == nil {
			return false, nil
		}
		xdrAsset.AlphaNum4 = &xdr.AssetAlphaNum4{
			AssetCode: *frame.operation.Asset.AssetCode4,
			Issuer:    issuer,
		}
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		if frame.operation.Asset.AssetCode12 == nil {
			return false, nil
		}
		xdrAsset.AlphaNum12 = &xdr.AssetAlphaNum12{
			AssetCode: *frame.operation.Asset.AssetCode12,
			Issuer:    issuer,
		}
	default:
		return false, nil
	}
	return validators.NewAssetsValidator(historyQ).IsAssetValid(xdrAsset)
}

func (frame *AllowTrustOpFrame) getInnerResult() *xdr.AllowTrustResult {
	if frame.Result.Result.Tr.AllowTrustResult == nil {
		frame.Result.Result.Tr.AllowTrustResult = &xdr.AllowTrustResult{}
	}
	return frame.Result.Result.Tr.AllowTrustResult
}
