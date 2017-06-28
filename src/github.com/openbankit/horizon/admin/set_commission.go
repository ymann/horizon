package admin

import (
	"github.com/openbankit/horizon/assets"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"errors"
)

type SetCommissionAction struct {
	AdminAction
	CommissionKey history.CommissionKey
	FlatFee       int64
	PercentFee    int64
	Delete        bool
	commission    *history.Commission
	isNew bool
}

func NewSetCommissionAction(adminAction AdminAction) *SetCommissionAction {
	return &SetCommissionAction{
		AdminAction: adminAction,
	}
}

func (action *SetCommissionAction) Validate() {
	action.loadParams()
	if action.HasError() {
		return
	}

	var err error
	action.commission, err = history.NewCommission(action.CommissionKey, action.FlatFee, action.PercentFee)
	if err != nil {
		action.Log.WithStack(err).WithError(err).Error("Failed to create new commission")
		action.Err = errors.New("invalid commission_key")
		return
	}

	stored, err := action.HistoryQ().CommissionByHash(action.commission.KeyHash)
	if err != nil {
		action.Log.WithStack(err).WithError(err).Error("Failed to get commission by id")
		action.Err = &problem.ServerError
		return
	}

	if stored == nil {
		action.isNew = true
		if action.Delete {
			action.Err = &problem.NotFound
			return
		}
	}
}

func (action *SetCommissionAction) Apply() {
	if action.Err != nil {
		return
	}
	action.Log.WithField("commission", action.commission).Debug("Updating commission")
	var err error

	if action.isNew {
		action.Log.WithField("commission", action.commission).Debug("Trying to insert commission")
		err = action.HistoryQ().InsertCommission(action.commission)
		if err != nil {
			action.Log.WithField("commission", action.commission).WithError(err).Error("Failed to insert new commission")
			action.Err = &problem.ServerError
		}
		return
	}

	var updated bool
	if action.Delete {
		updated, err = action.HistoryQ().DeleteCommission(action.commission.KeyHash)
	} else {
		action.Log.WithField("commission", action.commission).Debug("Trying to update commission")
		updated, err = action.HistoryQ().UpdateCommission(action.commission)
	}

	if err != nil {
		action.Log.WithField("commission", action.commission).WithField("delete", action.Delete).WithError(err).Error("Failed to update/delete commission")
		action.Err = &problem.ServerError
		return
	}

	if !updated {
		action.Err = &problem.NotFound
	}
}

func (action *SetCommissionAction) loadParams() {
	action.CommissionKey.From = action.GetOptionalAddress("from")
	action.CommissionKey.To = action.GetOptionalAddress("to")
	action.CommissionKey.FromType = action.GetOptionalRawAccountType("from_type")
	action.CommissionKey.ToType = action.GetOptionalRawAccountType("to_type")
	xdrAsset := action.GetOptionalAsset("")
	if xdrAsset != nil {
		action.CommissionKey.Asset = assets.ToBaseAsset(*xdrAsset)
	}

	action.FlatFee = action.GetInt64("flat_fee")
	if action.FlatFee < 0 {
		action.SetInvalidField("flat_fee", errors.New("flat_fee can not be negative"))
		return
	}
	action.PercentFee = action.GetInt64("percent_fee")
	if action.PercentFee < 0 {
		action.SetInvalidField("percent_fee", errors.New("percent_fee can not be negative"))
		return
	}
	action.Delete = action.GetBool("delete")
}
