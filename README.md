# grpc-rest-go-example 
In this repository we show an example of how to make a service 
expose both a REST API and a gRPC API *on the same port*.

This repository belongs together with an 
[article](https://medium.com/@thatcher/making-rest-and-grpc-apis-share-one-port-bc0d351f2f84). The focus for 
this example is to get it to work on the same port.

On the [master](https://github.com/dhrp/grpc-rest-go-example/tree/master) 
branch we show an approach using the [GRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway), 
a project to auto-generate HTTP endpoints for gRPC protobuf definitions. 

On the [nogateway](https://github.com/dhrp/grpc-rest-go-example/tree/nogateway) 
branch we show an alternative approach without the auto-code generation, and using 
[CMUX](https://github.com/soheilhy/cmux), a way to share a TCP port between different protocols.
