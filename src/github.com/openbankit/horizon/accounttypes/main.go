package accounttype

import "github.com/openbankit/go-base/xdr"

func GetAll() []xdr.AccountType {
	return []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountDistributionAgent,
		xdr.AccountTypeAccountSettlementAgent,
		xdr.AccountTypeAccountExchangeAgent,
		xdr.AccountTypeAccountBank,
		xdr.AccountTypeAccountScratchCard,
	}
}
