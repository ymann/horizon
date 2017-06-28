package admin

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/helpers"
	"github.com/openbankit/horizon/log"
	"github.com/spf13/cast"
)

type AdminActionInterface interface {
	// returns error, If any occurred
	GetError() error
	// validates actions. If action is not valid. Sets error
	Validate()
	// applies action
	Apply()
}

type AdminAction struct {
	Err     error
	rawData map[string]interface{}
	Log     *log.Entry

	hq history.QInterface
}

func (action *AdminAction) HistoryQ() history.QInterface {
	return action.hq
}

func (action *AdminAction) GetError() error {
	return action.Err
}

func NewAdminAction(data map[string]interface{}, hq history.QInterface) AdminAction {
	return AdminAction{
		rawData: data,
		Log:     log.WithField("service", "admin_action"),
		hq:      hq,
	}
}

func (p *AdminAction) SetInvalidField(name string, reason error) {
	p.Err = InvalidField(name, reason)
}

func (p *AdminAction) GetString(name string) string {
	if p.Err != nil {
		return ""
	}
	value, ok := p.rawData[name]
	if !ok {
		return ""
	}
	strValue, err := cast.ToStringE(value)
	if err != nil {
		p.SetInvalidField(name, err)
		return ""
	}
	return strValue
}

func (p *AdminAction) HasError() bool {
	return p.Err != nil
}

func (p *AdminAction) GetOptionalAddress(name string) string {
	return helpers.GetOptionalAddress(p, name)
}

func (p *AdminAction) GetOptionalRawAccountType(name string) *int32 {
	return helpers.GetOptionalRawAccountType(p, name)
}

func (p *AdminAction) GetAsset(prefix string) xdr.Asset {
	return helpers.GetAsset(p, prefix)
}

func (p *AdminAction) GetInt64(name string) int64 {
	return helpers.GetInt64(p, name)
}

func (p *AdminAction) GetBool(name string) bool {
	return helpers.GetBool(p, name)
}

func (p *AdminAction) GetAddress(name string) string {
	return helpers.GetAddress(p, name)
}

func (p *AdminAction) GetOptionalBool(name string) *bool {
	return helpers.GetOptionalBool(p, name)
}

func (p *AdminAction) GetOptionalAsset(prefix string) *xdr.Asset {
	return helpers.GetOptionalAsset(p, prefix)
}

func (p *AdminAction) GetOptionalAmount(name string) int64 {
	return int64(helpers.GetOptionalAmount(p, name))
}
