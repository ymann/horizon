package history

import (
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/horizon/db2"
	"github.com/openbankit/horizon/db2/history/details"
	"github.com/openbankit/horizon/log"
	"encoding/json"
	sq "github.com/lann/squirrel"
	"sort"
	"strings"
)

// CommissionQ is a helper struct to aid in configuring queries that loads
// slices of Commission.
type CommissionQ struct {
	Err    error
	parent *Q
	sql    sq.SelectBuilder
}

// Commissions provides a helper to filter rows from the `commission`
// table with pre-defined filters.
func (q *Q) Commissions() *CommissionQ {
	return &CommissionQ{
		parent: q,
		sql:    selectCommission,
	}
}

// ForAccount filters the commission collection to a specific account
func (q *CommissionQ) ForAccount(aid string) *CommissionQ {
	q.sql = q.sql.Where("(com.key_value->>'from' = ? OR com.key_value->>'to' = ?)", aid, aid)
	return q
}

// ForAccountType filters the query to only commission for a specific account type
func (q *CommissionQ) ForAccountType(accountType int32) *CommissionQ {
	q.sql = q.sql.Where("(com.key_value->>'from_type' = ? OR com.key_value->>'to_type' = ?)", accountType, accountType)
	return q
}

// ForAccountType filters the query to only commission for a specific asset
func (q *CommissionQ) ForAsset(asset details.Asset) *CommissionQ {

	if asset.Type == xdr.AssetTypeAssetTypeNative.String() {
		clause := `(com.key_value->>'asset_type' = ?
		AND com.key_value ?? 'asset_code' = false
		AND com.key_value ?? 'asset_issuer' = false)`
		q.sql = q.sql.Where(clause, asset.Type)
		return q
	}

	clause := `(com.key_value->>'asset_type' = ?
	AND com.key_value->>'asset_code' = ?
	AND com.key_value->>'asset_issuer' = ?)`
	q.sql = q.sql.Where(clause, asset.Type, asset.Code, asset.Issuer)
	return q
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *CommissionQ) Page(page db2.PageQuery) *CommissionQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "com.id")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *CommissionQ) Select(dest interface{}) error {
	if q.Err != nil {
		log.WithStack(q.Err).WithError(q.Err).Error("Failed to create query to select commissions")
		return q.Err
	}

	strSql, args, _ := q.sql.ToSql()
	log.WithField("query", strSql).WithField("args", args).Debug("Tring to get commissions")
	q.Err = q.parent.Select(dest, q.sql)
	if q.Err != nil {
		log.WithStack(q.Err).WithError(q.Err).WithField("query", strSql).Error("Failed to select commissions")
	}
	return q.Err
}

func (q *Q) InsertCommission(commission *Commission) (err error) {
	if commission == nil {
		return
	}

	insert := insertCommission.Values(commission.KeyHash, commission.KeyValue, commission.FlatFee, commission.PercentFee)
	_, err = q.Exec(insert)
	if err != nil {
		log.WithStack(err).WithError(err).WithField("commission", *commission).Error("Failed to insert commission")
	}
	return
}

func (q *Q) UpdateCommission(commission *Commission) (bool, error) {
	if commission == nil {
		return false, nil
	}
	update := updateCommission.SetMap(map[string]interface{}{
		"key_value":   commission.KeyValue,
		"flat_fee":    commission.FlatFee,
		"percent_fee": commission.PercentFee,
	}).Where("key_hash = ?", commission.KeyHash)
	result, err := q.Exec(update)
	if err != nil {
		log.WithStack(err).WithField("commission", *commission).WithError(err).Error("Failed to update commission")
		return false, nil
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.WithStack(err).WithField("commission", *commission).WithError(err).Error("Failed to update commission")
		return false, nil
	}
	return rows > 0, nil
}

func (q *Q) DeleteCommission(hash string) (bool, error) {
	deleteQ := deleteCommission.Where("key_hash = ?", hash)
	result, err := q.Exec(deleteQ)
	if err != nil {
		log.WithStack(err).WithError(err).Error("Failed to delete commission")
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows != 0, err
}

// AccountByAddress loads a row from `history_accounts`, by address
func (q *Q) CommissionByKey(keys map[string]CommissionKey) (resultingCommissions []Commission, err error) {
	if len(keys) == 0 {
		return
	}
	hashes := getHashes(keys)
	sql := selectCommission.Where("com.key_hash IN (?"+strings.Repeat(", ?", len(hashes)-1)+")", hashes...)
	var storedCommissions []Commission
	err = q.Select(&storedCommissions, sql)
	if err != nil {
		log.WithStack(err).Error("Failed to get commission by key: " + err.Error())
		return nil, err
	}
	resultingCommissions = make([]Commission, 0, len(storedCommissions))
	for _, canBeCom := range storedCommissions {
		var canBeKey CommissionKey
		err := json.Unmarshal([]byte(canBeCom.KeyValue), &canBeKey)
		if err != nil {
			log.WithField("hash", canBeCom.KeyHash).WithError(err).Error("Failed to get key value for commission")
			return nil, err
		}
		key, isExist := keys[canBeCom.KeyHash]
		if !isExist {
			continue
		}
		if key.Equals(canBeKey) {
			canBeCom.weight = canBeKey.CountWeight()
			resultingCommissions = append(resultingCommissions, canBeCom)
		}
	}
	return resultingCommissions, nil
}

func (q *Q) CommissionByHash(hash string) (*Commission, error) {
	sql := selectCommission.Where("com.key_hash = ?", hash)
	var storedCommissions []Commission
	err := q.Select(&storedCommissions, sql)
	if err != nil {
		log.Error("Failed to get commission by key: " + err.Error())
		return nil, err
	}

	if len(storedCommissions) == 0 {
		return nil, nil
	}
	return &storedCommissions[0], nil
}

func (q *Q) DeleteCommissions() error {
	_, err := q.Exec(deleteCommission)
	return err
}

func (q *Q) GetHighestWeightCommission(keys map[string]CommissionKey) (resultingCommissions []Commission, err error) {
	rawCommissions, err := q.CommissionByKey(keys)
	if err != nil {
		return
	}
	log.WithField("len", len(rawCommissions)).Debug("Got commissions")
	return filterByWeight(rawCommissions), nil
}

type ByWeight []Commission

func (a ByWeight) Len() int           { return len(a) }
func (a ByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeight) Less(i, j int) bool { return a[i].weight > a[j].weight }

func filterByWeight(rawCommissions []Commission) []Commission {
	if len(rawCommissions) == 0 {
		return rawCommissions
	}
	sort.Sort(ByWeight(rawCommissions))
	bestTo := 0
	for i, val := range rawCommissions {
		if i == 0 {
			continue
		}
		if val.weight != rawCommissions[i-1].weight {
			bestTo = i - 1
			break
		}
	}
	result := rawCommissions[:bestTo+1]
	log.WithField("len", len(result)).WithField("commissions", result).Debug("Filtered commissions")
	return result
}

func getHashes(keys map[string]CommissionKey) []interface{} {
	result := make([]interface{}, len(keys))
	idx := 0
	for key := range keys {
		result[idx] = key
		idx++
	}
	return result
}

var selectCommission = sq.Select("com.*").From("commission com")
var insertCommission = sq.Insert("commission").Columns("key_hash", "key_value", "flat_fee", "percent_fee")
var updateCommission = sq.Update("commission")
var deleteCommission = sq.Delete("commission")
