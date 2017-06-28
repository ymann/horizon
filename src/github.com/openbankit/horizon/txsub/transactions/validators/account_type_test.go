package validators

import (
	"testing"

	"github.com/openbankit/go-base/xdr"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAccountTypes(t *testing.T) {
	Convey("VerifyAccountTypesForPayment:", t, func() {
		Convey("Bank can't send to anon user", func() {
			validator := NewAccountTypeValidator()
			err := validator.VerifyAccountTypesForPayment(xdr.AccountTypeAccountBank, xdr.AccountTypeAccountAnonymousUser)
			So(err, ShouldNotBeNil)
		})

	})
}
