package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	pb "github.com/dhrp/grpc-rest-go-example/echo-proto"
	"github.com/dhrp/grpc-rest-go-example/insecure"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

const (
	port     = 10000
	hostname = "localhost"
)

type server struct{}

// implements echo function of EchoServiceServer
func (s *server) Echo(ctx context.Context, in *pb.EchoMessage) (*pb.EchoMessage, error) {
	return &pb.EchoMessage{Value: "Hello xx " + in.Value}, nil
}

func simpleHttpHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is a test endpoint"))
}

func makeGRPCServer(certPool *x509.CertPool) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(certPool, fmt.Sprintf("%v:%d", hostname, port)))}

	//setup grpc server
	s := grpc.NewServer(opts...)
	pb.RegisterEchoServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	return s
}

func getRestMux(certPool *x509.CertPool, opts ...runtime.ServeMuxOption) (*runtime.ServeMux, error) {

	echoEndpoint := "localhost:10000"

	// build context
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: echoEndpoint,
		RootCAs:    certPool,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	gwmux := runtime.NewServeMux()
	err := pb.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, echoEndpoint, dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return nil, err
	}

	return gwmux, nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	keyPair, certPool := insecure.GetCert()

	grpcServer := makeGRPCServer(certPool)
	restMux, _ := getRestMux(certPool)

	// register core mux
	mux := http.NewServeMux()
	mux.HandleFunc("/foobar/", simpleHttpHello)
	mux.Handle("/", restMux)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%v:%d", hostname, port),
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*keyPair},
			NextProtos:   []string{"h2"},
		},
	}

	// start listening on the socket
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// start the server
	// err = srv.Serve(conn)
	fmt.Printf("grpc on port: %d\n", port)

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
