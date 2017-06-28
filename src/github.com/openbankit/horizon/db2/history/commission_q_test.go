package history

import (
	"testing"

	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
)

func TestCommissionQ(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()

	account, err := keypair.Random()
	assert.Nil(t, err)

	h := Q{tt.HorizonRepo()}
	Convey("Select by hash:", t, func() {
		commissionKey := CommissionKey{
			From: account.Address(),
		}
		commission, err := NewCommission(commissionKey, rand.Int63(), rand.Int63())
		So(err, ShouldBeNil)
		Convey("By hash returns nil - if no commission", func() {
			stored, err := h.CommissionByHash(commission.KeyHash)
			So(err, ShouldBeNil)
			So(stored, ShouldBeNil)
		})
		err = h.InsertCommission(commission)
		So(err, ShouldBeNil)
		Convey("Returns correct value", func() {
			stored, err := h.CommissionByHash(commission.KeyHash)
			So(err, ShouldBeNil)
			So(stored.FlatFee, ShouldEqual, commission.FlatFee)
			So(stored.PercentFee, ShouldEqual, commission.PercentFee)
			So(stored.KeyHash, ShouldEqual, commission.KeyHash)
		})

		isDeleted, err := h.DeleteCommission(commission.KeyHash)
		So(err, ShouldBeNil)
		So(isDeleted, ShouldBeTrue)
		Convey("Return nil after deleted", func() {
			stored, err := h.CommissionByHash(commission.KeyHash)
			So(err, ShouldBeNil)
			So(stored, ShouldBeNil)
		})
	})
}
