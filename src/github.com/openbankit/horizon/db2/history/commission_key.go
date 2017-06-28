package history

import (
	"github.com/openbankit/go-base/hash"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"encoding/hex"
	"encoding/json"
)

type CommissionKey struct {
	details.Asset
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	FromType *int32 `json:"from_type,omitempty"`
	ToType   *int32 `json:"to_type,omitempty"`
	hash     string
}

func (k *CommissionKey) Equals(o CommissionKey) bool {
	if k.Asset != o.Asset || k.From != o.From || k.To != o.To {
		return false
	}
	return equals(k.FromType, o.FromType) && equals(k.ToType, o.ToType)
}

func equals(l, r *int32) bool {
	if l == nil || r == nil {
		return r == l
	}
	return *l == *r
}

func CreateCommissionKeys(from, to string, fromType, toType int32, asset details.Asset) map[string]CommissionKey {
	keys := make([]CommissionKey, 1, 32)
	defaultFee := CommissionKey{}
	keys[0] = defaultFee
	keys = set(keys, &from, nil, nil, nil, nil)
	keys = set(keys, nil, &to, nil, nil, nil)
	keys = set(keys, nil, nil, &fromType, nil, nil)
	keys = set(keys, nil, nil, nil, &toType, nil)
	keys = set(keys, nil, nil, nil, nil, &asset)
	result := make(map[string]CommissionKey)
	for _, key := range keys {
		result[key.UnsafeHash()] = key
	}
	return result
}

func set(keys []CommissionKey, from, to *string, fromType, toType *int32, asset *details.Asset) []CommissionKey {
	size := len(keys)
	var value CommissionKey
	for j := 0; j < size; j++ {
		value = keys[j]
		switch {
		case from != nil:
			value.From = *from
		case to != nil:
			value.To = *to
		case fromType != nil:
			value.FromType = fromType
		case toType != nil:
			value.ToType = toType
		case asset != nil:
			value.Asset = *asset
		}
		keys = append(keys, value)
	}
	return keys
}

func (c *Commission) GetKey() CommissionKey {
	var key CommissionKey
	c.UnmarshalKeyDetails(&key)
	return key
}

func (key *CommissionKey) Hash() (string, error) {
	_, hash, err := key.HashData()
	return hash, err
}

func (key *CommissionKey) UnsafeHash() string {
	result, _ := key.Hash()
	return result
}

func (key *CommissionKey) HashData() (hashData string, hashValue string, err error) {
	hashDataByte, err := json.Marshal(key)
	if err != nil {
		log.WithField("Error", err).Error("Failed to marshal commission key")
		return "", "", err
	}
	hashBase := hash.Hash(hashDataByte)
	hashValue = hex.EncodeToString(hashBase[:])
	hashData = string(hashDataByte)
	return
}

// returns 1 if key hash higher priority, -1 if lower, 0 if equal
func (key *CommissionKey) Compare(other *CommissionKey) int {
	if other == nil {
		return 1
	}

	keyWeight := key.CountWeight()
	otherWeight := other.CountWeight()
	log.WithField("keyWeight", keyWeight).WithField("otherWeight", otherWeight).Debug("counted weight")

	if keyWeight > otherWeight {
		return 1
	} else if keyWeight < otherWeight {
		return -1
	}
	return 0
}

const (
	assetWeight   = 1
	typeWeight    = assetWeight + 1
	accountWeight = typeWeight*2 + assetWeight + 1
)

func (key *CommissionKey) IsAssetSet() bool {
	asset := &key.Asset
	return asset.Type != "" || asset.Code != "" || asset.Issuer != ""
}

func (key *CommissionKey) CountWeight() int {
	weight := 0

	if key.IsAssetSet() {
		weight += assetWeight
	}

	if key.FromType != nil {
		weight += typeWeight
	}

	if key.ToType != nil {
		weight += typeWeight
	}

	if key.From != "" {
		weight += accountWeight
	}

	if key.To != "" {
		weight += accountWeight
	}

	return weight
}
