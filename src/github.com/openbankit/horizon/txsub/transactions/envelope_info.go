package transactions

import "github.com/openbankit/go-base/xdr"

type EnvelopeInfo struct {
	ContentHash   string
	Sequence      uint64
	SourceAddress string
	Tx            *xdr.TransactionEnvelope
}
