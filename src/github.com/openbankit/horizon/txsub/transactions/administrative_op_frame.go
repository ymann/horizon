package transactions

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/admin"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"github.com/openbankit/horizon/txsub/results"
	"encoding/json"
)

type AdministrativeOpFrame struct {
	*OperationFrame
	operation           xdr.AdministrativeOp
	adminActionProvider admin.AdminActionProviderInterface
}

func NewAdministrativeOpFrame(opFrame *OperationFrame) *AdministrativeOpFrame {
	return &AdministrativeOpFrame{
		OperationFrame: opFrame,
		operation:      opFrame.Op.Body.MustAdminOp(),
	}
}

func (frame *AdministrativeOpFrame) getAdminActionProvider(historyQ history.QInterface) admin.AdminActionProviderInterface {
	if frame.adminActionProvider == nil {
		frame.adminActionProvider = admin.NewAdminActionProvider(historyQ)
	}
	return frame.adminActionProvider
}

func (frame *AdministrativeOpFrame) DoCheckValid(manager *Manager) (bool, error) {
	var opData map[string]interface{}
	err := json.Unmarshal([]byte(frame.operation.OpData), &opData)
	if err != nil {
		frame.getInnerResult().Code = xdr.AdministrativeResultCodeAdministrativeMalformed
		frame.Result.Info = results.AdditionalErrorInfoError(err)
		return false, nil
	}

	adminAction, err := frame.getAdminActionProvider(manager.HistoryQ).CreateNewParser(opData)
	if err != nil {
		frame.getInnerResult().Code = xdr.AdministrativeResultCodeAdministrativeMalformed
		frame.Result.Info = results.AdditionalErrorInfoError(err)
		return false, nil
	}

	adminAction.Validate()
	err = adminAction.GetError()
	if err != nil {
		switch err.(type) {
		case *admin.InvalidFieldError:
			frame.getInnerResult().Code = xdr.AdministrativeResultCodeAdministrativeMalformed
			invalidField := err.(*admin.InvalidFieldError)
			frame.Result.Info = results.AdditionalErrorInfoInvField(*invalidField)
			return false, nil
		case *problem.P:
			prob := err.(*problem.P)
			if prob.Type == problem.ServerError.Type {
				return false, err
			}
			frame.getInnerResult().Code = xdr.AdministrativeResultCodeAdministrativeMalformed
			frame.Result.Info = results.AdditionalErrorInfoError(err)
			return false, nil
		default:
			return false, err
		}
	}
	frame.getInnerResult().Code = xdr.AdministrativeResultCodeAdministrativeSuccess
	return true, nil
}

func (frame *AdministrativeOpFrame) getInnerResult() *xdr.AdministrativeResult {
	if frame.Result.Result.Tr.AdminResult == nil {
		frame.Result.Result.Tr.AdminResult = &xdr.AdministrativeResult{}
	}
	return frame.Result.Result.Tr.AdminResult
}
