package zgoredis

import (
	"context"
	"errors"
	"git.zhugefang.com/gocore/zgo/config"
	"github.com/mediocregopher/radix"
	"sync"
)

//NsqResourcer 给service使用
type RedisResourcer interface {
	GetConnChan(label string) chan *radix.Pool
	Do(ctx context.Context, rcv interface{}, cmd string, args ...string) (interface{}, error)
	//Post
	Set(ctx context.Context, key string, value string, time int) (interface{}, error)
	Expire(ctx context.Context, key string, time int) (interface{}, error)
	Hset(ctx context.Context, key string, name string, value string) (interface{}, error)
	Lpush(ctx context.Context, key string, value string) (interface{}, error)
	Rpush(ctx context.Context, key string, value string) (interface{}, error)
	Sadd(ctx context.Context, key string, value string) (interface{}, error)
	Srem(ctx context.Context, key string, value string) (interface{}, error)
	//Get
	Exists(ctx context.Context, key string) (interface{}, error)
	Get(ctx context.Context, key string) (interface{}, error)
	Keys(ctx context.Context, pattern string) (interface{}, error)
	//hget
	Hget(ctx context.Context, key string, name string) (interface{}, error)
	Ttl(ctx context.Context, key string) (interface{}, error)
	Type(ctx context.Context, key string) (interface{}, error)
	Hlen(ctx context.Context, key string) (interface{}, error)
	Hdel(ctx context.Context, key string, name string) (interface{}, error)
	Hgetall(ctx context.Context, key string) (interface{}, error)
	Del(ctx context.Context, key string) (interface{}, error)

	Llen(ctx context.Context, key string) (interface{}, error)
	Lrange(ctx context.Context, key string, start int, stop int) (interface{}, error)
	Lpop(ctx context.Context, key string) (interface{}, error)
	Rpop(ctx context.Context, key string) (interface{}, error)

	Scard(ctx context.Context, key string) (interface{}, error)
	Smembers(ctx context.Context, key string) (interface{}, error)
	Sismember(ctx context.Context, key string, value string) (interface{}, error)
}

type redisResource struct {
	label    string
	mu       sync.RWMutex
	connpool ConnPooler
}

func InitRedisResource(hsm map[string][]*config.ConnDetail) {
	InitConnPool(hsm)
}

func NewRedisResource(label string) RedisResourcer {
	return &redisResource{
		label:    label,
		connpool: NewConnPool(label)}
}

//GetConnChan 返回存放连接的chan
func (r *redisResource) GetConnChan(label string) chan *radix.Pool {
	return r.connpool.GetConnChan(label)
}

func (r *redisResource) Do(ctx context.Context, rcv interface{}, cmd string, args ...string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.Cmd(rcv, cmd, args...))
}

func (r *redisResource) Set(ctx context.Context, key string, value string, time int) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	pipline := radix.Pipeline(
		radix.FlatCmd(nil, "SET", key, value),
		radix.FlatCmd(nil, "Expire", key, time),
	)
	return nil, s.Do(pipline)
}

func (r *redisResource) Expire(ctx context.Context, key string, time int) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Expire", key, time))
}

func (r *redisResource) Hset(ctx context.Context, key string, name string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Hset", key, name, value))
}

func (r *redisResource) Lpush(ctx context.Context, key string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Lpush", key, value))
}

func (r *redisResource) Rpush(ctx context.Context, key string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Rpush", key, value))
}

func (r *redisResource) Sadd(ctx context.Context, key string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Sadd", key, value))
}

func (r *redisResource) Srem(ctx context.Context, key string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	return nil, s.Do(radix.FlatCmd(nil, "Srem", key, value))
}

func (r *redisResource) Exists(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var flag int
	if err := s.Do(radix.Cmd(&flag, "Exists", key)); err != nil {
		return nil, err
	} else {
		return flag, err
	}
}

func (r *redisResource) Get(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var fooVal string
	if err := s.Do(radix.FlatCmd(&fooVal, "Get", key)); err != nil {
		return nil, err
	} else {
		return fooVal, err
	}
}

func (r *redisResource) Keys(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	if key == "*" {
		return nil, errors.New("forbidden")
	}
	var bazEls []string
	if err := s.Do(radix.FlatCmd(&bazEls, "Keys", key)); err != nil {
		return nil, err
	} else {
		return bazEls, err
	}
}

func (r *redisResource) Hget(ctx context.Context, key string, name string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var fooVal string
	if err := s.Do(radix.FlatCmd(&fooVal, "Hget", key, name)); err != nil {
		return nil, err
	} else {
		return fooVal, err
	}
}

func (r *redisResource) Ttl(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var intervltime int
	if err := s.Do(radix.FlatCmd(&intervltime, "Ttl", key)); err != nil {
		return nil, err
	} else {
		return intervltime, err
	}
}

func (r *redisResource) Type(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var dataType interface{}
	if err := s.Do(radix.FlatCmd(&dataType, "Type", key)); err != nil {
		return nil, err
	} else {
		return dataType, err
	}
}

func (r *redisResource) Hlen(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var dataLen int
	if err := s.Do(radix.FlatCmd(&dataLen, "Hlen", key)); err != nil {
		return nil, err
	} else {
		return dataLen, err
	}
}

func (r *redisResource) Hdel(ctx context.Context, key string, name string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var flag int
	if err := s.Do(radix.FlatCmd(&flag, "Hlen", key)); err != nil {
		return nil, err
	} else {
		return flag, err
	}
}

func (r *redisResource) Hgetall(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var buzMap map[string]string
	if err := s.Do(radix.FlatCmd(&buzMap, "Hgetall", key)); err != nil {
		return nil, err
	} else {
		return buzMap, err
	}
}

func (r *redisResource) Del(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var flag int
	if err := s.Do(radix.FlatCmd(&flag, "del", key)); err != nil {
		return nil, err
	} else {
		return flag, err
	}
}

func (r *redisResource) Llen(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var dataLen int
	if err := s.Do(radix.FlatCmd(&dataLen, "Llen", key)); err != nil {
		return nil, err
	} else {
		return dataLen, err
	}
}

func (r *redisResource) Lrange(ctx context.Context, key string, start int, stop int) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var listContent []string
	if err := s.Do(radix.FlatCmd(&listContent, "Lrange", key, start, stop)); err != nil {
		return nil, err
	} else {
		return listContent, err
	}
}

func (r *redisResource) Lpop(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var listContent int
	if err := s.Do(radix.FlatCmd(&listContent, "Lpop", key)); err != nil {
		return nil, err
	} else {
		return listContent, err
	}
}

func (r *redisResource) Rpop(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var listContent int
	if err := s.Do(radix.FlatCmd(&listContent, "Lpop", key)); err != nil {
		return nil, err
	} else {
		return listContent, err
	}
}

func (r *redisResource) Scard(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var setLen int
	if err := s.Do(radix.FlatCmd(&setLen, "Scard", key)); err != nil {
		return nil, err
	} else {
		return setLen, err
	}
}

func (r *redisResource) Smembers(ctx context.Context, key string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var setContent []string
	if err := s.Do(radix.FlatCmd(&setContent, "Smembers", key)); err != nil {
		return nil, err
	} else {
		return setContent, err
	}
}

func (r *redisResource) Sismember(ctx context.Context, key string, value string) (interface{}, error) {
	s := <-r.connpool.GetConnChan(r.label)
	var flag int
	if err := s.Do(radix.FlatCmd(&flag, "Sismember", key)); err != nil {
		return nil, err
	} else {
		return flag, err
	}
}
