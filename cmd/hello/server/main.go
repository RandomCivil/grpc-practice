package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	pb "gitlab.com/love_little_fat_cat/grpc-practice/hello"
	"google.golang.org/grpc"
)

const (
	port = ":8000"
)

type GreeterService struct{}

// sayHello implements helloworld.GreeterService.SayHello
func (s *GreeterService) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *GreeterService) LotsOfReplies(in *pb.HelloRequest, stream pb.Greeter_LotsOfRepliesServer) error {
	log.Printf("Received: %v", in.GetName())
	for i := 0; i < 10; i++ {
		r := &pb.HelloReply{Message: in.GetName() + strconv.Itoa(i)}
		if err := stream.Send(r); err != nil {
			return err
		}
	}
	return nil
}

func (s *GreeterService) LotsOfGreetings(stream pb.Greeter_LotsOfGreetingsServer) error {
	for {
		req, err := stream.Recv()
		fmt.Println("LotsOfGreetings recv", req, err)
		if err == io.EOF {
			return stream.SendAndClose(&pb.HelloReply{Message: "aaaa"})
		}
		if err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	//pb.RegisterGreeterService(s, &pb.GreeterService{SayHello: sayHello, LotsOfReplies: lotsOfReplies, LotsOfGreetings: lotsOfGreetings})
	pb.RegisterGreeterServer(s, &GreeterService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
