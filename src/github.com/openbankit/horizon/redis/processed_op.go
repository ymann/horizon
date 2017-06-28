package redis

import (
	"strconv"
	"time"
)

type ProcessedOp struct {
	TxHash      string
	Index       int
	Amount      int64
	IsIncoming  bool
	TimeUpdated time.Time
}

// Creates new instance of processed op from response.
func NewProcessedOp(txHash string, index int, amount int64, isIncoming bool, timeUpdated time.Time) *ProcessedOp {
	return &ProcessedOp{
		TxHash:      txHash,
		Index:       index,
		Amount:      amount,
		IsIncoming:  isIncoming,
		TimeUpdated: timeUpdated,
	}
}

func ReadProcessedOp(txHash string, index int, isIncoming bool, data map[string]int64) *ProcessedOp {
	timeUpdated := time.Unix(data["tu"], 0)
	amount := data["a"]
	return NewProcessedOp(txHash, index, amount, isIncoming, timeUpdated)
}

func (op *ProcessedOp) ToArray() []interface{} {
	return []interface{}{
		op.GetKey(),
		"a", op.Amount,
		"tu", op.TimeUpdated.Unix(),
	}
}

func (op *ProcessedOp) GetKey() string {
	return GetProcessedOpKey(op.TxHash, op.Index, op.IsIncoming)
}

func GetProcessedOpKey(txHash string, opIndex int, isIncoming bool) string {
	var direction string
	if isIncoming {
		direction = "i"
	} else {
		direction = "o"
	}
	return getKey(namespace_processed_op, txHash, strconv.Itoa(opIndex), direction)
}
