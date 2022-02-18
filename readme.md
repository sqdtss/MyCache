### 借鉴GeeCache实现了MyCache（https://geektutu.com/post/geecache.html）
#### 主要功能：
##### 1.实现了fifo和lru两种淘汰算法
##### 2.基于一致性哈希支持分布式并发读写
##### 3.使用了singleflight防止缓存击穿

### 结构
##### consistenthash/consistenthash.go实现了一致性哈希，对外暴露的数据结构及方法包括:
```go
// Map constains all hashed keys
type Map struct {
	hash     Hash           // Hash function
	replicas int            // Virtual node multiple
	keys     []int          // Sorted
	hashMap  map[int]string // Mapping Table of Virtual Node and Real Node
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map 

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string)

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string
```

##### eliminationstrategy/eliminationstrategy.go实现了fifo和lru淘汰算法的单机非并发cache，对外暴露的数据结构及方法包括：
```go
// Cache is a LRU/FIFO cache. It is not safe for concurrent access.
type Cache struct {
	eliminationstrategy int
	maxBytes            int64
	nbytes              int64
	ll                  *list.List
	cache               map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

// New is the Constructor of Cache
func New(maxBytes int64, eliminationstrategy int, onEvicted func(string, Value)) *Cache

// Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) 

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value)
```

##### singleflight/singleflight.go实现了只向远端节点发起一次请求以防止缓存击穿，对外暴露的数据结构及方法包括：
```go
type Group struct {
	mu sync.Mutex // protects m
	m  map[string]*call
}

// Do only initiate one request to the same key
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error)
```

##### byteview.go实现了 ByteView 用来表示缓存值，对外暴露的数据结构及方法包括：
```go
// A ByteView holds an immutable view of bytes.
type ByteView struct {
	b []byte
}

// Len returns the view's length
func (v ByteView) Len() int

// ByteSlice returns a copy of the data as a byte slice.
func (v ByteView) ByteSlice() []byte

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string 
```

##### cache.go为eliminationstrategy/eliminationstrategy.go中实现的cache添加了并发能力，对外暴露的数据结构及方法包括：
```go
type cache struct {
	mu         sync.Mutex
	es         int
	esCache    *eliminationstrategy.Cache
	cacheBytes int64
}

// Add() includes New()
func (c *cache) Add(key string, value ByteView)

func (c *cache) Get(key string) (value ByteView, ok bool)
```

##### http.go提供了mycache的分布式http请求，peers.go提供了PeerPicker和PeerGetter接口

#### mycache.go组成了最终的分布式缓存，对外暴露的数据结构及方法包括：
```go
// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
    name      string
    getter    Getter
    mainCache cache
    peers     PeerPicker
    // use singleflight.Group to make sure that each key is only fetched once
    loader *singleflight.Group
}

// NewGroup create a new instance of Group
func NewGroup(name string, es int, cacheBytes int64, getter Getter) *Group

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error)

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker)
```
