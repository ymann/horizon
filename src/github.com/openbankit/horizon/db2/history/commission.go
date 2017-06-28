package history

import (
	"github.com/openbankit/horizon/log"
	"encoding/json"
	"github.com/go-errors/errors"
)

func NewCommission(key CommissionKey, flatFee, percentFee int64) (*Commission, error) {
	hashData, hash, err := key.HashData()
	if err != nil {
		log.WithStack(err).Error("Failed to get hash for commission key: " + err.Error())
		return nil, err
	}
	return &Commission{
		KeyHash:    hash,
		KeyValue:   hashData,
		FlatFee:    flatFee,
		PercentFee: percentFee,
	}, nil
}

func (c Commission) Equals(o Commission) bool {
	if c.KeyHash != o.KeyHash || c.FlatFee != o.FlatFee || c.PercentFee != o.PercentFee {
		return false
	}
	cKey := c.GetKey()
	return cKey.Equals(o.GetKey())
}

// UnmarshalDetails unmarshals the details of this effect into `dest`
func (r *Commission) UnmarshalKeyDetails(dest interface{}) error {

	err := json.Unmarshal([]byte(r.KeyValue), &dest)
	if err != nil {
		err = errors.Wrap(err, 1)
	}
	return err
}
