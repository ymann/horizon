package helpers

import (
	"github.com/openbankit/go-base/xdr"
)

// bankAgent returns true if specified user type is a bank agent
func IsBankOrAgent(accountType xdr.AccountType) bool {
	switch accountType {
	case xdr.AccountTypeAccountDistributionAgent, xdr.AccountTypeAccountSettlementAgent, xdr.AccountTypeAccountExchangeAgent, xdr.AccountTypeAccountBank:
		return true
	}
	return false
}

func IsUser(accountType xdr.AccountType) bool {
	switch accountType {
	case xdr.AccountTypeAccountAnonymousUser, xdr.AccountTypeAccountRegisteredUser:
		return true
	}
	return false
}
