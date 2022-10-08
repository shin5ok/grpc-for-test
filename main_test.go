package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSimpleClient(conn)

	for _, param := range []*pb.Name{
		{Id: 10, Text: "foo"},
		{Id: 20, Text: "テスト"},
		{Id: 100000, Text: "big int number"},
	} {

		t.Run(param.Text, func(t *testing.T) {
			resp, err := client.GetMessage(ctx, param)
			// log.Printf("%+v\n", param)
			if err != nil {
				t.Error(err)
			}

			result := fmt.Sprintf("The message is from Id:'%d'", param.Id)
			if resp.Message != result {
				t.Error(resp.Message)
			}
		})

	}
}

func TestListMessage(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSimpleClient(conn)

	stream, err := client.ListMessage(ctx, &pb.Request{Number: 10})

	if err != nil {
		t.Fatal(err)
	}

	n := 0
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
		}

		shouldValue := fmt.Sprintf("send %d", n)
		if response.Message != shouldValue {
			t.Error(response.Message)
		}
		n++
	}
}

func TestPutMessage(t *testing.T) {

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSimpleClient(conn)

	name := &pb.Name{Id: 10, Text: "foo"}
	message := &pb.Message{Message: "foo is message", Name: name}

	resp, err := client.PutMessage(ctx, message)
	if err != nil {
		t.Error(err)
	}

	regexStr := `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`
	re := regexp.MustCompile(regexStr)

	if !re.MatchString(resp.Text) {
		t.Errorf("not match message: %s", resp.Text)

	}

}
