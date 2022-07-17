package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/shin5ok/proto-grpc-simple/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	GRPC_HOST = os.Getenv("GRPC_HOST")
)

func main() {
	host := flag.String("host", GRPC_HOST, "")
	number := flag.Int("number", 1, "")
	insecure := flag.Bool("insecure", false, "")
	stdout := flag.Bool("stdout", false, "")
	mode := flag.String("mode", "list-message", "")

	flag.Parse()

	var conn *grpc.ClientConn
	var err error
	if *insecure {
		conn, err = grpc.Dial(*host, grpc.WithInsecure())
	} else {
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
		}
		conn, err = grpc.Dial(*host, opts...)
	}

	client := pb.NewSimpleClient(conn)

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if *mode == "list-message" {
		ctx := context.Background()
		start := time.Now()
		request := &pb.Request{Number: int32(*number)}
		stream, _ := client.ListMessage(ctx, request)
		for {
			reponse, err := stream.Recv()
			if err == io.EOF {
				finish := time.Now()
				delta := finish.Sub(start)
				fmt.Printf("%s\n", delta)
				os.Exit(0)
			}
			if err != nil {
				log.Fatal(err)
			}
			if *stdout {
				fmt.Println(reponse.Message)
			}
		}
	} else if *mode == "put-message" {
		ctx := context.Background()
		start := time.Now()

		message := "foo"
		for id := 0; id < *number; id++ {
			name := &pb.Name{Id: int32(id), Text: "foo"}
			client.PutMessage(ctx, &pb.Message{Name: name, Message: message})
		}
		finish := time.Now()
		delta := finish.Sub(start)
		fmt.Printf("%s\n", delta)

	}

}
