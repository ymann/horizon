package helpers

import (
	"github.com/openbankit/go-base/amount"
	"github.com/openbankit/go-base/strkey"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/assets"
	"errors"
	"strconv"
	"time"
)

type ParserInterface interface {
	GetString(name string) string
	SetInvalidField(name string, reason error)
	HasError() bool
}

func GetOptionalBool(base ParserInterface, name string) *bool {
	if base.HasError() {
		return nil
	}

	asStr := base.GetString(name)
	if asStr == "" {
		return nil
	}

	result, err := strconv.ParseBool(asStr)
	if err != nil {
		base.SetInvalidField(name, err)
		return nil
	}

	return &result
}

func GetBool(base ParserInterface, name string) bool {
	result := GetOptionalBool(base, name)
	if result != nil {
		return *result
	}
	return false
}

// GetInt64 retrieves an int64 from the action parameter of the given name.
// Populates err if the value is not a valid int64
func GetInt64(base ParserInterface, name string) int64 {
	if base.HasError() {
		return 0
	}

	asStr := base.GetString(name)
	if asStr == "" {
		return 0
	}

	asI64, err := strconv.ParseInt(asStr, 10, 64)
	if err != nil {
		base.SetInvalidField(name, err)
		return 0
	}

	return asI64
}

func GetInt32(base ParserInterface, name string) int32 {
	result := GetInt32Pointer(base, name)
	if result == nil {
		return 0
	}
	return *result
}

// GetInt32 retrieves an int32 from the action parameter of the given name.
// Populates err if the value is not a valid int32
func GetInt32Pointer(base ParserInterface, name string) *int32 {
	if base.HasError() {
		return nil
	}

	asStr := base.GetString(name)

	if asStr == "" {
		return nil
	}

	asI64, err := strconv.ParseInt(asStr, 10, 32)

	if err != nil {
		base.SetInvalidField(name, err)
		return nil
	}

	result := int32(asI64)
	return &result
}

// GetAddress retrieves a stellar address.  It confirms the value loaded is a
// valid stellar address, setting an invalid field error if it is not.
func GetAddress(base ParserInterface, name string) (result string) {
	if base.HasError() {
		return
	}

	result = GetOptionalAddress(base, name)
	if result == "" {
		base.SetInvalidField(name, errors.New("Can't be empty"))
		return ""
	}

	return result
}

func GetOptionalTime(base ParserInterface, name string) *time.Time {
	if base.HasError() {
		return nil
	}

	str := base.GetString(name)
	if str == "" {
		return nil
	}

	result := new(time.Time)
	var err error
	*result, err = time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		base.SetInvalidField(name, err)
		return nil
	}

	return result
}

func GetOptionalAddress(base ParserInterface, name string) (result string) {
	if base.HasError() {
		return
	}

	result = base.GetString(name)
	if result == "" {
		return ""
	}

	_, err := strkey.Decode(strkey.VersionByteAccountID, result)

	if err != nil {
		base.SetInvalidField(name, err)
	}

	return result
}

func GetAccountID(base ParserInterface, name string) (result xdr.AccountId) {
	if base.HasError() {
		return
	}

	accountId := GetOptionalAccountID(base, name)
	if base.HasError() {
		return
	}

	if accountId == nil {
		base.SetInvalidField(name, errors.New("can not be empty"))
		return
	}
	result = *accountId
	return
}

func GetOptionalAccountType(base ParserInterface, name string) (result *xdr.AccountType) {
	if base.HasError() {
		return
	}
	rawType := GetInt32Pointer(base, name)
	if rawType == nil {
		return nil
	}

	if !xdr.AccountTypeAccountAnonymousUser.ValidEnum(*rawType) {
		base.SetInvalidField(name, errors.New("invalid value for account type"))
		return
	}
	accountType := xdr.AccountType(*rawType)
	return &accountType
}

// GetAccountID retireves an xdr.AccountID by attempting to decode a stellar
// address at the provided name.
func GetOptionalAccountID(base ParserInterface, name string) *xdr.AccountId {
	if base.HasError() {
		return nil
	}

	strData := base.GetString(name)
	if strData == "" {
		return nil
	}

	result, err := ParseAccountId(strData)
	if err != nil {
		base.SetInvalidField(name, err)
		return nil
	}

	return &result
}

func GetPositiveAmount(base ParserInterface, name string) (result xdr.Int64) {
	if base.HasError() {
		return 0
	}
	result = GetAmount(base, name)
	if base.HasError() {
		return 0
	}

	if result <= 0 {
		base.SetInvalidField(name, errors.New("must be positive"))
	}
	return
}

// GetAmount returns a native amount (i.e. 64-bit integer) by parsing
// the string at the provided name in accordance with the stellar client
// conventions
func GetAmount(base ParserInterface, name string) (result xdr.Int64) {
	if base.HasError() {
		return 0
	}
	strAmount := base.GetString(name)
	result, err := amount.Parse(strAmount)

	if err != nil {
		base.SetInvalidField(name, err)
		return
	}

	return
}

func GetOptionalAmount(base ParserInterface, name string) (result xdr.Int64) {
	if base.HasError() {
		return 0
	}
	strAmount := base.GetString(name)
	if strAmount == "" {
		return 0
	}

	result, err := amount.Parse(strAmount)

	if err != nil {
		base.SetInvalidField(name, err)
		return
	}

	return
}

// GetAssetType is a helper that returns a xdr.AssetType by reading a string
func GetAssetType(base ParserInterface, name string) xdr.AssetType {
	if base.HasError() {
		return xdr.AssetTypeAssetTypeNative
	}

	r, err := assets.Parse(base.GetString(name))

	if base.HasError() {
		return xdr.AssetTypeAssetTypeNative
	}

	if err != nil {
		base.SetInvalidField(name, err)
	}

	return r
}

// GetAsset decodes an asset from the request fields prefixed by `prefix`.  To
// succeed, three prefixed fields must be present: asset_type, asset_code, and
// asset_issuer.
func GetAsset(base ParserInterface, prefix string) (result xdr.Asset) {
	if base.HasError() {
		return
	}
	var value interface{}

	t := GetAssetType(base, prefix+"asset_type")

	switch t {
	case xdr.AssetTypeAssetTypeCreditAlphanum4:
		a := xdr.AssetAlphaNum4{}
		a.Issuer = GetAccountID(base, prefix+"asset_issuer")

		c := base.GetString(prefix + "asset_code")
		if len(c) > len(a.AssetCode) {
			base.SetInvalidField(prefix+"asset_code", nil)
			return
		}

		copy(a.AssetCode[:len(c)], []byte(c))
		value = a
	case xdr.AssetTypeAssetTypeCreditAlphanum12:
		a := xdr.AssetAlphaNum12{}
		a.Issuer = GetAccountID(base, prefix+"asset_issuer")

		c := base.GetString(prefix + "asset_code")
		if len(c) > len(a.AssetCode) {
			base.SetInvalidField(prefix+"asset_code", nil)
			return
		}

		copy(a.AssetCode[:len(c)], []byte(c))
		value = a
	}

	result, err := xdr.NewAsset(t, value)
	if err != nil {
		panic(err)
	}
	return
}

func GetOptionalAsset(base ParserInterface, prefix string) (result *xdr.Asset) {
	if base.GetString(prefix+"asset_type") == "" {
		return nil
	}
	asset := GetAsset(base, prefix)
	if base.HasError() {
		return nil
	}
	return &asset
}

func GetOptionalRawAccountType(base ParserInterface, name string) *int32 {
	if base.HasError() {
		return nil
	}
	accountType := GetOptionalAccountType(base, name)
	if accountType == nil {
		return nil
	}
	rawAccountType := int32(*accountType)
	return &rawAccountType
}
