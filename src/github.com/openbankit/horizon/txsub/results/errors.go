package results

import (
	"errors"
	"fmt"

	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/admin"
	"github.com/openbankit/horizon/codes"
)

var (
	ErrNoResults = errors.New("No result found")
	ErrCanceled  = errors.New("canceled")
	ErrTimeout   = errors.New("timeout")

	// ErrBadSequence is a canned error response for transactions whose sequence
	// number is wrong.
	ErrBadSequence = &FailedTransactionError{"////+wAAAAA="}
	// ErrNoAccount is returned when the source account for the transaction
	// cannot be found in the database
	ErrNoAccount = &FailedTransactionError{"////+AAAAAA="}
)

// FailedTransactionError represent an error that occurred because
// stellar-core rejected the transaction.  ResultXDR is a base64
// encoded TransactionResult struct
type FailedTransactionError struct {
	ResultXDR string
}

func (err *FailedTransactionError) Error() string {
	return fmt.Sprintf("tx failed: %s", err.ResultXDR)
}

func (fte *FailedTransactionError) Result() (result xdr.TransactionResult, err error) {
	err = xdr.SafeUnmarshalBase64(fte.ResultXDR, &result)
	return
}

func (fte *FailedTransactionError) TransactionResultCode() (result string, err error) {
	r, err := fte.Result()
	if err != nil {
		return
	}

	result, err = codes.String(r.Result.Code)
	return
}

func (fte *FailedTransactionError) OperationResultCodes() (result []string, err error) {
	r, err := fte.Result()
	if err != nil {
		return
	}

	oprs, ok := r.Result.GetResults()

	if !ok {
		return
	}

	result = make([]string, len(oprs))

	for i, opr := range oprs {
		result[i], err = codes.ForOperationResult(opr)
		if err != nil {
			return
		}
	}

	return
}

type AdditionalErrorInfo map[string]string

func AdditionalErrorInfoError(err error) AdditionalErrorInfo {
	return AdditionalErrorInfoStrError(err.Error())
}

func (err AdditionalErrorInfo) GetData() map[string]string {
	return map[string]string(err)
}

func (err AdditionalErrorInfo) GetError() string {
	return err.GetData()["error"]
}

func (err AdditionalErrorInfo) GetInvalidField() string {
	return err.GetData()["invalid_field"]
}

func AdditionalErrorInfoStrError(err string) AdditionalErrorInfo {
	details := make(map[string]string)
	details["error"] = err
	return AdditionalErrorInfo(details)
}

func AdditionalErrorInfoInvField(err admin.InvalidFieldError) AdditionalErrorInfo {
	details := AdditionalErrorInfoError(err.Reason)
	details["invalid_field"] = err.FieldName
	return details
}

// RestrictedTransactionError represent an error that occurred because
// horizon rejected the transaction.  ResultXDR is a base64
// encoded TransactionResult struct
type RestrictedTransactionError struct {
	FailedTransactionError
	AdditionalErrors []AdditionalErrorInfo
	TransactionErrorInfo *AdditionalErrorInfo
}

type OperationResult struct {
	Result xdr.OperationResult
	Info AdditionalErrorInfo
}

func NewRestrictedTransactionErrorTx(code xdr.TransactionResultCode, txError AdditionalErrorInfo) (*RestrictedTransactionError, error) {
	restricted, err := newRestrictedTransactionErrorOp(code, []xdr.OperationResult{}, []AdditionalErrorInfo{})
	if err != nil {
		return nil, err
	}
	restricted.TransactionErrorInfo = &txError
	return restricted, err
}

func NewRestrictedTransactionErrorOp(code xdr.TransactionResultCode, opResults []OperationResult) (*RestrictedTransactionError, error) {
	operationResults := make([]xdr.OperationResult, len(opResults))
	additionalErrors := make([]AdditionalErrorInfo, len(opResults))
	for i := range opResults {
		operationResults[i] = opResults[i].Result
		additionalErrors[i] = opResults[i].Info
	}
	return newRestrictedTransactionErrorOp(code, operationResults, additionalErrors)
}

func newRestrictedTransactionErrorOp(code xdr.TransactionResultCode, operationResults interface{}, additionalErrors []AdditionalErrorInfo) (*RestrictedTransactionError, error) {
	var xdrResult xdr.TransactionResult
	xdrResult.Result, _ = xdr.NewTransactionResultResult(code, operationResults)
	resEnv, err := xdr.MarshalBase64(xdrResult)
	if err != nil {
		return nil, err
	}

	return &RestrictedTransactionError{FailedTransactionError{resEnv}, additionalErrors, nil}, nil
}

func (err *RestrictedTransactionError) Error() string {
	return "tx violates some restrictions"
}

// MalformedTransactionError represent an error that occurred because
// a TransactionEnvelope could not be decoded from the provided data.
type MalformedTransactionError struct {
	EnvelopeXDR string
}

func (err *MalformedTransactionError) Error() string {
	return "tx malformed"
}

// RestrictedForAccountTypeError represent an error that occurred because
// operation is restricted for specified account types
type RestrictedForAccountTypeError struct {
	Reason string
}

func (err *RestrictedForAccountTypeError) Error() string {
	return err.Reason
}

// ExceededLimitError represent an error that occurred because
// operation is restricted for specified account types
type ExceededLimitError struct {
	Description string
}

func (err *ExceededLimitError) Error() string {
	return err.Description
}

// RestrictedForAccountError represent an error that occurred because
// operation is restricted for specified accounts
type RestrictedForAccountError struct {
	Reason  string
}

func (err *RestrictedForAccountError) Error() string {
	return err.Reason
}
