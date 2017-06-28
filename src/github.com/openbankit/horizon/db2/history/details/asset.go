package details

type Asset struct {
	Type   string `json:"asset_type,omitempty"`
	Code   string `json:"asset_code,omitempty"`
	Issuer string `json:"asset_issuer,omitempty"`
}
