package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/shin5ok/proto-grpc-simple/pb"
)

var lis *bufconn.Listener

const bufSize = 1024 * 1024

// https://github.com/castaneai/grpc-testing-with-bufconn/blob/master/server/server_test.go
func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	server := newServerImplement{}
	pb.RegisterSimpleServer(s, &server)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

}

func bufDialer(ctx context.Context, address string) (net.Conn, error) {
	return lis.Dial()
}

func TestGetMessage(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSimpleClient(conn)

	for _, param := range []*pb.Name{
		&pb.Name{Id: 10, Text: "foo"},
		&pb.Name{Id: 20, Text: "テスト"},
		&pb.Name{Id: 100000, Text: "big int number"},
	} {
		resp, err := client.GetMessage(ctx, param)
		// log.Printf("%+v\n", param)
		if err != nil {
			t.Fatal(err)
		}

		result := fmt.Sprintf("The message is from Id:'%d'", param.Id)
		if resp.Message != result {
			t.Fatal(resp.Message)
		}

	}
}

func TestListMessage(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSimpleClient(conn)

	stream, err := client.ListMessage(ctx, &pb.Request{Number: 10})

	n := 0
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		shouldValue := fmt.Sprintf("send %d", n)
		if response.Message != shouldValue {
			t.Fatal(response.Message)
		}
		n++
	}
}
