package blueleveldb

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blueberryserver/bluecore/bluemq"
)

// BLLevelDBServer ...
type BLLevelDBServer struct {
	_pubAddr   string          // pub addr
	_repAddr   string          // rep addr
	_pubSubKey string          // pub -> sub key
	_pubServer *bluemq.BMqPub  // publish server
	_repServer *bluemq.BMqRep  // response server
	_dataDB    *BLLevelDB      // data db
	_logDB     *BLLevelDB      // for log
	_exitChan  chan struct{}   // goroutine exit channel
	_waitGroup *sync.WaitGroup // goroutine wait group
}

var (
	logDBCreateTime time.Time
	identityNum     uint32
)

// NewServer ...
func NewServer(pubAddr, repAddr string, pubSubKey string) *BLLevelDBServer {
	// open leveldb for log
	logDBCreateTime = time.Now()
	timeKey := logDBCreateTime.Format("060102-150405")
	logDB, err := NewOpen("log/server_"+timeKey, nil, false)
	if err != nil {
		log.Println(err)
		return nil
	}

	// open leveldb for master data
	dataDB, err := NewOpen("db/data_"+timeKey, nil, false)
	if err != nil {
		log.Println(err)
		return nil
	}

	return &BLLevelDBServer{
		_pubAddr:   pubAddr,
		_repAddr:   repAddr,
		_pubSubKey: pubSubKey,
		_repServer: bluemq.NewRep(100),
		_pubServer: bluemq.NewPub(100),
		_dataDB:    dataDB,
		_logDB:     logDB,
		_exitChan:  make(chan struct{}), // exit channel
		_waitGroup: &sync.WaitGroup{},   // goroutine wait group
	}
}

// StartServer ...
func (s *BLLevelDBServer) StartServer() {
	// start zmq publish server for replication
	if err := s._pubServer.Start(s._pubAddr); err != nil {
		log.Println(err)
		return
	}

	// start zmq response server for request save
	if err := s._repServer.Start(s._repAddr); err != nil {
		log.Println(err)
		return
	}

	// start job tick leveldb for log
	s._logDB.StartLevelDBJob(1)

	// rep msg job tick
	s._waitGroup.Add(1)
	go s.tick()
}

// StopServer ...
func (s *BLLevelDBServer) StopServer() {
	s._pubServer.Stop()
	s._logDB.StopLevelDBJob()
	close(s._exitChan)
	s._waitGroup.Wait()
}

// SendData ...
func (s *BLLevelDBServer) SendData(key, value []byte) error {
	// create buffer of log(key + white space + value)
	data := make([]byte, len(key)+len(value)+1)
	copy(data[:len(key)], key)
	copy(data[len(key)+1:], value)
	log.Println(string(data))

	// add job leveldb for check change date and new leveldb
	s._logDB.AddLevelDBJob(func() error {
		err := s.checkAndNewLogDB()
		if err != nil {
			log.Println(err)
			return err
		}

		err = s._logDB.Put(s.timeKey(), data)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	})

	// save master db
	err := s._dataDB.Put(key, value)
	if err != nil {
		log.Println(err)
	}

	// pub msg for replication
	s._pubServer.Send(s._pubSubKey, string(data))
	return nil
}

// generator time key for leveldb dir name
func (s *BLLevelDBServer) timeKey() []byte {
	atomic.AddUint32(&identityNum, 1)
	timeKey := time.Now().Format("060102-150405-") + strconv.Itoa(int(atomic.LoadUint32(&identityNum)))
	return []byte(timeKey)
}

// check change date and new leveldb of log
func (s *BLLevelDBServer) checkAndNewLogDB() error {
	now := time.Now()
	if logDBCreateTime.Day() != now.Day() {
		log.Println("changelog leveldb dir")
		timeKey := now.Format("060102-150405")
		db, err := NewOpen("log/server_"+timeKey, nil, false)
		if err != nil {
			log.Println(err)
			return nil
		}

		// update new creation time
		logDBCreateTime = now

		// stop job proc goroutine
		s._logDB.StopLevelDBJob()
		// close open leveldb
		s._logDB.Close()

		// new leveldb bind
		s._logDB = db
		// start job proc goroutine
		s._logDB.StartLevelDBJob(1)
	}
	return nil
}

func (s *BLLevelDBServer) tick() {
	log.Println("leveldb server start tick")

	defer func() {
		recover()
		s._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-s._exitChan:
			return
		case <-time.After(10 * time.Second):
			s._pubServer.Send(s._pubSubKey, "keepalive")
		case msg := <-s._repServer.RepChan():
			data := msg.Frames[0]
			log.Println("rep: ", string(data))
			// splite key and value
			splitindex := 0
			for i, d := range data {
				if d == 0 {
					splitindex = i
					break
				}
			}

			log.Println("req data key: ", string(data[:splitindex]), " value: ", string(data[splitindex+1:]))
			if err := s.SendData(data[:splitindex], data[splitindex+1:]); err != nil {
				log.Println(err)
			}
		default:
		}
	}
}
