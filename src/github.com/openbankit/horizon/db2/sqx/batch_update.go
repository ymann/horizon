package sqx

import (
	"errors"
	sq "github.com/lann/squirrel"
)

type BatchUpdateBuilder struct {
	*BatchInsertBuilder
	columnsToUpdate []string
	where           string
}

func BatchUpdate(insert *BatchInsertBuilder, columnsToUpdate []string, where string) *BatchUpdateBuilder {
	return &BatchUpdateBuilder{
		BatchInsertBuilder: insert,
		columnsToUpdate:    columnsToUpdate,
		where:              where,
	}
}

func (update *BatchUpdateBuilder) Update(value Updateable) error {
	if update.Err != nil {
		return update.Err
	}

	if update.tryUpdate(value.Hash(), value) {
		return update.Err
	}

	params := value.GetParams()
	if len(params) != len(update.columnsToUpdate) {
		update.Err = errors.New("Mistmatched number of columns and params for update")
		return update.Err
	}

	query := sq.Update(update.table)
	for i := range params {
		query = query.Set(update.columnsToUpdate[i], params[i])
	}
	query = query.Where(update.where, value.GetKeyParams()...)
	_, update.Err = update.repo.Exec(query)
	return update.Err
}
