package resource

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/httpx"
	"github.com/openbankit/horizon/render/hal"
	"fmt"
	"golang.org/x/net/context"
)

// AccountTraits shows if account's incoming, outgoing payments are blocked
type AccountTraits struct {
	Links struct {
		Self    hal.Link `json:"self"`
		Account hal.Link `json:"account"`
	} `json:"_links"`
	PT        string `json:"paging_token"`
	AccountID string `json:"account_id"`
	BlockIn   bool   `json:"block_incoming_payments"`
	BlockOut  bool   `json:"block_outcoming_payments"`
}

func (at *AccountTraits) Populate(ctx context.Context, hat history.Account) (err error) {
	at.AccountID = hat.Address
	at.PT = hat.PagingToken()
	at.BlockIn = hat.BlockIncomingPayments
	at.BlockOut = hat.BlockOutcomingPayments
	lb := hal.LinkBuilder{httpx.BaseURL(ctx)}
	at.Links.Account = lb.Link(fmt.Sprintf("/accounts/%s", hat.Address))
	at.Links.Self = lb.Link(fmt.Sprintf("/accounts/%s/traits", hat.Address))
	return
}

func (at AccountTraits) PagingToken() string {
	return at.PT
}
