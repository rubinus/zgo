package zgoredis

import (
	"context"
	"git.zhugefang.com/gocore/zgo/comm"
	"git.zhugefang.com/gocore/zgo/config"
	"sync"
)

var (
	currentLabels = make(map[string][]*config.ConnDetail)
	muLabel       sync.RWMutex
)

//Redis 对外
type Rediser interface {
	New(label ...string) (*zgoredis, error)
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

func Redis(l string) Rediser {
	return &zgoredis{
		res: NewRedisResource(l),
	}
}

//zgoredis实现了Redis的接口
type zgoredis struct {
	res RedisResourcer //使用resource另外的一个接口
}

//InitRedis 初始化连接redis
func InitRedis(hsmIn map[string][]*config.ConnDetail, label ...string) chan *zgoredis {
	muLabel.Lock()
	defer muLabel.Unlock()

	var hsm map[string][]*config.ConnDetail

	if len(label) > 0 && len(currentLabels) > 0 { //此时是destory操作,传入的hsm是nil
		//fmt.Println("--destory--前",currentLabels)
		for _, v := range label {
			delete(currentLabels, v)
		}
		hsm = currentLabels
		//fmt.Println("--destory--后",currentLabels)

	} else { //这是第一次创建操作或etcd中变更时init again操作
		hsm = hsmIn
		//currentLabels = hsm	//this operation is error
		for k, v := range hsm { //so big bug can't set hsm to currentLabels，must be for, may be have old label
			currentLabels[k] = v
		}
	}

	if len(hsm) == 0 {
		return nil
	}

	InitRedisResource(hsm)

	//自动为变量初始化对象
	initLabel := ""
	for k, _ := range hsm {
		if k != "" {
			initLabel = k
			break
		}
	}
	out := make(chan *zgoredis)
	go func() {
		in, err := GetRedis(initLabel)
		if err != nil {
			out <- nil
		}
		out <- in
		close(out)
	}()
	return out
}

func (n *zgoredis) New(label ...string) (*zgoredis, error) {
	return GetRedis(label...)
}

//GetRedis zgo内部获取一个连接redis
func GetRedis(label ...string) (*zgoredis, error) {
	l, err := comm.GetCurrentLabel(label, muLabel, currentLabels)
	if err != nil {
		return nil, err
	}
	return &zgoredis{
		res: NewRedisResource(l), //interface
	}, nil
}

func (r *zgoredis) Set(ctx context.Context, key string, value string, time int) (interface{}, error) {
	return r.res.Set(ctx, key, value, time)
}

func (r *zgoredis) Expire(ctx context.Context, key string, time int) (interface{}, error) {
	return r.res.Expire(ctx, key, time)
}

func (r *zgoredis) Hset(ctx context.Context, key string, name string, value string) (interface{}, error) {
	return r.res.Hset(ctx, key, name, value)
}

func (r *zgoredis) Lpush(ctx context.Context, key string, value string) (interface{}, error) {
	return r.res.Lpush(ctx, key, value)
}

func (r *zgoredis) Rpush(ctx context.Context, key string, value string) (interface{}, error) {
	return r.res.Lpush(ctx, key, value)
}

func (r *zgoredis) Sadd(ctx context.Context, key string, value string) (interface{}, error) {
	return r.res.Sadd(ctx, key, value)
}

func (r *zgoredis) Srem(ctx context.Context, key string, value string) (interface{}, error) {
	return r.res.Srem(ctx, key, value)
}

func (r *zgoredis) Exists(ctx context.Context, key string) (interface{}, error) {
	return r.res.Exists(ctx, key)
}

func (r *zgoredis) Get(ctx context.Context, key string) (interface{}, error) {
	return r.res.Get(ctx, key)
}

func (r *zgoredis) Keys(ctx context.Context, key string) (interface{}, error) {
	return r.res.Keys(ctx, key)
}

func (r *zgoredis) Hget(ctx context.Context, key string, name string) (interface{}, error) {
	return r.res.Hget(ctx, key, name)
}

func (r *zgoredis) Ttl(ctx context.Context, key string) (interface{}, error) {
	return r.res.Ttl(ctx, key)
}

func (r *zgoredis) Type(ctx context.Context, key string) (interface{}, error) {
	return r.res.Type(ctx, key)
}

func (r *zgoredis) Hlen(ctx context.Context, key string) (interface{}, error) {
	return r.res.Hlen(ctx, key)
}

func (r *zgoredis) Hdel(ctx context.Context, key string, name string) (interface{}, error) {
	return r.res.Hdel(ctx, key, name)
}

func (r *zgoredis) Hgetall(ctx context.Context, key string) (interface{}, error) {
	return r.res.Hgetall(ctx, key)
}

func (r *zgoredis) Del(ctx context.Context, key string) (interface{}, error) {
	return r.res.Del(ctx, key)
}

func (r *zgoredis) Llen(ctx context.Context, key string) (interface{}, error) {
	return r.res.Llen(ctx, key)
}

func (r *zgoredis) Lrange(ctx context.Context, key string, start int, stop int) (interface{}, error) {
	return r.res.Lrange(ctx, key, start, stop)
}

func (r *zgoredis) Lpop(ctx context.Context, key string) (interface{}, error) {
	return r.res.Lpop(ctx, key)
}

func (r *zgoredis) Rpop(ctx context.Context, key string) (interface{}, error) {
	return r.res.Lpop(ctx, key)
}

func (r *zgoredis) Scard(ctx context.Context, key string) (interface{}, error) {
	return r.res.Scard(ctx, key)
}

func (r *zgoredis) Smembers(ctx context.Context, key string) (interface{}, error) {
	return r.res.Smembers(ctx, key)
}

func (r *zgoredis) Sismember(ctx context.Context, key string, value string) (interface{}, error) {
	return r.res.Sismember(ctx, key, value)
}
