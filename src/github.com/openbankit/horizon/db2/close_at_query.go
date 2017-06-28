package db2

import (
	"fmt"
	sq "github.com/lann/squirrel"
	"time"
)

// NewCloseAtQuery creates a new CloseAtQuery struct, ensuring the start and end are set to the appropriate defaults and are valid.
func NewCloseAtQuery(start *time.Time, end *time.Time) (result CloseAtQuery, err error) {
	result.Start = start
	result.End = end
	return
}

// ApplyTo returns a new SelectBuilder after applying the time bounds effects of
// `t` to `sql`.
func (t *CloseAtQuery) ApplyTo(sql sq.SelectBuilder, col string) (sq.SelectBuilder, error) {

	if t.Start != nil {
		sql = sql.Where(fmt.Sprintf("%s >= ?", col), *t.Start)
	}

	if t.End != nil {
		sql = sql.Where(fmt.Sprintf("%s <= ?", col), *t.End)
	}

	return sql, nil
}
