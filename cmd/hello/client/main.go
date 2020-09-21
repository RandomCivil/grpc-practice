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
}
