package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	pb "github.com/shin5ok/grpc-for-test/pb"

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
var promPort = "18080"

type healthCheck struct{}

func init() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

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
		port = "8080"
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

type newServerImplement struct{}

func (n *newServerImplement) GetMessage(ctx context.Context, name *pb.Name) (*pb.Message, error) {
	return &pb.Message{}, nil
}

func (n *newServerImplement) PutMessage(ctx context.Context, name *pb.Message) (*pb.Name, error) {
	return &pb.Name{}, nil
}
