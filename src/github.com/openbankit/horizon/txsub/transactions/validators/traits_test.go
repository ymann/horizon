package validators

import (
	"testing"

	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/results"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
)

func TestTraits(t *testing.T) {
	traits := NewTraitsValidator()
	Convey("Traits test:", t, func() {
		sourceKP, err := keypair.Random()
		So(err, ShouldBeNil)
		source := &history.Account{
			TotalOrderID: history.TotalOrderID{
				ID: rand.Int63(),
			},
			Address: sourceKP.Address(),
		}
		destKP, err := keypair.Random()
		So(err, ShouldBeNil)
		dest := &history.Account{
			TotalOrderID: history.TotalOrderID{
				ID: rand.Int63(),
			},
			Address: destKP.Address(),
		}
		Convey("Both accounts does not have traits", func() {
			result, err := traits.CheckTraits(source, dest)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})
		Convey("Both accounts have traits, but not blocked", func() {
			source.BlockIncomingPayments = true
			dest.BlockOutcomingPayments = true
			result, err := traits.CheckTraits(source, dest)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})
		Convey("Source is blocked", func() {
			source.BlockOutcomingPayments = true
			result, err := traits.CheckTraits(source, dest)
			So(err, ShouldBeNil)
			assert.Equal(t, result, &results.RestrictedForAccountError{
				Reason: fmt.Sprintf("Outcoming payments for account (%s) are restricted by administrator.", source.Address),
			})
		})
		Convey("Dest is blocked", func() {
			source.BlockOutcomingPayments = false
			dest.BlockIncomingPayments = true
			result, err := traits.CheckTraits(source, dest)
			So(err, ShouldBeNil)
			assert.Equal(t, result, &results.RestrictedForAccountError{
				Reason: fmt.Sprintf("Incoming payments for account (%s) are restricted by administrator.", dest.Address),
			})
		})

	})
}
