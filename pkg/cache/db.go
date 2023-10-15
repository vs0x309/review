package cache

import (
	"sync"
	"time"
)

func NewDB() *DB {
	return &DB{
		mu: new(sync.Mutex),
		data: make(map[string]struct {
			uTime time.Time
			data  any
		}),
	}
}

type DB struct {
	mu   *sync.Mutex
	data map[string]struct {
		uTime time.Time
		data  any
	}
}

func (d *DB) Set(key string, timeout time.Duration, data any) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.data[key] = struct {
		uTime time.Time
		data  any
	}{
		uTime: time.Now().Add(timeout),
		data:  data,
	}
}

func (d *DB) Get(key string) any {
	d.mu.Lock()
	defer d.mu.Unlock()

	if value, ok := d.data[key]; !ok {
		return nil
	} else {
		if value.uTime.After(time.Now()) {
			return value.data
		} else {
			delete(d.data, key)
			return nil
		}
	}
}
