package session

import (
	"github.com/openbankit/horizon/cache"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/ingest/session/ingestion"
)

// Session represents a single attempt at ingesting data into the history
// database.
type Session struct {
	Cursor    *Cursor
	Ingestion *ingestion.Ingestion

	// ClearExisting causes the session to clear existing data from the horizon db
	// when the session is run.
	ClearExisting bool

	// Metrics is a reference to where the session should record its metric information
	Metrics *IngesterMetrics

	//
	// Results fields
	//

	// Ingested is the number of ledgers that were successfully ingested during
	// this session.
	Ingested int
}

// NewSession initialize a new ingestion session, from `first` to `last`
func NewSession(first, last int32, horizonDB *db2.Repo, coreDB *db2.Repo, historyAccountCache *cache.HistoryAccount, metrics *IngesterMetrics, currentVersion int) *Session {
	hdb := horizonDB.Clone()

	return &Session{
		Ingestion: ingestion.New(hdb, historyAccountCache, currentVersion),
		Cursor:    NewCursor(coreDB, first, last, metrics.LoadLedgerTimer),
		Metrics:   metrics,
	}
}
