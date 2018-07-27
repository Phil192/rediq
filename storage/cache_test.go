package storage

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

var myCache Storer

func TestMain(m *testing.M) {
	myCache = NewCache(
		DumpPath("./var/cache.dump"),
	)
	myCache.Run()
	code := m.Run()
	myCache.Close()
	os.Exit(code)
}

func TestGet(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("testGet", "ok", 5)
	r.NoError(err)
	val, err := myCache.Get("testGet")
	r.NoError(err)
	r.Equal("ok", val.Body)
}

func TestGetByIndexSlice(t *testing.T) {
	r := require.New(t)
	var innerArr = []string{"ok"}
	var arr = [][]string{innerArr}
	err := myCache.Set("testContentArr", arr, 5)
	r.NoError(err)
	val, err := myCache.GetBy("testContentArr", 0)
	r.NoError(err)
	r.Equal(innerArr, val)
}

func TestGetByIndexMap(t *testing.T) {
	r := require.New(t)
	var innerMp = map[string]string{"innerKey": "ok"}
	var mp = map[string]map[string]string{"key": innerMp}
	err := myCache.Set("testContentMap", mp, 5)
	r.NoError(err)
	val, err := myCache.GetBy("testContentMap", "key")
	r.NoError(err)
	r.Equal(innerMp, val)
}

func TestSetString(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("testSetString", "test", 1)
	r.NoError(err)
	time.Sleep(time.Second * 2)
	_, err = myCache.Get("testSetString")
	r.Error(err)
}

func TestSetSlice(t *testing.T) {
	r := require.New(t)
	var arr = []string{"test", "test"}
	err := myCache.Set("testSetSlice", arr, 1)
	r.NoError(err)
	time.Sleep(time.Second * 2)
	_, err = myCache.Get("testSetSlice")
	r.Error(err)
}

func TestSetMap(t *testing.T) {
	r := require.New(t)
	var mp = map[string]string{"test": "test"}
	err := myCache.Set("testSetMap", mp, 1)
	r.NoError(err)
	time.Sleep(time.Second * 2)
	_, err = myCache.Get("testSetMap")
	r.Error(err)
}

func TestSetWithTTL(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("testSetTTL", "test", 3)
	r.NoError(err)
	time.Sleep(time.Second * 4)
	_, err = myCache.Get("testSetTTL")
	r.Error(err)
}

func TestRemove(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("testRemove", "test", 3)
	r.NoError(err)
	err = myCache.Remove("testRemove")
	r.NoError(err)
}

func TestKeys(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("TestKeys", "ok", 5)
	r.NoError(err)
	result := myCache.Keys("Test*eys")
	r.Equal("TestKeys", result[0])
}
