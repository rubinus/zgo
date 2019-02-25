package zgoes

import (
	"context"
	"git.zhugefang.com/gocore/zgo.git/comm"
	"sync"
)

var (
	currentLabels = make(map[string][]string)
	muLabel       sync.RWMutex
)

type Eser interface {
	NewEs(label ...string) (*zgoes, error) //初始化方法

	Add(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error)
	Del(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error)
	Set(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error)
	Get(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) ([]interface{}, error)
	Search(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error)
}

//项目初始化  根据用户选择label 初始化Es实例
func InitEs(hsm map[string][]string) {
	muLabel.Lock()
	defer muLabel.Unlock()
	currentLabels = hsm
	InitEsResource(hsm)
}

//GetMongo zgo内部获取一个连接mongo
func GetEs(label ...string) (*zgoes, error) {
	l, err := comm.GetCurrentLabel(label, muLabel, currentLabels)
	if err != nil {
		return nil, err
	}
	return &zgoes{
		res: NewEsResourcer(l), //interface
	}, nil
}

type zgoes struct {
	res EsResourcer //使用resource另外的一个接口
}

func (e *zgoes) Add(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error) {
	return e.res.Add(ctx, index, table, dsl, args)
}

func (e *zgoes) Del(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error) {
	return e.res.Del(ctx, index, table, dsl, args)
}

func (e *zgoes) Set(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error) {
	return e.res.Set(ctx, index, table, dsl, args)
}

func (e *zgoes) Get(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error) {
	return e.res.Get(ctx, index, table, dsl, args)
}

func (e *zgoes) Search(ctx context.Context, index string, table string, dsl string, args map[string]interface{}) (interface{}, error) {
	return e.res.Search(ctx, index, table, dsl, args)
}
