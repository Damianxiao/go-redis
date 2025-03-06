package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	address string
}

func NewClient(address string) *Client {
	return &Client{
		address: address,
	}
}

func (c *Client) Set(ctx context.Context, key, val string) error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	err = wr.WriteArray([]resp.Value{
		resp.StringValue("SET"), resp.StringValue(key), resp.StringValue(val),
	})
	if err != nil {
		return err
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) error {
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	wr := resp.NewWriter(&buf)
	err = wr.WriteArray([]resp.Value{
		resp.StringValue("GET"), resp.StringValue(key),
	})
	if err != nil {
		return err
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
