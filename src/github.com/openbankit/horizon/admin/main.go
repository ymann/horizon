package admin

import "fmt"

type AdminActionSubject string

const (
	SubjectCommission                 AdminActionSubject = "commission"
	SubjectTraits                     AdminActionSubject = "traits"
	SubjectAccountLimits              AdminActionSubject = "account_limits"
	SubjectAsset                      AdminActionSubject = "asset"
	SubjectMaxPaymentReversalDuration AdminActionSubject = "max_reversal_duration"
)

type InvalidFieldError struct {
	FieldName string
	Reason    error
}

func InvalidField(name string, reason error) *InvalidFieldError {
	return &InvalidFieldError{
		FieldName: name,
		Reason:    reason,
	}
}

func (err *InvalidFieldError) Error() string {
	return fmt.Sprintf("invalid_field: %s; reason: %s", err.FieldName, err.Reason.Error())
}
