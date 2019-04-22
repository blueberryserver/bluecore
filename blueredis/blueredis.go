package blueredis

import (
	"log"
	"sync"

	redis "gopkg.in/redis.v4"
)

// BRClient ...
type BRClient struct {
	_client       *redis.Client
	_exitChan     chan struct{}
	_redisJobChan chan interface{}
	_waitGroup    *sync.WaitGroup
}

// NewClient ...
/* options {
	Addr:	"127.0.0.1:6379",
	Password: "",
	DB: 0,
	PoolSize: 10,
	ReadOnly: false,
}
*/
func NewClient(opt *redis.Options) *BRClient {
	return &BRClient{
		_client:       redis.NewClient(opt),
		_exitChan:     make(chan struct{}), // exit channel
		_redisJobChan: make(chan interface{}, 10),
		_waitGroup:    &sync.WaitGroup{}, // goroutine wait greoup
	}
}

// PoolStats ...
func (client *BRClient) PoolStats() *redis.PoolStats {
	return client._client.PoolStats()
}

// GetClient ...
func (client *BRClient) GetClient() *redis.Client {
	return client._client
}

// Close ...
func (client *BRClient) Close() {
	client._client.Close()
}

// StartRedisJob ...
func (client *BRClient) StartRedisJob(pool int) {
	for i := 0; i < pool; i++ {
		client._waitGroup.Add(1)
		go client.tick(i)
	}
}

// StopRedisJob ..
func (client *BRClient) StopRedisJob() {
	close(client._exitChan)
}

// DBJob ...
func (client *BRClient) RedisJob(callback interface{}) {
	client._redisJobChan <- callback
}

func (client *BRClient) tick(id int) {

	defer func() {
		recover()
		client._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-client._exitChan:
			return
		case dbJob := <-client._redisJobChan:
			log.Println("redisjob id: ", id)
			dbJob.(func() error)()
		}
	}
}

// HScan ...
func (client *BRClient) HScan(key string, cursor uint64, match string, count int64) ([]string, error) {
	outputs, cursor, err := client._client.HScan(key, cursor, match, count).Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	result := []string{}
	for i := 0; i < len(outputs); i += 2 {
		result = append(result, outputs[i+1])
	}
	return result, nil
}

// HSet ...
func (client *BRClient) HSet(key, field, value string) (bool, error) {

	result, err := client._client.HSet(key, field, value).Result()
	if err != nil {
		log.Println(err)
		return false, err
	}
	return result, nil
}

// Incr ...
func (client *BRClient) Incr(key string) (int64, error) {
	result, err := client._client.Incr(key).Result()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return result, err
}
