package main

import (
	"fmt"
	"os"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/groupcache"
)

var (
	etcdAddressRaw string
	etcdAddressKey = "groupcache/addresses/"
	defaultTTL     = 30 // second
)

func init() {
	flag.StringVar(&etcdAddressRaw, "etcd-addresses", "http://127.0.0.1:4001", "etcd address, seperated by ,")
}

type EtcdClient struct {
	cachePool        *groupcache.HTTPPool
	etcdAddr         string
	etcdClient       *etcd.Client
	etcdResponseChan chan *etcd.Response
	exitChan         chan bool

	nodeAddr          string
	cacheClusterAddrs []string
}

func NewEtcdClient(nodeAddr string, cachePool *groupcache.HTTPPool) *EtcdClient {
	etcdAddresses := strings.Split(etcdAddressRaw, ",")
	etcdClient := etcd.NewClient(etcdAddresses)
	resp, err := etcdClient.Get(etcdAddressKey, false, false)
	if err != nil {
		fmt.Printf("ETCD GET ERROR: %s", err.Error())
		os.Exit(-1)
	}
	clusterAddrs := strings.Split(resp.Node.Value, ",")
	fmt.Printf("GROUP CACHE ADDRESSES: %v\n", clusterAddrs)
	cachePool.Set(clusterAddrs...)
	return &etcdClient{
		etcdClient:        etcdClient,
		cachePool:         cachePool,
		cacheClusterAddrs: clusterAddrs,
		nodeAddr:          nodeAddr,
		etcdResponseChan:  make(chan *etcd.Response),
		exitChan:          make(chan bool),
	}
}

func (e *etcdClient) Start() {
	e.etcdClient.Create(etcdAddressKey+e.nodeAddr, "", defaultTTL)
	go e.etcdWatch()
	go e.etcdLoop()
}

func (e *etcdClient) Close() {
	close(e.exitChan)
}

func (e *etcdClient) etcdWatch() {
	_, err := e.etcdClient.Watch(etcdAddressKey, 0, false, e.etcdResponseChan, e.exitChan)
	if err != nil {
		fmt.Printf("GET Addresses of GROUP CACHE FROM ETCD ERROR: %s\n", err.Error())
	}
}

func (e *etcdClient) etcdLoop() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case resp := <-e.etcdResponseChan:
			if resp.Action == "set" {
				fmt.Printf("ETCD: comes a new node %s\n", resp.Node.Key)
				e.cacheClusterAddrs = append(e.cacheClusterAddrs, resp.Node.Key)
			} else if resp.Action == "delete" {
				for i, node := range e.cacheClusterAddrs {
					if node == resp.Node.Key {
						if i == 0 {
							e.cacheClusterAddrs = e.cacheClusterAddrs[1:]
						} else if i == len(e.cacheClusterAddrs)-1 {
							e.cacheClusterAddrs = e.cacheClusterAddrs[:i-1]
						} else {
							e.cacheClusterAddrs = append(e.cacheClusterAddrs[:i-1], e.cacheClusterAddrs[:i+1])
						}
					}
				}
			} else {
				fmt.Printf("UNKNOWN ACTION: %s\n", resp.Action)
			}
			e.cachePool.Set(e.cacheClusterAddrs...)
		case <-tick:
			if _, err := e.etcdClient.Set(etcdAddressKey+e.nodeAddr, "", defaultTTL); err != nil {
				fmt.Printf("ETCD: UPDATE ERROR: %s\n", err.Error())
			}
		}
	}
}
