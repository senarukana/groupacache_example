package dbclient

import (
	"fmt"
	"net/rpc"

	"test/groupcache/protocol"
)

const (
	IDLE = iota
	CONNECTED
)

type Client struct {
	*rpc.Client
	addr  string
	state int
}

func NewClient(addr string) *Client {
	return &Client{
		addr:  addr,
		state: IDLE,
	}
}

func (c *Client) dial() error {
	conn, err := rpc.DialHTTP("tcp", c.addr)
	if err != nil {
		return fmt.Errorf("CONNECT ERROR: %s", err.Error())
	}
	c.Client = conn
	c.state = CONNECTED
	return nil
}

func (c *Client) Get(key string) (string, error) {
	if c.state == IDLE {
		if err := c.dial(); err != nil {
			return "", err
		}
	}
	args := &protocol.GetRequest{
		Key: key,
	}
	var reply protocol.GetResponse

	if err := c.Call("DBServer.Get", args, &reply); err != nil {
		c.state = IDLE
		return "", err
	}
	return reply.Value, nil
}

func (c *Client) Set(key string, value string) error {
	if c.state == IDLE {
		if err := c.dial(); err != nil {
			return err
		}
	}
	args := &protocol.SetRequest{
		Key:   key,
		Value: value,
	}
	var reply protocol.SetResult

	if err := c.Call("DBServer.Set", args, &reply); err != nil {
		c.state = IDLE
		return err
	}
	// TODO: reply
	if reply != 0 {

	}
	return nil
}
