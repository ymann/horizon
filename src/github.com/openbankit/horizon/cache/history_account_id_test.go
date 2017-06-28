package cache

import (
	"github.com/openbankit/horizon/test"
	"testing"
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/go-base/xdr"
)

func TestHistoryAccountID(t *testing.T) {
	tt := test.Start(t).Scenario("base")
	defer tt.Finish()
	db := tt.HorizonRepo()
	c := NewHistoryAccount(&history.Q{
		Repo: db,
	})
	tt.Assert.Equal(0, c.Cache.ItemCount())

	address := test.NewTestConfig().BankMasterKey
	account, err := c.Get(address)
	if tt.Assert.NoError(err) {
		tt.Assert.Equal(&history.Account{
			TotalOrderID: history.TotalOrderID{
				ID: 1,
			},
			Address: address,
			AccountType: xdr.AccountTypeAccountBank,
		}, account)
		tt.Assert.Equal(1, c.Cache.ItemCount())
	}

	account, err = c.Get("NOT_REAL")
	tt.Assert.True(db.NoRows(err))
	var noAccount *history.Account
	tt.Assert.Equal(noAccount, account)
}
