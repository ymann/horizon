package history

import (
	sq "github.com/lann/squirrel"
	"github.com/openbankit/horizon/log"
)

// Tries to select options by name. If not found, returns nil,nil
func (q *Q) OptionsByName(name string) (*Options, error) {
	sql := selectOptions.Where("opts.name = ?", name)
	var options Options
	err := q.Get(&options, sql)
	if err != nil {
		if q.Repo.NoRows(err) {
			return nil, nil
		}

		log.Error("Failed to get options by key: " + err.Error())
		return nil, err
	}

	return &options, nil
}

// Tries to insert options
func (q *Q) OptionsInsert(options *Options) (err error) {
	if options == nil {
		return
	}

	insert := insertOptions.Values(options.Name, options.Data)
	_, err = q.Exec(insert)
	if err != nil {
		log.WithStack(err).WithError(err).WithField("options", *options).Error("Failed to insert options")
	}
	return
}

func (q *Q) OptionsUpdate(options *Options) (bool, error) {
	if options == nil {
		return false, nil
	}
	update := updateOptions.SetMap(map[string]interface{}{
		"data":   options.Data,
	}).Where("name = ?", options.Name)
	result, err := q.Exec(update)
	if err != nil {
		log.WithStack(err).WithField("options", *options).WithError(err).Error("Failed to update options")
		return false, nil
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.WithStack(err).WithField("options", *options).WithError(err).Error("Failed to update options")
		return false, nil
	}

	return rows > 0, nil
}

func (q *Q) OptionsDelete(name string) (bool, error) {
	deleteQ := deleteOptions.Where("name = ?", name)
	result, err := q.Exec(deleteQ)
	if err != nil {
		log.WithStack(err).WithError(err).Error("Failed to delete options")
		return false, err
	}

	rows, err := result.RowsAffected()
	return rows != 0, err
}

var selectOptions = sq.Select("opts.*").From("options opts")
var insertOptions = sq.Insert("options").Columns("name", "data")
var updateOptions = sq.Update("options")
var deleteOptions = sq.Delete("options")
