package blueleveldb

import (
	"log"
	"sync"
	"time"

	"github.com/blueberryserver/bluecore/bluemq"
)

// BLLevelDBClient ...
type BLLevelDBClient struct {
	_addr      string
	_key       string
	_sub       *bluemq.BMqSub
	_exitChan  chan struct{}
	_waitGroup *sync.WaitGroup
	_db        *BLLevelDB
}

// NewClient ...
func NewClient(addr, key string) *BLLevelDBClient {

	t := time.Now()
	timeKey := t.Format("060102-150405")
	db, err := NewOpen("db/data_"+timeKey, nil, false)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &BLLevelDBClient{
		_addr:      addr,
		_key:       key,
		_sub:       bluemq.NewSub(100),
		_exitChan:  make(chan struct{}), // exit channel
		_waitGroup: &sync.WaitGroup{},   // goroutine wait group
		_db:        db,
	}
}

// StartClient ...
func (c *BLLevelDBClient) StartClient() {
	//check timeout 10 second
	now := time.Now()
	for {
		if time.Now().After(now.Add(time.Second * 10)) {
			log.Println("timeout subscribe connection")
			return
		}

		if err := c._sub.Start(c._addr, c._key); err == nil {
			break
		}
	}
	go c._db.StartLevelDBJob(1)

	c._waitGroup.Add(1)
	go c.tick()

	c._waitGroup.Wait()
}

// StopClient ...
func (c *BLLevelDBClient) StopClient() {
	c._sub.Stop()
	c._db.StopLevelDBJob()
	close(c._exitChan)
}

func (c *BLLevelDBClient) tick() {

	defer func() {
		recover()
		c._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-c._exitChan:
			return
		case msg := <-c._sub.SubChan():
			data := msg.Frames[1]
			log.Println("subscirbe: ", string(data))
			if string(data) == "keepalive" {
				continue
			}

			// add job for leveldb
			c._db.AddLevelDBJob(func() error {
				// splite key and value
				splitindex := 0
				for i, d := range data {
					if d == 0 {
						splitindex = i
						break
					}
				}
				log.Println("put key: ", string(data[:splitindex]), " value: ", string(data[splitindex+1:]))
				return c._db.Put(data[:splitindex], data[splitindex+1:])
			})
		default:
		}
	}
}
