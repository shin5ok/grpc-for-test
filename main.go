package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pb "github.com/shin5ok/proto-grpc-simple/pb"

	"github.com/google/uuid"
	"github.com/pereslava/grpc_zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var port string = os.Getenv("PORT")
var appPort = "8080"
var promPort = "18080"

type healthCheck struct{}

func init() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

}

type newServerImplement struct{}

func (n *newServerImplement) GetMessage(ctx context.Context, name *pb.Name) (*pb.Message, error) {
	log.
		Info().
		Str("method", "PutMessage").
		Str("Name as args", fmt.Sprintf("%+v", fmt.Sprintf("%+v", name))).
		Send()

	newName := name
	message := fmt.Sprintf("The message is from Id:'%d'", newName.Id)
	return &pb.Message{Name: newName, Message: message}, nil
}

func (n *newServerImplement) PutMessage(ctx context.Context, message *pb.Message) (*pb.Name, error) {
	log.
		Info().
		Str("method", "PutMessage").
		Str("Params", fmt.Sprintf("%+v", message)).
		Send()

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100)
	nameText := uuid.New().String()
	return &pb.Name{Text: nameText, Id: int32(id)}, nil
}

func (n *newServerImplement) PingPong(ctx context.Context, message *pb.Message) (*pb.Message, error) {
	return &pb.Message{Message: "Pong"}, nil
}

func (n *newServerImplement) ListMessage(req *pb.Request, stream pb.Simple_ListMessageServer) error {
	max := int(req.Number)
	for n := 0; n < max; n++ {
		result := &pb.Message{Message: fmt.Sprintf("send %d", n)}
		if err := stream.Send(result); err != nil {
			return err
		}
	}
	return nil
}

func (n *newServerImplement) BulkPutMessage(stream pb.Simple_BulkPutMessageServer) error {
	var results []*pb.Message
	var i = 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Info().
				Str("Results", fmt.Sprintf("message %+v", results)).
				Send()
			return stream.SendAndClose(&emptypb.Empty{})
		}
		log.Info().
			Int("i", i).
			Str("data", fmt.Sprintf("%+v", req.Message)).
			Send()
		results = append(results, req)
	}
}

func main() {
	serverLogger := log.Level(zerolog.TraceLevel)
	grpc_zerolog.ReplaceGrpcLogger(zerolog.New(os.Stderr).Level(zerolog.ErrorLevel))

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(serverLogger),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewStreamServerInterceptor(serverLogger),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zerolog.NewPayloadStreamServerInterceptor(serverLogger),
		),
	)

	if port == "" {
		port = appPort
	}

	listenPort, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		serverLogger.Fatal().Msg(err.Error())
	}

	var newServer = newServerImplement{}
	pb.RegisterSimpleServer(server, &newServer)

	var h = &healthCheck{}
	health.RegisterHealthServer(server, h)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(server)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":"+promPort, nil); err != nil {
			panic(err)
		}
		serverLogger.Info().Msgf("prometheus listening on :%s\n", promPort)
	}()

	reflection.Register(server)
	serverLogger.Info().Msgf("Listening on %s\n", port)
	server.Serve(listenPort)

}

func (h *healthCheck) Check(context.Context, *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	return &health.HealthCheckResponse{
		Status: health.HealthCheckResponse_SERVING,
	}, nil
}

func (h *healthCheck) Watch(*health.HealthCheckRequest, health.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "No implementation for Watch")
}
