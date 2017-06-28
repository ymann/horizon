package history

import (
	"strconv"
	"time"
)

type MaxReversalDuration Options

func NewMaxReversalDuration() *MaxReversalDuration {
	result := MaxReversalDuration(Options{
		Name: OPTIONS_MAX_REVERSAL_DURATION,
		Data: "0",
	})
	return &result
}

func (r *MaxReversalDuration) GetMaxDuration() (time.Duration, error) {
	rawDuration, err := strconv.ParseInt(r.Data, 10, 64)
	if err != nil {
		return time.Duration(0), err
	}

	return time.Duration(rawDuration) * time.Second, nil
}

func (r *MaxReversalDuration) SetMaxDuration(val time.Duration) {
	r.Data = strconv.FormatInt(int64(val/time.Second), 10)
}
