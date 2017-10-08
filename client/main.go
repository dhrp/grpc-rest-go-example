package main

import (
	"log"
	"os"

	pb "github.com/dhrp/grpc-rest-go-example/echo-proto"
	"github.com/dhrp/grpc-rest-go-example/insecure"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

const (
	address     = "localhost:10000"
	defaultName = "world"
)

func main() {
	keyPair, certPool := insecure.GetCert()
	_ = keyPair

	var opts []grpc.DialOption
	creds := credentials.NewClientTLSFromCert(certPool, "localhost:10000")
	opts = append(opts, grpc.WithTransportCredentials(creds))
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	// client := pb.NewEchoServiceClient(conn)

	// Set up a connection to the server.
	// conn, err := grpc.Dial(address, grpc.WithInsecure())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()
	c := pb.NewEchoServiceClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r, err := c.Echo(context.Background(), &pb.EchoMessage{Value: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Value)
}
