package options

import (
	"github.com/openbankit/horizon/db2/history"
	"time"
)

type MaxReversalDuration struct {
	DurationInSeconds int64  `json:"duration_in_seconds"`
	DurationStr       string `json:"duration_str"`
}

func (m *MaxReversalDuration) Populate(maxReversalDuration history.MaxReversalDuration) {
	duration, _ := maxReversalDuration.GetMaxDuration()
	m.DurationInSeconds = int64(duration/time.Second)
	m.DurationStr = duration.String()
}
