package actions

import (
	"mime"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/helpers"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/render/problem"
	"time"
)

const (
	// ParamCursor is a query string param name
	ParamCursor = "cursor"
	// ParamOrder is a query string param name
	ParamOrder = "order"
	// ParamLimit is a query string param name
	ParamLimit = "limit"
)

// GetString retrieves a string from either the URLParams, form or query string.
// This method uses the priority (URLParams, Form, Query).
func (base *Base) GetString(name string) string {
	if base.Err != nil {
		return ""
	}

	fromURL, ok := base.GojiCtx.URLParams[name]

	if ok {
		return fromURL
	}

	fromForm := base.R.FormValue(name)

	if fromForm != "" {
		return fromForm
	}

	return base.R.URL.Query().Get(name)
}

// SetInvalidField establishes an error response triggered by an invalid
// input field from the user.
func (base *Base) SetInvalidField(name string, reason error) {
	log.WithField("name", name).WithError(reason).Info("Setting invalid field")
	br := problem.BadRequest

	br.Extras = map[string]interface{}{}
	br.Extras["invalid_field"] = name
	br.Extras["reason"] = reason.Error()

	base.Err = &br
}

func (base *Base) HasError() bool {
	return base.Err != nil
}

func (base *Base) GetOptionalBool(name string) *bool {
	return helpers.GetOptionalBool(base, name)
}

func (base *Base) GetBool(name string) bool {
	return helpers.GetBool(base, name)
}

// GetInt64 retrieves an int64 from the action parameter of the given name.
// Populates err if the value is not a valid int64
func (base *Base) GetInt64(name string) int64 {
	return helpers.GetInt64(base, name)
}

func (base *Base) GetInt32(name string) int32 {
	return helpers.GetInt32(base, name)
}

// GetInt32 retrieves an int32 from the action parameter of the given name.
// Populates err if the value is not a valid int32
func (base *Base) GetInt32Pointer(name string) *int32 {
	return helpers.GetInt32Pointer(base, name)
}

// GetPagingParams returns the cursor/order/limit triplet that is the
// standard way of communicating paging data to a horizon endpoint.
func (base *Base) GetPagingParams() (cursor string, order string, limit uint64) {
	if base.Err != nil {
		return
	}

	cursor = base.GetString(ParamCursor)
	order = base.GetString(ParamOrder)
	// TODO: add GetUint64 helpers
	limit = uint64(base.GetInt64(ParamLimit))

	if lei := base.R.Header.Get("Last-Event-ID"); lei != "" {
		cursor = lei
	}

	return
}

func (base *Base) GetCloseAtParams() (after, before *time.Time) {
	if base.Err != nil {
		return
	}

	after = base.GetOptionalTime("after")
	before = base.GetOptionalTime("before")
	return
}

// GetPageQuery is a helper that returns a new db.PageQuery struct initialized
// using the results from a call to GetPagingParams()
func (base *Base) GetPageQuery() db2.PageQuery {
	if base.Err != nil {
		return db2.PageQuery{}
	}

	r, err := db2.NewPageQuery(base.GetPagingParams())

	if err != nil {
		base.Err = err
	}

	return r
}

// GetAddress retrieves a stellar address.  It confirms the value loaded is a
// valid stellar address, setting an invalid field error if it is not.
func (base *Base) GetAddress(name string) (result string) {
	return helpers.GetAddress(base, name)
}

func (base *Base) GetOptionalTime(name string) (*time.Time) {
	return helpers.GetOptionalTime(base, name)
}

func (base *Base) GetOptionalAddress(name string) (result string) {
	return helpers.GetOptionalAddress(base, name)
}

func (base *Base) GetAccountID(name string) (result xdr.AccountId) {
	return helpers.GetAccountID(base, name)
}

func (base *Base) GetOptionalAccountType(name string) (result *xdr.AccountType) {
	return helpers.GetOptionalAccountType(base, name)
}

// GetAccountID retireves an xdr.AccountID by attempting to decode a stellar
// address at the provided name.
func (base *Base) GetOptionalAccountID(name string) (result *xdr.AccountId) {
	return helpers.GetOptionalAccountID(base, name)
}

func (base *Base) GetPositiveAmount(name string) (result xdr.Int64) {
	return helpers.GetPositiveAmount(base, name)
}

func (base *Base) GetOptionalRawAccountType(name string) *int32 {
	return helpers.GetOptionalRawAccountType(base, name)
}

// GetAmount returns a native amount (i.e. 64-bit integer) by parsing
// the string at the provided name in accordance with the stellar client
// conventions
func (base *Base) GetAmount(name string) (result xdr.Int64) {
	return helpers.GetAmount(base, name)
}

// GetAssetType is a helper that returns a xdr.AssetType by reading a string
func (base *Base) GetAssetType(name string) xdr.AssetType {
	return helpers.GetAssetType(base, name)
}

// GetAsset decodes an asset from the request fields prefixed by `prefix`.  To
// succeed, three prefixed fields must be present: asset_type, asset_code, and
// asset_issuer.
func (base *Base) GetAsset(prefix string) (result xdr.Asset) {
	return helpers.GetAsset(base, prefix)
}

// Path returns the current action's path, as determined by the http.Request of
// this action
func (base *Base) Path() string {
	return base.R.URL.Path
}

// ValidateBodyType sets an error on the action if the requests Content-Type
//  is not `application/x-www-form-urlencoded`
func (base *Base) ValidateBodyType() {
	c := base.R.Header.Get("Content-Type")

	if c == "" {
		return
	}

	mt, _, err := mime.ParseMediaType(c)

	if err != nil {
		base.Err = err
		return
	}

	switch {
	case mt == "application/x-www-form-urlencoded":
		return
	case mt == "multipart/form-data":
		return
	default:
		base.Err = &problem.UnsupportedMediaType
	}
}
