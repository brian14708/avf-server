package avfclient

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"

	avfv1 "github.com/brian14708/avf-server/sdk/go/avf/v1"
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

func (c *Client) Transform(
	ctx context.Context,
	req *avfv1.TransformSpec,
	input io.Reader,
	output io.Writer,
) error {
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cli, err := c.service.TransformStream(cctx, grpc.UseCompressor("gzip"))
	if err != nil {
		return err
	}
	err = cli.Send(&avfv1.TransformStreamRequest{
		RequestType: &avfv1.TransformStreamRequest_Initialize{
			Initialize: &avfv1.TransformInitialize{
				Spec: req,
			},
		},
	})
	if err != nil {
		return err
	}

	go func() {
		port := req.Inputs[0].Name
		var buf [4096]byte
		for {
			inputBytes, err := input.Read(buf[:])
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				cancel()
				return
			}
			err = cli.Send(&avfv1.TransformStreamRequest{
				RequestType: &avfv1.TransformStreamRequest_Data{
					Data: &avfv1.StreamData{
						Name:    port,
						Payload: buf[:inputBytes],
					},
				},
			})
			if err != nil {
				cancel()
				return
			}
		}
		err := cli.CloseSend()
		if err != nil {
			cancel()
			return
		}
	}()

	for {
		msg, err := cli.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if data := msg.GetData(); data != nil {
			_, err := output.Write(data.Payload)
			if err != nil {
				return err
			}
		}
	}
}
