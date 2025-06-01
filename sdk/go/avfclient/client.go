package avfclient

import (
	avfv1 "github.com/brian14708/avf-server/sdk/go/avf/v1"
	"google.golang.org/grpc"
)

type Client struct {
	service avfv1.TransformServiceClient
}

func New(conn grpc.ClientConnInterface) *Client {
	return &Client{
		service: avfv1.NewTransformServiceClient(conn),
	}
}

func Dial(target string, opts ...grpc.DialOption) (*Client, func() error, error) {
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, nil, err
	}
	closeConn := func() error { return conn.Close() }
	return New(conn), closeConn, nil
}
