//Package helpers contains miscellaneous helper functions
package helpers

import (
    "time"
	"strconv"
	"github.com/openbankit/go-base/xdr"
	"github.com/openbankit/go-base/strkey"
)

// SameWeek resturns true if both of the timestamps are within the same week 
func SameWeek (t1 time.Time, t2 time.Time) bool {
    year1, week1 := t1.ISOWeek()
	year2, week2 := t2.ISOWeek()
			
	return week1 == week2 && year1 == year2 
}

// MaxDate returns latest of two timestamps
func MaxDate(t1 time.Time, t2 time.Time) time.Time{
    if t1.Unix() > t2.Unix() {
        return t1
    }
    
    return t2
}

// Tries to parse timestamp, if fails return zero time
func ParseTimestamp(data string) time.Time {
	i, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(i, 0)
}

func ParseAccountId(address string) (result xdr.AccountId, err error) {
	raw, err := strkey.Decode(strkey.VersionByteAccountID, address)
	if err != nil {
		return
	}
	var key xdr.Uint256
	copy(key[:], raw)
	return xdr.NewAccountId(xdr.CryptoKeyTypeKeyTypeEd25519, key)
}