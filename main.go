package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"encoding/json"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"

	pb "github.com/shin5ok/proto-grpc-simple/pb"

	"github.com/google/uuid"
	"github.com/pereslava/grpc_zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var port string = os.Getenv("PORT")
var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
var domain = os.Getenv("DOMAIN")
var sleep = os.Getenv("SLEEP")
var sleepSecond int

var appPort = "8080"
var promPort = "18080"

type healthCheck struct{}
type newServerImplement struct {
	tracer trace.Tracer
}

func init() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if domain == "" {
		log.Info().Msg("domain is not set")
		os.Exit(1)
	}

	if sleep != "" {
		sleepSecond, _ = strconv.Atoi(sleep)
	}

}

func (n *newServerImplement) GetMessage(ctx context.Context, name *pb.Name) (*pb.Message, error) {
	ctx, span := n.tracer.Start(ctx, "get message")
	defer span.End()

	log.
		Info().
		Str("logging.googleapis.com/trace", span.SpanContext().TraceID().String()).
		Str("logging.googleapis.com/spanId", span.SpanContext().SpanID().String()).
		Str("method", "GetMessage").
		Str("Name as args", fmt.Sprintf("%+v", fmt.Sprintf("%+v", name))).
		Send()

	newName, err := func(ctx context.Context) (*pb.Name, error) {

		newName := name
		return newName, nil

	}(ctx)
	if err != nil {
		return nil, err
	}
	message := fmt.Sprintf("The message is from Id:'%d'", newName.Id)
	return &pb.Message{Name: newName, Message: message}, nil
}

func (n *newServerImplement) PutMessage(ctx context.Context, message *pb.Message) (*pb.Name, error) {
	ctx, span := n.tracer.Start(ctx, "put message")
	defer span.End()

	log.
		Info().
		Str("logging.googleapis.com/trace", span.SpanContext().TraceID().String()).
		Str("logging.googleapis.com/spanId", span.SpanContext().SpanID().String()).
		Str("method", "PutMessage").
		Str("Params", fmt.Sprintf("%+v", message)).
		Send()

	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(100)
	nameText := uuid.New().String()
	return &pb.Name{Text: nameText, Id: int32(id)}, nil
}

func (n *newServerImplement) PingPong(ctx context.Context, message *pb.Message) (*pb.Message, error) {
	ctx, span := otel.Tracer("main").Start(ctx, "PingPong")
	defer span.End()

	return &pb.Message{Message: "Pong"}, nil
}

func (n *newServerImplement) ListMessage(req *pb.Request, stream pb.Simple_ListMessageServer) error {

	ctx := context.Background()
	ctx, span := n.tracer.Start(ctx, "invoking list message")
	defer span.End()

	log.
		Info().
		Str("logging.googleapis.com/trace", span.SpanContext().TraceID().String()).
		Str("logging.googleapis.com/spanId", span.SpanContext().SpanID().String()).
		Str("method", "PutMessage").
		Str("Params", fmt.Sprintf("%+v", req)).
		Send()

	max := int(req.Number)

	_, span = n.tracer.Start(ctx, "doing list message")

	for n := 0; n < max; n++ {
		result := &pb.Message{Message: fmt.Sprintf("send %d", n)}
		if err := stream.Send(result); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		if sleepSecond > 0 {
			time.Sleep(time.Second * time.Duration(sleepSecond))
		}
	}

	span.End()
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
			break
		}
		log.Info().
			Int("i", i).
			Str("data", fmt.Sprintf("%+v", req.Message)).
			Send()
		results = append(results, req)
	}
	data, err := json.Marshal(results)
	if err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	log.Info().RawJSON("result", data).Send()
	return stream.SendAndClose(&emptypb.Empty{})
}

func main() {
	serverLogger := log.Level(zerolog.TraceLevel)
	grpc_zerolog.ReplaceGrpcLogger(zerolog.New(os.Stderr).Level(zerolog.ErrorLevel))

	tp := tpExporter(projectID, "sample")
	ctx := context.Background()
	defer tp.ForceFlush(ctx)
	otel.SetTracerProvider(tp)

	t := otel.GetTracerProvider().Tracer(domain)

	interceptorOpt := otelgrpc.WithTracerProvider(otel.GetTracerProvider())

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_zerolog.NewPayloadUnaryServerInterceptor(serverLogger),
			grpc_prometheus.UnaryServerInterceptor,
			otelgrpc.UnaryServerInterceptor(interceptorOpt),
		),
		grpc.ChainStreamInterceptor(
			grpc_zerolog.NewStreamServerInterceptor(serverLogger),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zerolog.NewPayloadStreamServerInterceptor(serverLogger),
			otelgrpc.StreamServerInterceptor(interceptorOpt),
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
	newServer.tracer = t

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
		serverLogger.Info().Msgf("prometheus listening on :%s for %s\n", promPort, projectID)
	}()

	reflection.Register(server)
	serverLogger.Info().Msgf("Listening on %s for %s\n", port, projectID)
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
