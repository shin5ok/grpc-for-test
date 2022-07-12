package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shin5ok/grpc-for-test/pb"
	"google.golang.org/grpc"
)

var (
	GRPC_HOST = os.Getenv("GRPC_HOST")
)

func main() {
	host := flag.String("host", GRPC_HOST, "")
	number := flag.Int("number", 1, "")

	flag.Parse()

	conn, err := grpc.Dial(*host, grpc.WithInsecure())
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
