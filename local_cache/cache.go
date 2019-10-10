package oauth

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type Cache struct {
	Cap    int
	bucket sync.Map
	lru    *LRU
}

type LRU struct {
	sync.RWMutex
	l *list.List
}

type Entry struct {
	Key        string
	Value      interface{}
	Expiration int64
}

func NewCache(cap int) *Cache {
	return &Cache{
		Cap:    cap,
		bucket: sync.Map{},
		lru:    &LRU{l: list.New()},
	}
}

// check if element expired
func (c *Cache) checkExpired(e *list.Element, force bool) bool {
	entry, ok := e.Value.(*Entry)
	if !ok {
		return true
	}
	c.lru.Lock()
	defer c.lru.Unlock()
	if time.Now().Unix() > entry.Expiration || force {
		c.lru.l.Remove(e)
		c.bucket.Delete(entry.Key)
		return true
	}
	c.lru.l.MoveToBack(e)
	return false
}

func (c *Cache) Get(key string) (interface{}, bool) {
	if c.lru == nil {
		return nil, false
	}
	v, ok := c.bucket.Load(key)
	if !ok {
		return nil, false
	}
	e, _ := v.(*list.Element)
	if ok := c.checkExpired(e, false); ok {
		return nil, false
	}
	return e.Value.(*Entry).Value, true
}

func (c *Cache) GetEntry(key string) (*Entry, bool) {
	if c.lru == nil {
		return nil, false
	}
	v, ok := c.bucket.Load(key)
	if !ok {
		return nil, false
	}
	e, _ := v.(*list.Element)
	if ok := c.checkExpired(e, false); ok {
		return nil, false
	}
	return e.Value.(*Entry), true
}

func (c *Cache) Contains(key string) bool {
	v, ok := c.bucket.Load(key)
	if !ok {
		return false
	}
	e, _ := v.(*list.Element)
	return !c.checkExpired(e, false)
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) error {
	if c.lru == nil {
		return errors.New("uninitialized")
	}

	if v, ok := c.bucket.Load(key); ok {
		e, _ := v.(list.Element)
		if entry, ok := e.Value.(*Entry); ok {
			entry.Value = value
			entry.Expiration = time.Now().Add(duration).Unix()
			return nil
		}
	}

	entry := &Entry{Key: key, Value: value, Expiration: time.Now().Add(duration).Unix()}
	e := c.lru.l.PushFront(entry)
	c.bucket.Store(key, e)

	if c.lru.l.Len() > c.Cap {
		tmp := c.lru.l.Back()
		c.checkExpired(tmp, true)
	}
	return nil
}
