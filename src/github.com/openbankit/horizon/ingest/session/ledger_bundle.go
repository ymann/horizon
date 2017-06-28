package session

import (
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/core"
)

// LedgerBundle represents a single ledger's worth of novelty created by one
// ledger close
type LedgerBundle struct {
	Sequence        int32
	Header          core.LedgerHeader
	TransactionFees []core.TransactionFee
	Transactions    []core.Transaction
}

// Load runs queries against `core` to fill in the records of the bundle.
func (lb *LedgerBundle) Load(db *db2.Repo) error {
	q := &core.Q{Repo: db}

	// Load Header
	err := q.LedgerHeaderBySequence(&lb.Header, lb.Sequence)
	if err != nil {
		return err
	}

	// Load transactions
	err = q.TransactionsByLedger(&lb.Transactions, lb.Sequence)

	if err != nil {
		return err
	}

	err = q.TransactionFeesByLedger(&lb.TransactionFees, lb.Sequence)
	if err != nil {
		return err
	}

	return nil
}
