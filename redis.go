package redis_lock

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/cast"
	"time"
)

var (
	ErrNotLock       = errors.New("no lock")
	ErrWrongKeyOrTag = errors.New("wrong key or tag")
)

type Locker struct {
	key    string
	tag    string
	ch     chan struct{}
	expire time.Duration
	client *redis.Client
}

func New(key, tag string, expire time.Duration, client *redis.Client) *Locker {
	return &Locker{
		key:    key,
		tag:    tag,
		ch:     nil,
		expire: expire,
		client: client,
	}
}

func (l *Locker) Lock() error {
	result, err := l.client.SetNX(l.key, l.tag, l.expire).Result()

	if err != nil {
		return err
	}

	if !result {
		return ErrNotLock
	}

	return nil
}

func (l *Locker) Expire(expire time.Duration) (bool, error) {
	return l.client.Expire(l.key, expire).Result()
}

func (l *Locker) AutoExpire(expire, interval time.Duration) (err error) {
	ticker := time.NewTicker(interval)
	var ok bool
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//old,_ := l.TTl()
			ok, err = l.client.Expire(l.key, expire).Result()
			if err != nil || !ok {
				return fmt.Errorf("续约失败,err:%s,result: %v", err, ok)
			}
			//new,_ := l.TTl()
			//log.Infof("续约成功,key:%s, tag:%s, ttl old:%d,new:%d",l.key,l.tag, old,new)

		case <-l.ch:
			return fmt.Errorf("get close chan sign,续约结束, key:%s,tag:%s", l.key, l.tag)
		}
	}
}

func (l *Locker) Unlock() (err error) {
	//原子操作、避免误删
	var delScript = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end`)

	n, err := delScript.Eval(l.client, []string{l.key}, l.tag).Result()
	if err != nil {
		return
	}

	if cast.ToInt(n) == 0 {
		return ErrWrongKeyOrTag
	}

	l.ch = make(chan struct{})
	close(l.ch)
	return nil
}

func (l *Locker) TTl() (time.Duration, error) {
	return l.client.TTL(l.key).Result()
}

//type Redis interface {
//	SetNX(key,tag string, expire time.Duration)(bool,error)
//	Expire(key string,expire time.Duration)(bool,error)
//	TTL()(time.Duration,error)
//}
