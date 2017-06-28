package validators

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/txsub/results"
	"fmt"
)

type AccountTypeValidatorInterface interface {
	VerifyAccountTypesForPayment(from, to xdr.AccountType) *results.RestrictedForAccountTypeError
}

type AccountTypeValidator struct {
}

func NewAccountTypeValidator() *AccountTypeValidator {
	return &AccountTypeValidator{}
}

// VerifyAccountTypesForPayment performs account types check for payment operation
func (v *AccountTypeValidator) VerifyAccountTypesForPayment(from, to xdr.AccountType) *results.RestrictedForAccountTypeError {
	if !contains(typeRestrictions[from], to) {
		return &results.RestrictedForAccountTypeError{
			Reason: fmt.Sprintf("Payments from %s to %s are restricted.", from.String(), to.String()),
		}
	}

	return nil
}

func contains(list []xdr.AccountType, a xdr.AccountType) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// TODO: generate from template?
// TODO: use sets instead of arrays
// typeRestrictions describes who can send payments to whom
var typeRestrictions = map[xdr.AccountType][]xdr.AccountType{

	xdr.AccountTypeAccountBank: []xdr.AccountType{
		xdr.AccountTypeAccountDistributionAgent,
		xdr.AccountTypeAccountExchangeAgent,
	},

	xdr.AccountTypeAccountCommission: []xdr.AccountType{
		xdr.AccountTypeAccountDistributionAgent,
	},

	xdr.AccountTypeAccountDistributionAgent: []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountSettlementAgent,
		xdr.AccountTypeAccountScratchCard,
	},

	xdr.AccountTypeAccountSettlementAgent: []xdr.AccountType{
		xdr.AccountTypeAccountBank,
	},

	xdr.AccountTypeAccountExchangeAgent: []xdr.AccountType{
		xdr.AccountTypeAccountBank,
		xdr.AccountTypeAccountAnonymousUser,
	},

	xdr.AccountTypeAccountAnonymousUser: []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountSettlementAgent,
		xdr.AccountTypeAccountExchangeAgent,
	},

	xdr.AccountTypeAccountRegisteredUser: []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountSettlementAgent,
	},

	xdr.AccountTypeAccountMerchant: []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
		xdr.AccountTypeAccountMerchant,
		xdr.AccountTypeAccountSettlementAgent,
	},

	xdr.AccountTypeAccountScratchCard: []xdr.AccountType{
		xdr.AccountTypeAccountAnonymousUser,
		xdr.AccountTypeAccountRegisteredUser,
	},
}
