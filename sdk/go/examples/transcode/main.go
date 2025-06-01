package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	avfv1 "github.com/brian14708/avf-server/sdk/go/avf/v1"
	"github.com/brian14708/avf-server/sdk/go/avfclient"
)

var flagAddr = flag.String("addr", "localhost:4000", "Address of the AVF server")

func main() {
	flag.Parse()

	client, clientClose, err := avfclient.Dial(
		*flagAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer clientClose()

	in, err := os.OpenFile(flag.Arg(0), os.O_RDONLY, 0o644)
	if err != nil {
		panic(err)
	}
	defer in.Close()
	out, err := os.Create(flag.Arg(1))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = client.Transform(
		context.Background(),
		&avfv1.TransformSpec{
			Inputs: []*avfv1.InputSpec{
				{
					Name: "i",
					Format: &avfv1.FormatSpec{
						Tracks: []*avfv1.TrackSpec{
							{Type: avfv1.TrackType_TRACK_TYPE_AUDIO},
						},
					},
				},
			},
			Outputs: []*avfv1.OutputSpec{
				{
					Name: "o",
					Format: &avfv1.FormatSpec{
						Type:       avfv1.FormatType_FORMAT_TYPE_CUSTOM,
						CustomType: "mp3",
						Tracks: []*avfv1.TrackSpec{
							{
								Type: avfv1.TrackType_TRACK_TYPE_AUDIO,
								Codec: &avfv1.CodecSpec{
									Type:       avfv1.CodecType_CODEC_TYPE_CUSTOM,
									CustomType: "mp3",
								},
							},
						},
					},
				},
			},
		},
		in,
		out,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during transformation: %v\n", err)
		os.Exit(1)
	}
}
