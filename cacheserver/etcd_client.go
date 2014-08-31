package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/groupcache"
)

var (
	etcdAddressRaw string
	etcdAddressKey        = "groupcache/addresses/"
	defaultTTL     uint64 = 30 // second
)

func init() {
	flag.StringVar(&etcdAddressRaw, "etcd-addresses", "http://127.0.0.1:4001", "etcd address, seperated by ,")
}

type etcdClient struct {
	cachePool        *groupcache.HTTPPool
	etcdAddr         string
	etcdClient       *etcd.Client
	etcdResponseChan chan *etcd.Response
	exitChan         chan bool

	nodeAddr          string
	cacheClusterAddrs []string
}

func newEtcdClient(nodeAddr string, cachePool *groupcache.HTTPPool) *etcdClient {
	etcdAddresses := strings.Split(etcdAddressRaw, ",")
	client := etcd.NewClient(etcdAddresses)
	resp, err := client.Get(etcdAddressKey, false, true)
	if err != nil {
		if !strings.Contains(err.Error(), "100") {
			fmt.Printf("ETCD GET ERROR: %s\n", err.Error())
			os.Exit(-1)
		} else {
			if _, err := client.CreateDir(etcdAddressKey, 0); err != nil {
				fmt.Printf("ETCD CREATE DIR ERROR: %s\n", err.Error())
				os.Exit(-1)
			}
		}
	}
	clusterAddrs := strings.Split(resp.Node.Value, ",")
	fmt.Printf("GROUP CACHE ADDRESSES: %v\n", clusterAddrs)
	cachePool.Set(clusterAddrs...)
	return &etcdClient{
		etcdClient:        client,
		cachePool:         cachePool,
		cacheClusterAddrs: clusterAddrs,
		nodeAddr:          nodeAddr,
		etcdResponseChan:  make(chan *etcd.Response),
		exitChan:          make(chan bool),
	}
}

func (e *etcdClient) Start() {
	fmt.Printf("Start...")
	e.etcdClient.Create(etcdAddressKey+e.nodeAddr, "", defaultTTL)
	go e.etcdWatch()
	go e.etcdLoop()
}

func (e *etcdClient) Close() {
	close(e.exitChan)
}

func (e *etcdClient) etcdWatch() {
	_, err := e.etcdClient.Watch(etcdAddressKey, 0, true, e.etcdResponseChan, e.exitChan)
	if err != nil {
		fmt.Printf("GET Addresses of GROUP CACHE FROM ETCD ERROR: %s\n", err.Error())
	}
}

func (e *etcdClient) etcdLoop() {
	tick := time.NewTicker(time.Second)
	for {
		select {
		case resp := <-e.etcdResponseChan:
			if resp.Action == "create" {
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
							e.cacheClusterAddrs = append(e.cacheClusterAddrs[:i-1], e.cacheClusterAddrs[:i+1]...)
						}
					}
				}
			}
			e.cachePool.Set(e.cacheClusterAddrs...)
		case <-tick.C:
			if _, err := e.etcdClient.Update(etcdAddressKey+e.nodeAddr, "", defaultTTL); err != nil {
				fmt.Printf("ETCD: UPDATE ERROR: %s\n", err.Error())
			}
		}
	}
}
