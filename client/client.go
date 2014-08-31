package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"strings"

	"test/groupcache/dbclient"
	"test/groupcache/protocol"
)

func main() {
	var port = flag.String("port", "9001", "cacheserver port")
	var dbport = flag.String("dbport", "8000", "dbserver port")
	flag.Parse()

	rd := bufio.NewReader(os.Stdin)
	dbcli := dbclient.NewClient("localhost:" + *dbport)
	cacheCli, err := rpc.DialHTTP("tcp", "localhost:"+*port)
	if err != nil {
		fmt.Printf("CONNECT TO Cache Server ERROR: %s\n", err)
		return
	}
	for {
		fmt.Printf("GroupCache> ")
		line, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println("Goodbye")
			os.Exit(1)
		}
		strs := strings.Fields(line)
		switch strings.ToLower(strs[0]) {
		case "set":
			if len(strs) != 3 {
				fmt.Println("Sorry, invalid add command format. Usage: [Add key value]")
				continue
			}
			if err := dbcli.Set(strs[1], strs[2]); err != nil {
				fmt.Printf("ERROR: %s", err.Error())
			} else {
				fmt.Println("+OK")
			}
		case "get":
			if len(strs) != 2 {
				fmt.Println("Sorry, invalid find command format. Usage: [Get key]")
				continue
			}
			val, err := dbcli.Get(strs[1])
			if err != nil {
				fmt.Printf("ERROR: %s", err.Error())
			} else {
				fmt.Printf("VALUE: %s\n", val)
			}
		case "cget":
			if len(strs) != 2 {
				fmt.Println("Sorry, invalid cget command format. Usage: [CGet key]")
				continue
			}
			request := &protocol.GetRequest{
				Key: strs[1],
			}
			var resp protocol.GetResponse
			err := cacheCli.Call("CacheServer.Get", request, &resp)
			if err != nil {
				fmt.Println("ERROR: %s", err.Error())
			} else {
				fmt.Printf("VALUE: %s\n", resp.Value)
			}
		case "exit":
			fmt.Println("Goodbye~")
			os.Exit(1)
		default:
			fmt.Println("Command list: [Set key value], [Get key], [CGet key], [exit]")
		}
	}
}
