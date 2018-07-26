package storage

import (
	"testing"
	"os"
	"github.com/stretchr/testify/require"
	"time"
)

var myCache Storer

func TestMain(m *testing.M) {
	myCache = NewCache(
		DefaultTTL(1),
		DumpPath("./var/cache.dump"),
	)
	myCache.Run()
	code := m.Run()
	myCache.Close()
	os.Exit(code)
}

func TestGet(t *testing.T)  {
	r := require.New(t)
	err := myCache.SetWithTTL("testGet", "ok", 5)
	r.NoError(err)
	val, err := myCache.Get("testGet")
	r.NoError(err)
	r.Equal("ok", val.Body)
}

func TestGetContentSlice(t *testing.T)  {
	r := require.New(t)
	var arr = []string{"ok"}
	err := myCache.SetWithTTL("testContentArr", arr, 5)
	r.NoError(err)
	val, err := myCache.GetContent("testContentArr", 0)
	r.NoError(err)
	r.Equal("ok", val)
}

func TestGetContentMap(t *testing.T)  {
	r := require.New(t)
	var mp = map[string]string{"key": "ok"}
	err := myCache.SetWithTTL("testContentMap", mp, 5)
	r.NoError(err)
	val, err := myCache.GetContent("testContentMap", "key")
	r.NoError(err)
	r.Equal("ok", val)
}

func TestSetString(t *testing.T) {
	r := require.New(t)
	err := myCache.Set("testSetString", "test")
	r.NoError(err)
	time.Sleep(time.Second*2)
	_, err = myCache.Get("testSetString")
	r.Error(err)
}

func TestSetSlice(t *testing.T) {
	r := require.New(t)
	var arr = []string{"test", "test"}
	err := myCache.Set("testSetSlice", arr)
	r.NoError(err)
	time.Sleep(time.Second*2)
	_, err = myCache.Get("testSetSlice")
	r.Error(err)
}

func TestSetMap(t *testing.T) {
	r := require.New(t)
	var mp = map[string]string{"test": "test"}
	err := myCache.Set("testSetMap", mp)
	r.NoError(err)
	time.Sleep(time.Second*2)
	_, err = myCache.Get("testSetMap")
	r.Error(err)
}

func TestSetWithTTL(t *testing.T) {
	r := require.New(t)
	err := myCache.SetWithTTL("testSetTTL", "test", 3)
	r.NoError(err)
	time.Sleep(time.Second*4)
	_, err = myCache.Get("testSetTTL")
	r.Error(err)
}

func TestRemove(t *testing.T) {
	r := require.New(t)
	err := myCache.SetWithTTL("testRemove", "test", 3)
	r.NoError(err)
	err = myCache.Remove("testRemove")
	r.NoError(err)
}

func TestKeys(t *testing.T)  {
	r := require.New(t)
	err := myCache.SetWithTTL("TestKeys", "ok", 5)
	r.NoError(err)
	result := myCache.Keys("Test*eys")
	r.Equal("TestKeys", result[0])
}