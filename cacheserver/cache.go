package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"

	client "test/groupcache/dbclient"
	"test/groupcache/protocol"

	"github.com/golang/groupcache"
)

var (
	cachePort  int
	serverPort int
	dbAddr     string
)

func init() {
	flag.IntVar(&cachePort, "cache-port", 8001, "group cache port")
	flag.IntVar(&serverPort, "server-port", 9001, "group cache port")
	flag.StringVar(&dbAddr, "db-address", "localhost:8000", "dbserver port")
}

type CacheServer struct {
	cachePort  int
	serverPort int
	cachePool  *groupcache.HTTPPool
	cacheGroup *groupcache.Group
	etcdClient *etcdClient
}

func NewCacheServer() *CacheServer {
	cacheAddr := fmt.Sprintf("http://localhost:%d", cachePort)
	pool := groupcache.NewHTTPPool(cacheAddr)
	cli := client.NewClient(dbAddr)

	cacheGroup := groupcache.NewGroup("DBCache", 64<<20, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dst groupcache.Sink) error {
			fmt.Printf("ASK %s FROM SERVER\n", key)
			result, err := cli.Get(key)
			if err != nil {
				fmt.Printf("cerror %s\n", err.Error())
				return err
			}
			if err := dst.SetBytes([]byte(result)); err != nil {
				fmt.Printf("SET VALUE ERROR:\n", err.Error())
			}
			return nil
		}))

	return &CacheServer{
		etcdClient: newEtcdClient(cacheAddr, pool),
		cachePool:  pool,
		cachePort:  cachePort,
		serverPort: serverPort,
		cacheGroup: cacheGroup,
	}
}

func (c *CacheServer) Get(args *protocol.GetRequest, reply *protocol.GetResponse) error {
	fmt.Printf("Receive Request : %s\n", args.Key)
	var data []byte
	err := c.cacheGroup.Get(nil, args.Key, groupcache.AllocatingByteSliceSink(&data))
	reply.Value = string(data)
	if err != nil {
		fmt.Printf("ERROR: %s", err.Error())
	}

	return err
}

func (c *CacheServer) serverStart() error {
	if err := rpc.Register(c); err != nil {
		return err
	}

	rpc.HandleHTTP()

	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", c.serverPort))
	if err != nil {
		return fmt.Errorf("LISTEN PORT %d ERROR: %s\n", serverPort, err.Error())
	}
	fmt.Printf("CACHE SERVER START LISTEN, PORT %d\n", c.serverPort)
	if err := http.Serve(listen, nil); err != nil {
		return fmt.Errorf("SERVE ERROR: %s\n", err.Error())
	}
	return nil
}

func (c *CacheServer) cacheStart() {
	fmt.Printf("GROUPCACHE START LISTEN, PORT %d\n", c.cachePort)
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", c.cachePort), http.HandlerFunc(c.cachePool.ServeHTTP)); err != nil {
		fmt.Printf("GROUPCACHE PORT %d, ERROR: %s\n", c.cachePort, err.Error())
		os.Exit(-1)
	}
}

func (c *CacheServer) start() {
	go c.etcdClient.Start()
	go c.cacheStart()
	if err := c.serverStart(); err != nil {
		fmt.Print(err.Error())
		os.Exit(-1)
	}
}

func main() {
	flag.Parse()
	server := NewCacheServer()
	server.start()
}
