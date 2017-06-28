package validators

import (
	"github.com/openbankit/horizon/db2/history"
	"github.com/openbankit/horizon/txsub/results"
	"fmt"
)

type TraitsValidatorInterface interface {
	CheckTraits(source, destination *history.Account) (*results.RestrictedForAccountError, error)
	CheckTraitsForAccount(account *history.Account, isSource bool) (*results.RestrictedForAccountError, error)
}

type TraitsValidator struct {
}

func NewTraitsValidator() *TraitsValidator {
	return &TraitsValidator{
	}
}

// VerifyRestrictions checks traits of the involved accounts
func (v *TraitsValidator) CheckTraits(source, destination *history.Account) (*results.RestrictedForAccountError, error) {
	restriction, err := v.CheckTraitsForAccount(source, true)
	// if id is zero - account is new, so there are no traits for it
	if restriction == nil && err == nil && destination.ID != 0 {
		restriction, err = v.CheckTraitsForAccount(destination, false)
	}
	return restriction, err
}

func (v *TraitsValidator) CheckTraitsForAccount(account *history.Account, isSource bool) (*results.RestrictedForAccountError, error) {
	// Check restrictions
	if isSource && account.BlockOutcomingPayments {
		return &results.RestrictedForAccountError{
			Reason: fmt.Sprintf("Outcoming payments for account (%s) are restricted by administrator.", account.Address),
		}, nil
	}

	if !isSource && account.BlockIncomingPayments {
		return &results.RestrictedForAccountError{
			Reason: fmt.Sprintf("Incoming payments for account (%s) are restricted by administrator.", account.Address),
		}, nil
	}

	return nil, nil
}
