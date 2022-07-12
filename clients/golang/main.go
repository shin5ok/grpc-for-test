package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shin5ok/grpc-for-test/pb"
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

	flag.Parse()

	// https://github.com/grpc-ecosystem/grpc-cloud-run-example/blob/master/python/client.py
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

	{
		ctx := context.Background()
		request := &pb.Request{Number: int32(*number)}
		responses, _ := client.ListMessage(ctx, request)
		fmt.Printf("%-v\n", responses)
		// for r := responses {
		// 	fmt.Println(r)
		// }
	}

}
