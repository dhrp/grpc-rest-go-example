package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/dhrp/grpc-rest-go-example/certificates"
	pb "github.com/dhrp/grpc-rest-go-example/echo-proto"
)

const (
	port     = 8042
	hostname = "localhost"
)

// join the two constants for convenience
var serveAddress string = fmt.Sprintf("%v:%d", hostname, port)

type server struct{}

// implements echo function of EchoServiceServer
func (s *server) Echo(ctx context.Context, in *pb.EchoMessage) (*pb.EchoMessage, error) {
	return &pb.EchoMessage{Value: "Hello xx " + in.Value}, nil
}

func simpleHTTPHello(w http.ResponseWriter, r *http.Request) {
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

// getRestMux initializes a new multiplexer, and registers each endpoint
// - in this case only the EchoService

func getRestMux(certPool *x509.CertPool, opts ...runtime.ServeMuxOption) (*runtime.ServeMux, error) {

	// Because we run our REST endpoint on the same port as the GRPC the address is the same.
	upstreamGRPCServerAddress := serveAddress

	// get context, this allows control of the connection
	ctx := context.Background()

	// These credentials are for the upstream connection to the GRPC server
	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: upstreamGRPCServerAddress,
		RootCAs:    certPool,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// Which multiplexer to register on.
	gwmux := runtime.NewServeMux()
	err := pb.RegisterEchoServiceHandlerFromEndpoint(ctx, gwmux, upstreamGRPCServerAddress, dopts)
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
	keyPair, certPool := certificates.GetCert()

	grpcServer := makeGRPCServer(certPool)
	restMux, err := getRestMux(certPool)
	if err != nil {
		log.Panic(err)
	}

	// register root Http multiplexer (mux)
	mux := http.NewServeMux()

	// we can add any non-grpc endpoints here.
	mux.HandleFunc("/foobar/", simpleHTTPHello)

	// register the gateway mux onto the root path.
	mux.Handle("/", restMux)

	// the grpcHandlerFunc takes an grpc server and a http muxer and will
	// route the request to the right place at runtime.
	mergeHandler := grpcHandlerFunc(grpcServer, mux)

	// configure TLS for our server. TLS is REQUIRED to make this setup work.
	// check https://golang.org/src/net/http/server.go?s=69823:69872#L2666
	srv := &http.Server{
		Addr:    serveAddress,
		Handler: mergeHandler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*keyPair},
			NextProtos:   []string{"h2"},
		},
	}

	// start listening on the socket
	// Note that if you listen on localhost:<port> you'll not be able to accept
	// connections over the network. Change it to ":port"  if you want it.
	conn, err := net.Listen("tcp", serveAddress)
	if err != nil {
		panic(err)
	}

	// start the server
	fmt.Printf("starting GRPC and REST on: %v\n", serveAddress)
	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
