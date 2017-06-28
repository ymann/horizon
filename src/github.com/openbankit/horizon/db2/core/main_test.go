package core

import (
	"github.com/openbankit/horizon/test"
	"testing"
)

func _TestLatestLedger(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	q := &Q{tt.CoreRepo()}

	var seq int
	err := q.LatestLedger(&seq)

	if tt.Assert.NoError(err) {
		tt.Assert.Equal(3, seq)
	}
}
