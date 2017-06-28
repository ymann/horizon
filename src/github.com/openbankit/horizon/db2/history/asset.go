package history

import (
	"github.com/openbankit/horizon/db2"
	sq "github.com/lann/squirrel"
	"github.com/openbankit/go-base/xdr"
)

// Assets provides a helper to filter rows from the `asset` table
// with pre-defined filters.  See `AssetQ` methods for the available filters.
func (q *Q) Assets() *AssetQ {
	return &AssetQ{
		parent: q,
		sql:    selectAsset,
	}
}

func (q *Q) Asset(dest interface{}, asset xdr.Asset) error {
	var assetType xdr.AssetType
	var code, issuer string
	asset.Extract(&assetType, &code, &issuer)
	return q.AssetByParams(dest, int(assetType), code, issuer)
}

func (q *Q) AssetByParams(dest interface{}, assetType int, code string, issuer string) error {
	sql := selectAsset.Limit(1).Where("a.code = ? AND a.issuer = ? AND a.type = ?", code, issuer, assetType)
	return q.Get(dest, sql)
}

// Page specifies the paging constraints for the query being built by `q`.
func (q *AssetQ) Page(page db2.PageQuery) *AssetQ {
	if q.Err != nil {
		return q
	}

	q.sql, q.Err = page.ApplyTo(q.sql, "a.id")
	return q
}

// Select loads the results of the query specified by `q` into `dest`.
func (q *AssetQ) Select(dest interface{}) error {
	if q.Err != nil {
		return q.Err
	}

	q.Err = q.parent.Select(dest, q.sql)
	return q.Err
}

func (q *Q) InsertAsset(asset *Asset) (err error) {
	if asset == nil {
		return
	}

	insert := insertAsset.Values(asset.Type, asset.Code, asset.Issuer, asset.IsAnonymous)
	_, err = q.Exec(insert)
	return err
}

func (q *Q) UpdateAsset(asset *Asset) (bool, error) {
	if asset == nil {
		return false, nil
	}
	update := updateAsset.SetMap(map[string]interface{}{
		"type":         asset.Type,
		"code":         asset.Code,
		"issuer":       asset.Issuer,
		"is_anonymous": asset.IsAnonymous,
	}).Where("id = ?", asset.Id)
	result, err := q.Exec(update)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (q *Q) DeleteAsset(id int64) (bool, error) {
	deleteQ := deleteAsset.Where("id = ?", id)
	result, err := q.Exec(deleteQ)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows != 0, err
}

var (
	selectAsset = sq.Select("a.*").From("asset a")
	insertAsset = sq.Insert("asset").Columns("type", "code", "issuer", "is_anonymous")
	updateAsset = sq.Update("asset")
	deleteAsset      = sq.Delete("asset")
)
