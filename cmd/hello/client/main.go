package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "gitlab.com/love_little_fat_cat/grpc-practice/hello"
	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("127.0.0.1:8000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := &pb.HelloRequest{Name: "小胖猫"}
	r, err := c.SayHello(ctx, req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())

	stream, err := c.LotsOfReplies(ctx, req)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", c, err)
		}
		log.Println(feature)
	}

	stream1, err := c.LotsOfGreetings(ctx)
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", c, err)
	}
	for i := 0; i < 10; i++ {
		if err := stream1.Send(req); err != nil {
			log.Fatalf("%v.Send() = %v", stream1, err)
		}
	}
	reply, err := stream1.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream1, err, nil)
	}
	log.Printf("Route summary: %v", reply)
}
