package client

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	serverAddr string
	useTLS     bool
	conn       *grpc.ClientConn
}

func NewClient(serverAddr string, useTLS bool) *Client {
	return &Client{
		serverAddr: serverAddr,
		useTLS:     useTLS,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	var opts []grpc.DialOption
	if !c.useTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, c.serverAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
