package bolt

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func setupStore() (WatchedStore, error) {
	f, err := ioutil.TempFile("/tmp", "storetests")
	if err != nil {
		return nil, err
	}
	return NewBoltDB(f.Name(), 0644, log.New(ioutil.Discard, "test ", 0))
}

func TestWritingToBucketThatDoesNotExist(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}

	if err := s.Put("foo", "foo", nil); err == nil {
		t.Fatal("expected an error")
	}
}

func TestDeletingBucketThatDoesNotExist(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}

	if err := s.DeleteBucket("foo"); err == nil {
		t.Errorf("expected an error")
	}
}

func TestDeletingKeyInABucketThatDoesNotExist(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}

	if err := s.Delete("foo", "bar"); err == nil {
		t.Errorf("expected an error")
	}
}

func TestWritingToAKey(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}
	err = s.CreateBucket("b")
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	tests := []struct {
		k          string
		v          []byte
		shoudlFail bool
	}{
		{
			k: "foo",
			v: []byte("bar"),
		},
		{
			k: "bar",
			v: []byte("foo"),
		},
		{
			k: "foo",
			v: []byte(""),
		},
		{
			k: "foo",
			v: []byte{},
		},
		{
			k:          "foo",
			v:          nil,
			shoudlFail: true,
		},
		{
			k:          "",
			v:          []byte("bar"),
			shoudlFail: true,
		},
	}

	for i, test := range tests {
		err := s.Put("b", test.k, test.v)
		if err != nil && !test.shoudlFail {
			t.Errorf("unexpected error, test %d: %v", i, err)
		}
		if err == nil && test.shoudlFail {
			t.Errorf("expected an error, test %d", i)
		}
		// if err != nil && test.shoudlFail {
		// 	t.Logf("expected an error, test %d: %v", i, err)
		// }
	}
}

func TestWritingThenReadingKey(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}
	err = s.CreateBucket("b")
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	tests := []struct {
		k string
		v []byte
	}{
		{
			k: "foo",
			v: []byte("bar"),
		},
		{
			k: "bar",
			v: []byte("foo"),
		},
		{
			k: "foo",
			v: []byte(""),
		},
		{
			k: "foo",
			v: []byte{},
		},
	}

	for i, test := range tests {
		err := s.Put("b", test.k, test.v)
		v, err := s.Get("b", test.k)
		if err != nil {
			t.Errorf("unexpected error getting key %s, test %d: %v", test.k, i, err)
		}
		if bytes.Compare(v, test.v) != 0 {
			t.Errorf("returned value does not equal put value, test %d\nwatned: %v\ngot: %v", i, test.v, v)
		}
		if bytes.Compare(v, []byte("bad")) == 0 {
			t.Errorf("returned value should equal put value, test %d\nwatned: %v\ngot: %v", i, test.v, v)
		}
	}
}

func TestWatchingBucket(t *testing.T) {
	s, err := setupStore()
	if err != nil {
		t.Fatalf("unexpected error creating store: %v", err)
	}
	err = s.CreateBucket("b1")
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}
	err = s.CreateBucket("b2")
	if err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := s.Watch(ctx, "b1", 0)

	ctx2, cancel2 := context.WithCancel(context.Background())
	ch2 := s.Watch(ctx2, "b2", 0)

	type kv struct {
		k string
		v []byte
	}
	got1 := make([]kv, 0)
	// watch for messages
	go func(got1 *[]kv) {
		for r := range ch {
			t.Logf("notified ch1: %s", r.Key)
			*got1 = append(*got1, kv{k: r.Key, v: r.Value})
		}
	}(&got1)

	got2 := make([]kv, 0)
	// watch for messages
	go func(got2 *[]kv) {
		for r := range ch2 {
			t.Logf("notified ch2: %s", r.Key)
			*got2 = append(*got2, kv{k: r.Key, v: r.Value})
		}
	}(&got2)

	put1 := []kv{
		{
			k: "foo",
			v: []byte("bar"),
		},
		{
			k: "bar",
			v: []byte("foo"),
		},
		{
			k: "foo",
			v: []byte(""),
		},
		{
			k: "foo",
			v: []byte("foobar"),
		},
	}

	put2 := []kv{
		{
			k: "bar",
			v: []byte("foo"),
		},
		{
			k: "foo",
			v: []byte("bar"),
		},
		{
			k: "bar",
			v: []byte(""),
		},
		{
			k: "bar",
			v: []byte("foobar"),
		},
		{
			k: "barfoo",
			v: []byte("foobar"),
		},
	}

	// put on 2 different buckets
	go func() {
		for _, m := range put1 {
			if err := s.Put("b1", m.k, m.v); err != nil {
				t.Errorf("unexpected error putting key %s, value %v", m.k, m.v)
			}
		}
		fmt.Println("done")
	}()
	// write to a different bucket, should have 2 watchers not interfering
	go func() {
		for _, m := range put2 {
			if err := s.Put("b2", m.k, m.v); err != nil {
				t.Errorf("unexpected error putting key %s, value %v", m.k, m.v)
			}
		}
		fmt.Println("done")
	}()

	// let the gor be notified
	time.Sleep(100 * time.Millisecond)
	// compare put and got, should be the same
	if len(put1) != len(got1) {
		t.Errorf("length of got did not equal got, wanted %d, got %d", len(put1), len(got1))
	} else {
		for i, m := range got1 {
			if m.k != put1[i].k || bytes.Compare(m.v, put1[i].v) != 0 {
				t.Errorf("got did not equal %d got, wanted %v, got %v", i, put1[i], m)
			}
		}
	}
	// compare put and got, should be the same
	if len(put2) != len(got2) {
		t.Errorf("length of got did not equal got, wanted %d, got %d", len(put2), len(got2))
	} else {
		for i, m := range got2 {
			if m.k != put2[i].k || bytes.Compare(m.v, put2[i].v) != 0 {
				t.Errorf("got did not equal %d got, wanted %v, got %v", i, put2[i], m)
			}
		}
	}

	// delete key, should notify watcher with a nil value
	if err := s.Delete("b2", "foo"); err != nil {
		t.Errorf("unexpected error deleting key foo")
	}
	time.Sleep(100 * time.Millisecond)
	if len(put2) == len(got2) {
		t.Errorf("length of put equaled equal to got, shoudl not be as we deleted a key")
	} else {
		if got2[len(got2)-1].v != nil {
			t.Errorf("expected delted key to notify with a nil value, instead for %v", got2[len(got2)-1])
		}
	}

	cancel()
	// write after canceling watches, should not notify the watcher
	if err := s.Put("b1", "alice", []byte("bob")); err != nil {
		t.Errorf("unexpected error putting key after cancel()")
	}
	// compare put and got, should be the same
	if len(put1) != len(got1) {
		t.Errorf("length of got did not equal got, wanted %d, got %d", len(put1), len(got1))
	} else {
		for i, m := range got1 {
			if m.k != put1[i].k || bytes.Compare(m.v, put1[i].v) != 0 {
				t.Errorf("got did not equal got, wanted %v, got %v", put1[i], m)
			}
		}
	}
	cancel2()
}
