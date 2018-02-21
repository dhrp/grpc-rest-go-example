package main

import (
	"log"
	"os"

	pb "github.com/dhrp/grpc-rest-go-example/echo-proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	address     = "localhost:8042"
	defaultName = "world"
)

func main() {

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	c := pb.NewEchoServiceClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r1, err := c.Hello(context.Background(), &pb.EchoMessage{Body: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf(r1.Body)
	r2, err := c.Echo(context.Background(), &pb.EchoMessage{Body: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf(r2.Body)

}
