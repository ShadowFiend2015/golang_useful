package oauth

import (
	"fmt"
	"testing"
	"time"
)

var cacheObj = NewCache(100)

type cacheTest struct {
	key      string
	value    interface{}
	duration time.Duration
}

type cacheData struct {
	id   int
	name string
}

var cacheTestData = []cacheTest{
	cacheTest{key: "test1", value: cacheData{id: 1, name: "abc"}, duration: time.Minute * 1},
	cacheTest{key: "test2", value: cacheData{id: 2, name: "abc2"}, duration: time.Minute * 1},
	cacheTest{key: "test3", value: cacheData{id: 3, name: "abc3"}, duration: time.Minute * 1},
}

func TestSet(t *testing.T) {
	for _, v := range cacheTestData {
		if err := cacheObj.Set(v.key, v.value, v.duration); err != nil {
			t.Errorf("set key[%s] error - %v", v.key, err)
		}
	}
}

func TestGet(t *testing.T) {
	for _, v := range cacheTestData {
		value, ok := cacheObj.Get(v.key)
		if !ok {
			t.Errorf("can't get key[%s]", v.key)
		} else {
			data := value.(cacheData)
			fmt.Printf("id[%d], name[%s]\n", data.id, data.name)
		}
	}
}
