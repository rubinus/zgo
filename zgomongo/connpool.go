package zgomongo

import (
	"fmt"
	"git.zhugefang.com/gocore/zgo.git/config"
	"github.com/globalsign/mgo"
	"sync"
	"time"
)

const (
	//limitConn = 50    //如果是连接集群就是每台数据库长连接50个，单机连也是50个
	//mchSize   = 20000 //mchSize越大，越用不完，会休眠越久，不用长时间塞连接进channel
	sleepTime = 1000 //goroutine休眠时间为1000毫秒
)

var (
	connChanMap map[string]chan *mgo.Session
	mu          sync.RWMutex
	hsmu        sync.RWMutex
)

//连接对外的接口
type ConnPooler interface {
	GetConnChan(label string) chan *mgo.Session
}

type connPool struct {
	label        string
	hosts        *config.ConnDetail
	connChan     chan *mgo.Session
	clients      []*mgo.Session
	connChanChan chan chan *mgo.Session
}

func NewConnPool(label string) *connPool {
	return &connPool{
		label: label,
	}
}

//InitConnPool 对外暴露
func InitConnPool(hsm map[string][]config.ConnDetail) {
	initConnPool(hsm)
}

func initConnPool(hsm map[string][]config.ConnDetail) { //仅跑一次
	hsmu.RLock()
	defer hsmu.RUnlock()

	connChanMap = make(map[string]chan *mgo.Session)

	ch := make(chan *config.Labelconns)
	go func() {
		for lahosts := range ch {
			label := lahosts.Label
			hosts := lahosts.Hosts

			c := &connPool{
				label:        label,
				hosts:        hosts,
				connChan:     make(chan *mgo.Session, hosts.PoolSize),
				connChanChan: make(chan chan *mgo.Session, hosts.ConnSize),
			}
			connChanMap[label] = c.connChan
			go c.setConnPoolToChan(label, hosts) //call 创建连接到chan中
			//fmt.Println(label, hosts, "hsm=====",len(hsm), connChanMap)
		}
	}()

	for label, val := range hsm {
		for _, v := range val {
			lcs := &config.Labelconns{
				Label: label,
				Hosts: &v,
			}
			ch <- lcs
		}
	}
	close(ch)

}

//GetConnChan 通过label并发安全读map
func (cp *connPool) GetConnChan(label string) chan *mgo.Session {
	mu.RLock()
	defer mu.RUnlock()
	return connChanMap[label]
}

func (cp *connPool) setConnPoolToChan(label string, hosts *config.ConnDetail) {

	//每个host:port连接创建50个连接，放入slice中
	go func() {
		for sessionCh := range cp.connChanChan {
			if session, ok := <-sessionCh; ok {
				//保存channel中的连接到数组中
				cp.clients = append(cp.clients, session)
			}
		}
	}()

	for i := 0; i < hosts.ConnSize; i++ {
		//把并发创建的数据库的连接channel，放进channel中
		cp.connChanChan <- cp.createClient(fmt.Sprintf("%s:%d", hosts.Host, hosts.Port))
	}

	go func() {
		for {
			//如果连接全部创建完成，且channel中有了足够的mchSize个连接；循环确保channel中有连接
			//mchSize越大，越用不完，会休眠越久，不用长时间塞连接进channel
			if len(cp.connChan) < hosts.PoolSize && len(cp.clients) >= hosts.ConnSize/2 {
				for _, s := range cp.clients {
					if s != nil {
						cp.connChan <- s
					}
				}

			} else {
				//大多时间是在执行下面一行sleep
				time.Sleep(sleepTime * time.Millisecond)
				//fmt.Println(len(cp.connChan), "--connChan--", label, hosts.Host, hosts.Port)
			}
		}

	}()

	go func() {
		time.Sleep(2000 * time.Millisecond) //仅仅为了查看创建的连接数，创建数据库连接时间：90ms
		fmt.Println("init mongo connection to connChan ...", len(cp.connChan), label, hosts)
	}()
}

//createClient 创建客户端连接
func (cp *connPool) createClient(address string) chan *mgo.Session {
	out := make(chan *mgo.Session)
	go func() {
		//stime := time.Now()

		dialInfo := mgo.DialInfo{
			Addrs: []string{address},
			//Database: "local",
			//Username: username,
			//Password: password,
			//PoolLimit: 50000,
			Timeout: time.Duration(60 * time.Second),
		}

		session, err := mgo.DialWithInfo(&dialInfo)

		if err != nil || session == nil {
			fmt.Println(session, err)
			out <- nil
			return
		}
		session.SetMode(mgo.Monotonic, true)
		session.SetSafe(&mgo.Safe{
			WMode: "majority",
		})
		out <- session
		//fmt.Println(time.Now().Sub(stime))	//创建数据库连接时间：90ms
	}()
	return out
}
