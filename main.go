package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pb "github.com/shin5ok/grpc-for-test/pb"

	"github.com/google/uuid"
	"github.com/pereslava/grpc_zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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
	message := fmt.Sprintf("The message is from Id:'%s'", newName.Id)
	return &pb.Message{Id: newName, Message: message}, nil
}

func (n *newServerImplement) PutMessage(ctx context.Context, message *pb.Message) (*pb.Name, error) {
	log.
		Info().
		Str("method", "PutMessage").
		Str("Params", fmt.Sprintf("%+v", message)).
		Send()

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100)
	name := uuid.New().String()
	return &pb.Name{Name: name, Id: strconv.Itoa(id)}, nil
}

func main() {
	serverLogger := log.Level(zerolog.TraceLevel)
	grpc_zerolog.ReplaceGrpcLogger(zerolog.New(os.Stderr).Level(zerolog.ErrorLevel))

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(serverLogger),
			grpc_zerolog.NewPayloadUnaryServerInterceptor(serverLogger),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewPayloadStreamServerInterceptor(serverLogger),
			grpc_zerolog.NewStreamServerInterceptor(serverLogger),
			grpc_prometheus.StreamServerInterceptor,
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
