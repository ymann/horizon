package history

import (
	sq "github.com/lann/squirrel"
	"github.com/openbankit/horizon/log"
)

// CreateAuditLogEntry adds row to audit_log
func (q *Q) CreateAuditLogEntry(auditLog *AuditLog) error {
	if auditLog == nil {
		log.Warn("Tring to insern nil in audit log")
	}
	sql := createAuditLogEntry.Values(auditLog.Actor, auditLog.Subject, auditLog.Action, auditLog.Meta)
	_, err := q.Exec(sql)

	return err
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *Q) GetAllAuditLogs() (auditLog []AuditLog, err error) {
	err = q.Select(&auditLog, selectAuditLog)
	return
}

func (q *Q) DeleteAuditLog() error {
	_, err := q.Exec(sq.Delete("audit_log"))
	return err
}

var selectAuditLog = sq.Select("aud.*").From("audit_log aud")
var createAuditLogEntry = sq.Insert("audit_log").Columns(
	"actor",
	"subject",
	"action",
	"meta",
)
