package zgo

import (
	"errors"
	"fmt"
	"git.zhugefang.com/gocore/zgo/config"
	"git.zhugefang.com/gocore/zgo/zgoes"
	"git.zhugefang.com/gocore/zgo/zgokafka"
	"git.zhugefang.com/gocore/zgo/zgomongo"
	"git.zhugefang.com/gocore/zgo/zgonsq"
	"git.zhugefang.com/gocore/zgo/zgopika"
	"git.zhugefang.com/gocore/zgo/zgoredis"
	"go.etcd.io/etcd/mvcc/mvccpb"

	"strings"
)

const (
	mysqlT = "mysql"
	mongoT = "mongo"
	redisT = "redis"
	pikaT  = "pika"
	nsqT   = "nsq"
	kafkaT = "kafka"
	esT    = "es"
	etcdT  = "etcd"
)

type Options struct {
	Env      string   `json:"env"`
	Project  string   `json:"project"`
	Loglevel string   `json:"loglevel"`
	Mongo    []string `json:"mongo"`
	Mysql    []string `json:"mysql"`
	Es       []string `json:"es"`
	Redis    []string `json:"redis"`
	Pika     []string `json:"pika"`
	Kafka    []string `json:"kafka"`
	Nsq      []string `json:"nsq"`
}

func (opt *Options) init() (chan *mvccpb.KeyValue, error) {
	//init config
	if opt.Env == "" {
		opt.Env = "local"
	} else {
		if opt.Env != "local" && opt.Env != "dev" && opt.Env != "qa" && opt.Env != "pro" {
			return nil, errors.New("error env,must be local/dev/qa/pro !")
		}
	}

	//如果inch有值表示启用了etcd为配置中心，并watch了key，等待变更ing...
	ladech, inch := config.InitConfig(opt.Env)
	go func() {
		if inch != nil {
			for h := range inch {
				var keyType string
				for mkey, _ := range h {
					keyType = strings.Split(mkey, "/")[1]
					//key = mkey
					break
				}
				var hsm = make(map[string][]*config.ConnDetail)
				for mkey, v := range h {
					key := strings.Split(mkey, "/")[2]
					hsm[key] = v
				}
				fmt.Println(keyType, "有变化开始init again", hsm)

				switch keyType {
				case mysqlT:
				//init mysql again

				case mongoT:
					//init mongo again
					in := <-zgomongo.InitMongo(hsm)
					Mongo = in
				case redisT:
					//init redis again
					in := <-zgoredis.InitRedis(hsm)
					Redis = in
				case pikaT:
					//init pika again
					in := <-zgopika.InitPika(hsm)
					Pika = in
				case nsqT:
					//init nsq again
					in := <-zgonsq.InitNsq(hsm)
					Nsq = in
				case kafkaT:
					//init kafka again
					in := <-zgokafka.InitKafka(hsm)
					Kafka = in
				case esT:
					//init es again
					in := <-zgoes.InitEs(hsm)
					Es = in
				case etcdT:
					//init etcd again
				}
			}
		}

	}()

	if opt.Project == "" {
		opt.Project = config.Project
	}
	if opt.Loglevel == "" {
		opt.Loglevel = config.Loglevel
	}
	//fmt.Println("-------------------------------", opt.Project, opt.Loglevel)

	return ladech, nil
}
