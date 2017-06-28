package redis

import (
	"github.com/openbankit/go-base/keypair"
	"github.com/openbankit/horizon/log"
	"github.com/openbankit/horizon/test"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestProcessedOp(t *testing.T) {

	log.DefaultLogger.Entry.Logger.Level = log.DebugLevel

	err := Init(test.RedisURL())
	assert.Nil(t, err)

	conn := NewConnectionProvider().GetConnection()
	defer conn.Close()

	processedOpProvider := NewProcessedOpProvider(conn)

	Convey("Does not exist", t, func() {
		account, err := keypair.Random()
		So(err, ShouldBeNil)
		processedOp, err := processedOpProvider.Get(account.Address(), 1, false)
		So(err, ShouldBeNil)
		So(processedOp, ShouldBeNil)
	})
	Convey("Delete", t, func() {
		account, err := keypair.Random()
		isIncoming := true
		So(err, ShouldBeNil)
		txHash := account.Address()
		processedOp := NewProcessedOp(txHash, rand.Int(), rand.Int63(), true, time.Unix(time.Now().Unix(), 0))
		err = processedOpProvider.Insert(processedOp, time.Duration(5)*time.Second)
		So(err, ShouldBeNil)
		stored, err := processedOpProvider.Get(processedOp.TxHash, processedOp.Index, isIncoming)
		assert.Equal(t, processedOp, stored)
		err = processedOpProvider.Delete(processedOp.TxHash, processedOp.Index, isIncoming)
		So(err, ShouldBeNil)
		stored, err = processedOpProvider.Get(processedOp.TxHash, processedOp.Index, isIncoming)
		So(err, ShouldBeNil)
		So(stored, ShouldBeNil)
	})
	Convey("Storing", t, func() {
		account, err := keypair.Random()
		assert.Nil(t, err)
		txHash := account.Address()
		opIndex := rand.Int()
		isIncoming := false
		processedOp := NewProcessedOp(txHash, opIndex, rand.Int63(), false, time.Unix(time.Now().Unix(), 0))
		expireTime := time.Duration(2) * time.Second
		err = processedOpProvider.Insert(processedOp, expireTime)
		So(err, ShouldBeNil)
		storedProcessedOp, err := processedOpProvider.Get(txHash, opIndex, isIncoming)
		So(err, ShouldBeNil)
		assert.Equal(t, processedOp, storedProcessedOp)
		// timeout expires
		time.Sleep(expireTime + time.Duration(1)*time.Second)
		storedProcessedOp, err = processedOpProvider.Get(txHash, opIndex, isIncoming)
		So(err, ShouldBeNil)
		So(storedProcessedOp, ShouldBeNil)
	})
}
