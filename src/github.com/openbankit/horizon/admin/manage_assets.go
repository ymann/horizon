package admin

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/render/problem"
	"database/sql"
)

type ManageAssetsAction struct {
	AdminAction
	asset       xdr.Asset
	isNew       bool
	delete      bool
	isAnonymous bool
	storedAsset history.Asset
}

func NewManageAssetsAction(adminAction AdminAction) *ManageAssetsAction {
	return &ManageAssetsAction{
		AdminAction: adminAction,
	}
}

func (action *ManageAssetsAction) Validate() {
	action.loadParams()
	if action.Err != nil {
		return
	}

	err := action.HistoryQ().Asset(&action.storedAsset, action.asset)
	if err != nil {
		if err != sql.ErrNoRows {
			action.Log.WithStack(err).WithError(err).Error("Failed to get asset from db")
			action.Err = &problem.ServerError
			return
		}
		action.isNew = true
	}

	if action.isNew && action.delete {
		action.Err = &problem.NotFound
		return
	}

	var code, issuer string
	var assetType xdr.AssetType
	err = action.asset.Extract(&assetType, &code, &issuer)
	if err != nil {
		action.Log.WithError(err).Error("Failed to extract asset data")
		action.Err = &problem.ServerError
		return
	}
	action.storedAsset.Type = int(assetType)
	action.storedAsset.Code = code
	action.storedAsset.Issuer = issuer
	action.storedAsset.IsAnonymous = action.isAnonymous
}

func (action *ManageAssetsAction) Apply() {
	if action.Err != nil {
		return
	}

	if action.delete {
		_, action.Err = action.HistoryQ().DeleteAsset(action.storedAsset.Id)
		return
	}

	if action.isNew {
		action.Log.WithField("Asset", action.storedAsset).Warn("Inserting asset!")
		action.Err = action.HistoryQ().InsertAsset(&action.storedAsset)
		return
	}
	_, action.Err = action.HistoryQ().UpdateAsset(&action.storedAsset)
}

func (action *ManageAssetsAction) loadParams() {
	action.asset = action.GetAsset("")
	action.delete = action.GetBool("delete")
	action.isAnonymous = action.GetBool("is_anonymous")
}
