package session

import "github.com/rcrowley/go-metrics"

// IngesterMetrics tracks all the metrics for the ingestion subsystem
type IngesterMetrics struct {
	ClearLedgerTimer  metrics.Timer
	IngestLedgerTimer metrics.Timer
	LoadLedgerTimer   metrics.Timer
}

func NewMetrics() *IngesterMetrics {
	return &IngesterMetrics{
		ClearLedgerTimer: metrics.NewTimer(),
		IngestLedgerTimer: metrics.NewTimer(),
		LoadLedgerTimer: metrics.NewTimer(),
	}
}
