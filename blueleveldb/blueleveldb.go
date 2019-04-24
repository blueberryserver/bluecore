package blueleveldb

import (
	"bufio"
	"log"
	"os"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// BLLevelDB ...
type BLLevelDB struct {
	_origin    *leveldb.DB
	_rep       *leveldb.DB
	_exitChan  chan struct{}
	_dbJobChan chan interface{}
	_waitGroup *sync.WaitGroup
}

// NewOpen ...
func NewOpen(dirname string, opt *opt.Options, replication bool) (*BLLevelDB, error) {
	db, err := leveldb.OpenFile(dirname, opt)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if replication == true {
		rep, err := leveldb.OpenFile(dirname+"_replication", opt)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		return &BLLevelDB{
			_origin:    db,
			_rep:       rep,
			_exitChan:  make(chan struct{}), // exit channel
			_dbJobChan: make(chan interface{}, 10),
			_waitGroup: &sync.WaitGroup{}, // goroutine wait greoup
		}, nil
	}

	return &BLLevelDB{
		_origin:    db,
		_rep:       nil,
		_exitChan:  make(chan struct{}), // exit channel
		_dbJobChan: make(chan interface{}, 10),
		_waitGroup: &sync.WaitGroup{}, // goroutine wait greoup
	}, nil
}

// Remote Replication function

// Close ...
func (db *BLLevelDB) Close() {
	db._origin.Close()
	if db._rep != nil {
		db._rep.Close()
	}
}

// StartLevelDBJob ...
func (db *BLLevelDB) StartLevelDBJob(pool int) {
	for i := 0; i < pool; i++ {
		db._waitGroup.Add(1)
		go db.tick(i)
	}
}

// StopLevelDBJob ..
func (db *BLLevelDB) StopLevelDBJob() {
	close(db._exitChan)
	db._waitGroup.Wait()
}

// AddLevelDBJob ...
func (db *BLLevelDB) AddLevelDBJob(callback interface{}) {
	db._dbJobChan <- callback
}

func (db *BLLevelDB) tick(id int) {
	log.Println("tick leveldb job id: ", id)
	defer func() {
		recover()
		db._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-db._exitChan:
			return
		case leveldbJob := <-db._dbJobChan:
			err := leveldbJob.(func() error)()
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// Put ...  writeMerge: false sync: false
func (db *BLLevelDB) Put(key, value []byte) error {
	err := db._origin.Put(key, value, nil)
	if db._rep != nil {
		err = db._rep.Put(key, value, nil)
	}
	return err
}

// Get ...
func (db *BLLevelDB) Get(key []byte) ([]byte, error) {
	return db._origin.Get(key, nil)
}

// Delete ...
func (db *BLLevelDB) Delete(key []byte) error {
	err := db._origin.Delete(key, nil)
	if db._rep != nil {
		err = db._rep.Delete(key, nil)
	}
	return err
}

// Size ...
func (db *BLLevelDB) Size(start, limit string) []int64 {
	sizes, err := db._origin.SizeOf([]util.Range{util.Range{
		Start: []byte(start),
		Limit: []byte(limit),
	}})
	if err != nil {
		return nil
	}
	return []int64(sizes)
}

// Stats ...
func (db *BLLevelDB) Stats() *leveldb.DBStats {
	stats := &leveldb.DBStats{}
	err := db._origin.Stats(stats)
	if err != nil {
		return nil
	}
	return stats
}

// NewIter ...
/* example
iter := db.NewIterator(nil, nil)
for iter.Next() {
	// Remember that the contents of the returned slice should not be modified, and
	// only valid until the next call to Next.
	key := iter.Key()
	value := iter.Value()
	...
}
iter.Release()
err = iter.Error()

OR

iter := db.NewIterator(nil, nil)
for ok := iter.Seek(key); ok; ok = iter.Next() {
	// Use key/value.
	...
}
iter.Release()
err = iter.Error()
...
*/
func (db *BLLevelDB) NewIter() iterator.Iterator {
	return db._origin.NewIterator(nil, nil)
}

// NewIterRange ...
func (db *BLLevelDB) NewIterRange(start, limit string) iterator.Iterator {
	return db._origin.NewIterator(&util.Range{
		Start: []byte(start),
		Limit: []byte(limit),
	}, nil)
}

// Apply ...
func (db *BLLevelDB) Apply(job *leveldb.Batch) error {
	err := db._origin.Write(job, nil)
	if db._rep != nil {
		db._rep.Write(job, nil)
	}
	return err
}

// NewBatch ...
/* example
batch := NewBatch()
batch.Put([]byte("foo"), []byte("value"))
batch.Put([]byte("bar"), []byte("another value"))
batch.Delete([]byte("baz"))
err = db.Apply(batch)
*/
func NewBatch() *leveldb.Batch {
	return new(leveldb.Batch)
}

// Dump ...
func (db *BLLevelDB) Dump(filename string) error {
	//close opened db
	//Close()

	// read mode reopen

	//open dump file
	dump, err := os.Create(filename)
	defer dump.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(dump)

	iter := db.NewIter()
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		w.WriteString(key)
		w.WriteString(value + "\r\n")
	}
	iter.Release()
	err = iter.Error()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// DumpLevelDB ...
func DumpLevelDB(dbname, filename string) error {
	db, err := leveldb.OpenFile(dbname, nil)
	defer db.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	//open dump file
	dump, err := os.Create(filename)
	defer dump.Close()
	if err != nil {
		return err
	}

	w := bufio.NewWriter(dump)

	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		log.Println("key:", key, " Value:", value)

		w.WriteString(key + "-")
		w.WriteString(value + "\r\n")
	}
	w.Flush()
	iter.Release()
	err = iter.Error()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
