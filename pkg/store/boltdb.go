package store

import (
	"fmt"
	"log"
	"os"

	"context"

	"github.com/boltdb/bolt"
)

// WatchedStore exposes a Watch() method
type WatchedStore interface {
	Watcher
	Store
}

// Watcher exposes a Watch() method
type Watcher interface {
	Watch(ctx context.Context, bucket string, buffer uint) <-chan WatchResponse
}

type Store interface {
	Close()
	CreateBucket(name string) error
	DeleteBucket(name string) error
	Put(bucket string, key string, value []byte) error
	Get(bucket string, key string) ([]byte, error)
	List(bucket string) ([]Entry, error)
	Delete(bucket, key string) error
}

type boltDB struct {
	db       *bolt.DB
	notifier chan<- interface{}
	cancel   func()
}

type Entry struct {
	Key   string
	Value []byte
}

type WatchResponse struct {
	Key   string
	Value []byte
}

func NewBoltDB(path string, mode os.FileMode, logger *log.Logger) (WatchedStore, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger must be provided")
	}
	db, err := bolt.Open(path, mode, nil)
	if err != nil {
		return nil, err
	}
	// setup a watch manager to watch and react to events
	wmMailbox := make(chan interface{})
	wm := watchMgr{
		mailbox:        wmMailbox,
		watchersPerKey: make(map[string]map[uint64]chan WatchResponse),
		logger:         logger,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go wm.run(ctx)
	bdb := &boltDB{
		db:       db,
		notifier: wmMailbox,
		cancel:   cancel,
	}

	return bdb, nil
}

// Close cleans up the watch manager goroutine
func (bdb *boltDB) Close() {
	bdb.cancel()
}

// Watch enables to recieve notification anytime a key in the bucket is create, modified or deleted
// Watch will send notifications on the <-chan WatchResponse channel until the a message is recieved on the ctx.Done() channel
// Watch will send a key and a nil value when a key is deleted
// Setting the buffer to 0 to send all events non-blocking, messages may be lost
func (bdb *boltDB) Watch(ctx context.Context, bucket string, buffer uint) <-chan WatchResponse {
	ch := make(chan WatchResponse, buffer)
	bdb.notifier <- newWatchMsg{bucket: bucket, ctx: ctx, respChan: ch}
	return ch
}

func (bdb *boltDB) CreateBucket(name string) error {
	db := bdb.db
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		return err
	})
	if err != nil {
		return fmt.Errorf("error creating bucket %q: %v", name, err)
	}
	return nil
}

func (bdb *boltDB) DeleteBucket(name string) error {
	db := bdb.db
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})
	if err != nil {
		return fmt.Errorf("error deleting bucket %q: %v", name, err)
	}
	// TODO how to close watchers
	return err
}

func (bdb *boltDB) Get(bucket, key string) ([]byte, error) {
	db := bdb.db
	var v []byte
	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q is nil", bucket)
		}
		v = b.Get([]byte(key))
		return nil
	})
	value := make([]byte, len(v))
	copy(value, v)
	return value, nil
}

func (bdb *boltDB) List(bucket string) ([]Entry, error) {
	db := bdb.db
	responses := []Entry{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q is nil", bucket)
		}
		b.ForEach(func(k, v []byte) error {
			value := make([]byte, len(v))
			copy(value, v)
			responses = append(responses, Entry{Key: string(k), Value: value})
			return nil
		})
		return nil
	})
	return responses, nil
}

// Put will first persist the data to the database, then notify any watchers of the event
func (bdb *boltDB) Put(bucket string, key string, value []byte) error {
	if value == nil {
		return fmt.Errorf("value cannot be nil")
	}
	db := bdb.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q is nil", bucket)
		}
		return b.Put([]byte(key), value)
	})
	if err != nil {
		return fmt.Errorf("error updating key %q: %v", key, err)
	}
	bdb.notifier <- writeOnKeyMsg{bucket: bucket, key: key, value: value}
	return nil
}

// Delete will first delete the data from the database, then notify any watchers of the event
// The watchers will recieve the key and a nil value
func (bdb *boltDB) Delete(bucket, key string) error {
	db := bdb.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("bucket %q is nil", bucket)
		}
		err := b.Delete([]byte(key))
		return err
	})
	if err != nil {
		return fmt.Errorf("error deleting key %q: %v", key, err)
	}
	// send nil value when deleting
	bdb.notifier <- writeOnKeyMsg{bucket: bucket, key: key, value: nil}
	return nil
}
