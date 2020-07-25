package redis_lock

import (
	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestLocker_Lock(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		var key, tag = "new", randStringRunes(10)
		var expire = time.Second * 10
		var lock = New(key, tag, expire, getClient())
		assert.IsType(t, &Locker{}, lock)

		lock.Unlock()
	})

	t.Run("lock", func(t *testing.T) {
		var key, tag = "lock", randStringRunes(10)
		var expire = time.Second * 10
		var lock = New(key, tag, expire, getClient())
		var err = lock.Lock()
		assert.Nil(t, err)
		err = lock.Unlock()
		assert.Nil(t, err)

		lock.Unlock()
	})

	t.Run("lockAgain", func(t *testing.T) {
		var key, tag = "lockAgain", randStringRunes(10)
		var expire = time.Second * 10
		var lock = New(key, tag, expire, getClient())
		var err = lock.Lock()

		assert.Nil(t, err)
		err = lock.Lock()
		assert.NotNil(t, err)

		lock.Unlock()
	})

	t.Run("twoLockWithDifferentTag", func(t *testing.T) {
		var key = "twoLockWithDifferentTag"
		var expire = time.Second * 10
		var lock = New(key, randStringRunes(10), expire, getClient())
		var lock2 = New(key, randStringRunes(10), expire, getClient())

		var err = lock.Lock()
		assert.Nil(t, err)
		err = lock2.Lock()
		assert.NotNil(t, err)

		lock.Unlock()
	})
}

func TestLocker_Unlock(t *testing.T) {
	t.Run("unlockEmptyLock", func(t *testing.T) {
		var key, tag = "unlockEmptyLock", randStringRunes(10)
		var expire = time.Second * 10
		var lock = New(key, tag, expire, getClient())
		err := lock.Unlock()
		assert.NotNil(t, err)
	})

	t.Run("unlock", func(t *testing.T) {
		var key, tag = "hello", randStringRunes(10)
		var expire = time.Second * 10
		var lock = New(key, tag, expire, getClient())
		var err = lock.Lock()

		assert.Nil(t, err)
		err = lock.Unlock()
		assert.Nil(t, err)
	})

	t.Run("UnlockAndLockAgain", func(t *testing.T) {
		var lock = newLock()
		var err = lock.Lock()

		assert.Nil(t, err)
		err = lock.Unlock()
		assert.Nil(t, err)

		err = lock.Lock()
		assert.Nil(t, err)
		assert.Nil(t, lock.Unlock())
	})

	t.Run("UnlockAndLockAgain", func(t *testing.T) {
		var lock = newLock()
		var lock2 = newLock()
		var err = lock.Lock()

		assert.Nil(t, err)

		err = lock2.Unlock()
		assert.NotNil(t, err)
		err = lock.Unlock()
		assert.Nil(t, err)
	})
}

func TestLocker_Expire(t *testing.T) {
	expire := time.Second * 10
	lock := New("foo", randStringRunes(10), expire, getClient())
	lock.Lock()

	ttl, err := lock.TTl()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, int(ttl), 0)
	assert.LessOrEqual(t, int(ttl), int(expire))

	ok, err := lock.Expire(expire * 2)
	assert.Nil(t, err)
	assert.True(t, ok)

	ttl, err = lock.TTl()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, int(ttl), int(expire))
	assert.LessOrEqual(t, int(ttl), int(expire)*2)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Username: "",
		Password: "",
		DB:       0,
	})
}

func newLock() *Locker {
	var client = getClient()
	var key, tag = "hello", randStringRunes(10)
	var expire = time.Second * 10
	var lock = New(key, tag, expire, client)

	return lock
}
