package zgo

import (
	"errors"
	"fmt"
	"git.zhugefang.com/gocore/zgo/config"
	"git.zhugefang.com/gocore/zgo/zgocache"
	"git.zhugefang.com/gocore/zgo/zgoes"
	"git.zhugefang.com/gocore/zgo/zgokafka"
	"git.zhugefang.com/gocore/zgo/zgolog"
	"git.zhugefang.com/gocore/zgo/zgomongo"
	"git.zhugefang.com/gocore/zgo/zgomysql"
	"git.zhugefang.com/gocore/zgo/zgonsq"
	"git.zhugefang.com/gocore/zgo/zgopika"
	"git.zhugefang.com/gocore/zgo/zgoredis"
	"git.zhugefang.com/gocore/zgo/zgoutils"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"strings"
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

func (opt *Options) init() error {
	//init config
	if opt.Env == "" {
		opt.Env = config.Local
	} else {
		if opt.Env != config.Local && opt.Env != config.Dev && opt.Env != config.Qa && opt.Env != config.Pro {
			return errors.New("error env,must be local/dev/qa/pro !")
		}
		if opt.Project == "" {
			return errors.New("u msut input your Project name to zgo.Engine func .")
		}
	}

	//如果connCh有值表示启用了etcd为配置中心，并watch了key，等待变更ing...
	resKvs, connCh, cacheLogCh, delConnCh, delCacheLogCh := config.InitConfig(opt.Env, opt.Project)

	//监听put资源组件
	opt.watchPutConn(connCh)
	//监听delete资源组件
	opt.watchDeleteConn(delConnCh)

	//监听put的cache和log操作
	opt.watchPutCacheOrLog(cacheLogCh)
	//监听删除cache和log操作
	opt.watchDeleteCacheOrLog(delCacheLogCh)

	//解析etcd中的配置
	opt.parseConfig(resKvs, connCh, cacheLogCh)

	return nil
}

// parseConfig 解析etcd中的配置
func (opt *Options) parseConfig(resKvs []*mvccpb.KeyValue, connCh chan map[string][]*config.ConnDetail, cacheLogCh chan map[string]*config.CacheConfig) {
	for _, v := range resKvs {
		go func(v *mvccpb.KeyValue) {
			key := string(v.Key)
			smk := strings.Split(key, "/")
			labelType := smk[3]
			b := v.Value
			if labelType == config.EtcTKCache || labelType == config.EtcTKLog { //如果cache or log配置
				var cm config.CacheConfig

				err := zgoutils.Utils.Unmarshal(b, &cm)
				if err != nil {
					fmt.Println("反序列化当前值失败", key)
				}
				var hsm = make(map[string]*config.CacheConfig)

				hsm[key] = &cm

				cacheLogCh <- hsm

			} else if smk[1] == "project" && smk[2] == opt.Project { //init conn config by etcd

				var m []config.ConnDetail
				err := zgoutils.Utils.Unmarshal(b, &m)
				if err != nil {
					fmt.Println("反序列化当前值失败", key)
				}

				label := smk[4]
				var hsm = make(map[string][]*config.ConnDetail)
				var tmp []*config.ConnDetail
				for _, vv := range m {
					pvv := vv
					tmp = append(tmp, &pvv)

					fmt.Printf("\n**********************资源项: %s **************************\n", labelType)
					fmt.Printf("描述: %s\n", pvv.C)
					fmt.Printf("Label: %s\n", label)
					fmt.Printf("Host: %s\n", pvv.Host)
					fmt.Printf("Port: %d\n", pvv.Port)
					fmt.Printf("DbName: %s\n", pvv.DbName)
				}
				hsm[key] = tmp
				connCh <- hsm
			}
		}(v)
	}
}

func (opt *Options) watchPutCacheOrLog(cacheLogCh chan map[string]*config.CacheConfig) {
	go func() {
		if cacheLogCh != nil {
			for cm := range cacheLogCh {
				//KEY: zgo/project/项目名/log
				//KEY: zgo/project/项目名/cache

				for k, v := range cm {
					smk := strings.Split(k, "/")
					labelType := smk[3]

					config.Conf.Cache.Label = v.Label
					config.Conf.Cache.Rate = v.Rate
					config.Conf.Cache.Start = v.Start
					config.Conf.Cache.TcType = v.TcType
					config.Conf.Cache.DbType = v.DbType

					switch labelType {
					case config.EtcTKCache:

						// 从etcd初始化缓存模块
						in := zgocache.InitCacheByEtcd(v)
						Cache = <-in

						fmt.Println("[init Cache]watchPutCacheOrLog", labelType, v)
					case config.EtcTKLog:

						Log = zgolog.InitLog(opt.Project)

						cc := &config.CacheConfig{
							DbType: config.Conf.Log.DbType,
							Label:  config.Conf.Log.Label,
							Start:  config.Conf.Log.Start,
						}

						fmt.Println("[init Log]watchPutCacheOrLog", labelType, cm)

						zgolog.LogWatch <- cc
					}
				}

			}
		}

	}()
}

// watchDeleteConn 监听从etcd中删除的资源key，连接类型
func (opt *Options) watchDeleteConn(ch chan map[string][]*config.ConnDetail) {
	go func() {
		if ch != nil {
			//KEY: zgo/project/项目名/mysql/label名字
			for h := range ch {
				for k, v := range h {
					smk := strings.Split(k, "/")
					labelType := smk[3]
					label := smk[4]
					fmt.Println("[destroy conn]watchDeleteConn", labelType, label, v)
					//[destroy conn]watchDeleteConn nsq nsq_label_bj [0xc0004e6840 0xc0004e68f0]

					opt.destroyConn(labelType, label, v)
				}
			}
		}
	}()
}

// destroyConn 具体删除操作
func (opt *Options) destroyConn(labelType, label string, details []*config.ConnDetail) {
	switch labelType {
	case config.EtcTKMysql:
	case config.EtcTKMongo:
	case config.EtcTKRedis:
	case config.EtcTKPia:
	case config.EtcTKNsq:
	case config.EtcTKKafka:
	case config.EtcTKEs:
	case config.EtcTKEtc:
	}
}

// watchDeleteCacheAndLog 监听删除etcd中的 cache和log类型的key
func (opt *Options) watchDeleteCacheOrLog(ch chan map[string]*config.CacheConfig) {
	go func() {
		if ch != nil {
			//KEY: zgo/project/项目名/mysql/label名字
			for h := range ch {
				for k, v := range h {
					labelType := strings.Split(k, "/")[3]
					fmt.Println("[destroy]watchDeleteCacheAndLog:", labelType, v)
					//[destroy]watchDeleteCacheAndLog: log &{日志存储 0 /tmp 1 file 0}

					opt.destroyCacheAndLog(labelType, v)
				}
			}
		}
	}()
}

// destroyCacheAndLog 具体删除操作
func (opt *Options) destroyCacheAndLog(labelType string, cf *config.CacheConfig) {

	switch labelType {
	case config.EtcTKCache:
		//如果delete是cache todo something
		config.Conf.Cache.Label = cf.Label
		config.Conf.Cache.Rate = cf.Rate
		config.Conf.Cache.Start = 0 //停止
		config.Conf.Cache.TcType = cf.TcType
		config.Conf.Cache.DbType = cf.DbType

		in := <-zgocache.InitCache()
		Cache = in

	case config.EtcTKLog:
		//如果delete是log todo something
		config.Conf.Log.DbType = cf.DbType
		config.Conf.Log.Label = cf.Label
		config.Conf.Log.Start = 0

		cc := &config.CacheConfig{
			DbType: cf.DbType,
			Label:  cf.Label,
			Start:  0,
		}
		zgolog.LogWatch <- cc
	}
}

// watchPutConn 监听保存到etcd中的资源key，连接类型
func (opt *Options) watchPutConn(inch chan map[string][]*config.ConnDetail) {
	go func() {
		if inch != nil {
			for h := range inch {
				//KEY: zgo/project/项目名/mysql/label名字
				for k, _ := range h {
					smk := strings.Split(k, "/")
					labelType := smk[3]
					mysqlLabel := smk[4]
					hsm := make(map[string][]*config.ConnDetail)
					for k, v := range h {
						label := strings.Split(k, "/")[4] //改变label，去掉前缀
						hsm[label] = v
					}
					fmt.Println("[init again]watchPutConn:", labelType, hsm)
					//[init again]watchPutConn: nsq map[nsq_label_bj:[0xc0004e62c0 0xc0004e6420]]

					opt.initConn(labelType, hsm, mysqlLabel)
				}
			}
		}

	}()
}

// initConn具体的连接操作
func (opt *Options) initConn(labelType string, hsm map[string][]*config.ConnDetail, mysqlLabel string) {
	switch labelType {
	case config.EtcTKMysql:
		//init mysql again
		// 配置信息： 城市和数据库的关系
		cdc := config.Conf.CityDbConfig
		if len(hsm) > 0 {
			zgomysql.InitMysqlService(hsm, cdc)
			var err error
			Mysql, err = zgomysql.MysqlService(mysqlLabel)

			if err != nil {
				fmt.Println(err)
			}
		}

	case config.EtcTKMongo:
		//init mongo again
		if len(hsm) > 0 {
			in := <-zgomongo.InitMongo(hsm)
			Mongo = in
		}

	case config.EtcTKRedis:
		//init redis again
		if len(hsm) > 0 {
			in := <-zgoredis.InitRedis(hsm)
			Redis = in
		}

	case config.EtcTKPia:
		//init pika again
		if len(hsm) > 0 {
			in := <-zgopika.InitPika(hsm)
			Pika = in
		}

	case config.EtcTKNsq:
		//init nsq again
		if len(hsm) > 0 {
			in := <-zgonsq.InitNsq(hsm)
			Nsq = in

		}

	case config.EtcTKKafka:
		//init kafka again
		if len(hsm) > 0 {
			in := <-zgokafka.InitKafka(hsm)
			Kafka = in
		}

	case config.EtcTKEs:
		//init es again
		if len(hsm) > 0 {
			in := <-zgoes.InitEs(hsm)
			Es = in
		}

	case config.EtcTKEtc:
		//init etcd again
	}
}

func getMatchConfig(lds map[string][]*config.ConnDetail, us []string) map[string][]*config.ConnDetail {
	m := make(map[string][]*config.ConnDetail)
	for _, label := range us {
		//v == label_bj 用户传来的label，它并不知道具体的连接地址
		//v == label_sh 用户传来的label，它并不知道具体的连接地址
		for k, v := range lds {
			if label == k {
				m[label] = v
			}
		}
	}
	return m
}
