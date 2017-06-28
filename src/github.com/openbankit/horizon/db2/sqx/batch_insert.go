package sqx

import (
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/log"
	"container/list"
	"github.com/go-errors/errors"
	sqBuilder "github.com/lann/builder"
	sq "github.com/lann/squirrel"
)

const MAX_PARAMS = 65535

type Flushable interface {
	Flush() error
}

// NOT THREAD SAFE
type BatchInsertBuilder struct {
	batchSize          int
	requiredParamsSize int
	insertBuilder      sq.InsertBuilder
	totalParamsNum     int
	buffer             map[uint64]*list.List
	repo               *db2.Repo
	table              string
	log                *log.Entry

	Err error
}

// Creates new Batch inserter of columns into tableName using repo for flushes.
func BatchInsert(tableName string, repo *db2.Repo, columns ...string) *BatchInsertBuilder {
	return BatchInsertFromInsert(repo, sq.Insert(tableName).Columns(columns...))
}

// Creates new Batch inserter for builder. Values - must be empty
func BatchInsertFromInsert(repo *db2.Repo, insertBuilder sq.InsertBuilder) *BatchInsertBuilder {
	result := &BatchInsertBuilder{
		batchSize:     MAX_PARAMS,
		insertBuilder: insertBuilder,
		repo:          repo,
		buffer:        make(map[uint64]*list.List),
		log:           log.WithField("service", "batch_insert_builder"),
	}

	result.table, result.Err = getTableName(insertBuilder)
	if result.Err != nil {
		return result
	}

	if result.table == "" {
		result.Err = errors.New("Table must be set")
		return result
	}
	result.requiredParamsSize, result.Err = getColumnsNum(insertBuilder)
	if result.Err != nil {
		return result
	}

	if result.requiredParamsSize == 0 {
		result.Err = errors.New("Columns slice can't be empty!")
		return result
	}

	return result
}

func getColumnsNum(b sq.InsertBuilder) (int, error) {
	columns, ok := sqBuilder.Get(b, "Columns")
	if !ok {
		return 0, errors.New("Invalid builder. Columns must be set")
	}

	columnsList, ok := columns.([]string)
	if !ok {
		return 0, errors.New("Columns must be slice!")
	}
	return len(columnsList), nil
}

func getTableName(b sq.InsertBuilder) (string, error) {
	rawTableName, ok := sqBuilder.Get(b, "Into")
	if !ok {
		return "", errors.New("Table must be set")
	}

	table, ok := rawTableName.(string)
	if !ok {
		return "", errors.New("Table name must be string")
	}
	return table, nil
}

// Sets maximum for size of batch
func (builder *BatchInsertBuilder) BatchSize(batchSize int) *BatchInsertBuilder {
	builder.batchSize = batchSize
	return builder
}

// Returns true if flush needed. (Inserting new value after true will trigger Flush)
func (builder *BatchInsertBuilder) NeedFlush() bool {
	return builder.totalParamsNum+builder.requiredParamsSize > builder.batchSize
}

// Inserts new value into buffer. If flush is need makes it.
func (builder *BatchInsertBuilder) Insert(value Batchable) error {
	if builder.Err != nil {
		return builder.Err
	}
	builder.log.Debug("Inserting")
	hash := value.Hash()
	if builder.tryUpdate(hash, value) {
		return builder.Err
	}

	if builder.NeedFlush() {
		builder.Flush()
	}

	storage, ok := builder.buffer[hash]
	if !ok {
		storage = list.New()
		builder.buffer[hash] = storage
	}

	builder.totalParamsNum += builder.requiredParamsSize
	storage.PushBack(value)
	return builder.Err
}

// Returns true, if value was updated
func (builder *BatchInsertBuilder) tryUpdate(hash uint64, value Batchable) bool {
	storage, ok := builder.buffer[hash]
	if !ok {
		return false
	}

	for e := storage.Front(); e != nil; e = e.Next() {
		if value.Equals(e.Value) {
			e.Value = value
			return true
		}
	}
	return false
}

// Flushes buffer in repo.
func (builder *BatchInsertBuilder) Flush() error {
	builder.log.Debug("Flushing")
	if builder.totalParamsNum == 0 {
		return builder.Err
	}

	flusher := builder.insertBuilder
	for k := range builder.buffer {
		storage := builder.buffer[k]
		// clear buffer
		delete(builder.buffer, k)
		for value := storage.Front(); value != nil; value = value.Next() {
			batchable := value.Value.(Batchable)
			flusher = flusher.Values(batchable.GetParams()...)
		}
	}

	// flush
	builder.totalParamsNum = 0
	_, builder.Err = builder.repo.Exec(flusher)
	return builder.Err
}
