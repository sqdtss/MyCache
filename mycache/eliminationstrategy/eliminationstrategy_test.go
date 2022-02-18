package eliminationstrategy

import (
	"reflect"
	"testing"
)

type String string

func (d String) Len() int {
	return len(d)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), LRU, nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}

	fifo := New(int64(0), FIFO, nil)
	fifo.Add("key1", String("1234"))
	if v, ok := fifo.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := fifo.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), LRU, nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}

	fifo := New(int64(cap), FIFO, nil)
	fifo.Add(k1, String(v1))
	fifo.Add(k2, String(v2))
	fifo.Add(k3, String(v3))

	if _, ok := fifo.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	var keys []string

	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), LRU, callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect1 := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect1, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect1)
	}

	fifo := New(int64(10), FIFO, callback)
	fifo.Add("key1", String("123456"))
	fifo.Add("k2", String("k2"))
	fifo.Add("k3", String("k3"))
	fifo.Add("k4", String("k4"))

	expect2 := []string{"key1", "k2", "key1", "k2"}

	if !reflect.DeepEqual(expect2, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect2)
	}
}

func TestAdd(t *testing.T) {
	lru := New(int64(0), LRU, nil)
	lru.Add("key", String("1"))
	lru.Add("key", String("111"))

	if lru.nbytes != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got", lru.nbytes)
	}

	fifo := New(int64(0), FIFO, nil)
	fifo.Add("key", String("1"))
	fifo.Add("key", String("111"))

	if fifo.nbytes != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got", fifo.nbytes)
	}
}
