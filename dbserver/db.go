package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"test/groupcache/protocol"
)

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 8000, "dbserver port")
}

type SlowDB struct {
	data map[string]string
}

func (db *SlowDB) Get(key string) string {
	time.Sleep(time.Duration(300) * time.Millisecond)
	fmt.Printf("getting %s\n", key)
	return db.data[key]
}

func (db *SlowDB) Set(key string, value string) {
	fmt.Printf("setting %s to %s\n", key, value)
	db.data[key] = value
}

func NewSlowDB() *SlowDB {
	ndb := new(SlowDB)
	ndb.data = make(map[string]string)
	return ndb
}

type DBServer struct {
	port int
	db   *SlowDB
}

func NewDBServer(port int) *DBServer {
	return &DBServer{
		port: port,
		db:   NewSlowDB(),
	}
}

func (s *DBServer) Get(args *protocol.GetRequest, reply *protocol.GetResponse) error {
	data := s.db.Get(args.Key)
	reply.Value = data
	return nil
}

func (s *DBServer) Set(args *protocol.SetRequest, reply *protocol.SetResult) error {
	s.db.Set(args.Key, args.Value)
	*reply = 0
	return nil
}

func (s *DBServer) start() {
	rpc.Register(s)
	rpc.HandleHTTP()
	fmt.Printf("DB SERVER START LISTEN, PORT %d\n", s.port)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		fmt.Printf("LISTEN ERROR: %s", err.Error())
		os.Exit(-1)
	}

	http.Serve(listen, nil)
}

func main() {
	flag.Parse()
	server := NewDBServer(port)
	server.start()
}
