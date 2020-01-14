package bluedb

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// BDB ...
type BDB struct {
	_db        *sql.DB
	_exitChan  chan struct{}
	_dbJobChan chan interface{}
	_waitGroup *sync.WaitGroup
}

// NewOpen ...
func NewOpen(user, pw, host string, port int, database string) (*BDB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pw, host, port, database))
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(20)
	
	if err != nil {
		return nil, err
	}
	return &BDB{
		_db:        db,
		_exitChan:  make(chan struct{}), // exit channel
		_dbJobChan: make(chan interface{}, 10),
		_waitGroup: &sync.WaitGroup{}, // goroutine wait greoup
	}, nil
}

// NewOpen ...
func NewOpenEx(user, pw, host string, port int, database string, maxidle int maxopen int) (*BDB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, pw, host, port, database))
	db.SetMaxIdleConns(maxidle)
	db.SetMaxOpenConns(maxopen)
	
	if err != nil {
		return nil, err
	}
	return &BDB{
		_db:        db,
		_exitChan:  make(chan struct{}), // exit channel
		_dbJobChan: make(chan interface{}, 10),
		_waitGroup: &sync.WaitGroup{}, // goroutine wait greoup
	}, nil
}

// Close ...
func (db *BDB) Close() {
	db._db.Close()
}

// RowsScan ...
func (db *BDB) RowsScan(rows *sql.Rows) (map[int]([]string), error) {
	//get column info
	columnNames, _ := rows.Columns()
	columnCount := len(columnNames)

	dataIndex := 0
	datas := make(map[int]([]string))
	for rows.Next() {
		columns := make([]interface{}, columnCount)
		columnPointers := make([]interface{}, columnCount)

		for colIndex := 0; colIndex < columnCount; colIndex++ {
			columnPointers[colIndex] = &columns[colIndex]
		}
		err := rows.Scan(columnPointers...)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		rowStringArr := make([]string, columnCount)
		for colIndex := 0; colIndex < columnCount; colIndex++ {
			val := columnPointers[colIndex].(*interface{})
			valStr := fmt.Sprintf("%s", *val)
			if *val == nil {
				valStr = ""
			}
			rowStringArr[colIndex] = valStr
		}

		datas[dataIndex] = rowStringArr
		dataIndex++
	}
	return datas, nil
}

// StartDBJob ...
func (db *BDB) StartDBJob(pool int) {
	for i := 0; i < pool; i++ {
		db._waitGroup.Add(1)
		go db.tick(i)
	}
}

// StopDBJob ..
func (db *BDB) StopDBJob() {
	close(db._exitChan)
}

// GetDB ...
func (db *BDB) GetDB() *sql.DB {
	return db._db
}

// DBJob ...
func (db *BDB) DBJob(callback interface{}) {
	db._dbJobChan <- callback
}

func (db *BDB) tick(id int) {

	defer func() {
		recover()
		db._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-db._exitChan:
			return
		case dbJob := <-db._dbJobChan:
			log.Println("dbjob id: ", id)
			dbJob.(func() error)()
		}
	}
}
