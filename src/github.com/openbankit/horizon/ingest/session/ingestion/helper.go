package ingestion

import (
	"fmt"
	"github.com/openbankit/go-base/xdr"
	sq "github.com/lann/squirrel"
)

func formatTimeBounds(bounds *xdr.TimeBounds) interface{} {
	if bounds == nil {
		return nil
	}

	if bounds.MaxTime == 0 {
		return sq.Expr("?::int8range", fmt.Sprintf("[%d,]", bounds.MinTime))
	}

	return sq.Expr("?::int8range", fmt.Sprintf("[%d,%d]", bounds.MinTime, bounds.MaxTime))
}
