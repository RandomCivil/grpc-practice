package main

import (
	"context"
	"io"
	"log"
	"strconv"
	"time"

	pb "gitlab.com/love_little_fat_cat/grpc-practice/hello"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:8000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.HelloRequest{Name: "小胖猫"}
	r, err := c.SayHello(ctx, req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("SayHello: %s", r.GetMessage())

	stream, err := c.LotsOfReplies(ctx, req)
	if err != nil {
		log.Fatalf("%v.LotsOfReplies(_) = _, %v", c, err)
	}
	for {
		reply, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.LotsOfReplies(_) = _, %v", c, err)
		}
		log.Println("LotsOfReplies", reply)
	}

	stream1, err := c.LotsOfGreetings(ctx)
	if err != nil {
		log.Fatalf("%v.LotsOfGreetings(_) = _, %v", c, err)
	}
	for i := 0; i < 10; i++ {
		r := &pb.HelloRequest{Name: "小胖猫" + strconv.Itoa(i)}
		if err := stream1.Send(r); err != nil {
			log.Fatalf("%v.Send() = %v", stream1, err)
		}
	}
	reply, err := stream1.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream1, err, nil)
	}
	log.Printf("CloseAndRecv: %v", reply)

	stream2, err := c.LotsOfBoth(ctx)
	if err != nil {
		log.Fatalf("%v.LotsOfBoth(_) = _, %v", c, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			reply, err := stream2.Recv()
			if err == io.EOF {
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("%v.LotsOfBoth(_) = _, %v", c, err)
			}
			log.Println("LotsOfBoth", reply)
		}
	}()
	for i := 0; i < 10; i++ {
		r := &pb.HelloRequest{Name: "小胖猫" + strconv.Itoa(i)}
		if err := stream2.Send(r); err != nil {
			log.Fatalf("%v.Send() = %v", stream2, err)
		}
	}
	stream2.CloseSend()
	<-waitc
}
