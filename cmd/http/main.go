package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	pb "github.com/RandomCivil/grpc-practice/helloworld"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc/codes"
)

func parseGRPCBody(req *http.Request) (*pb.HelloRequest, error) {
	// Read the body of the request
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	// The first 5 bytes of the body are the gRPC header, so we skip them
	data := body[5:]

	// Unmarshal the protobuf message
	helloRequest := &pb.HelloRequest{}
	err = proto.Unmarshal(data, helloRequest)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Received: %v\n", helloRequest)

	return helloRequest, nil
}

type myServer struct{}

func (s *myServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the gRPC body
	helloRequest, err := parseGRPCBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the gRPC response
	helloReply := &pb.HelloReply{
		Message: "reply " + helloRequest.GetName(),
	}

	// Marshal the response
	data, err := proto.Marshal(helloReply)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Header().Set("Content-Type", "application/grpc")
	w.Header().Set("Grpc-Status", fmt.Sprintf("%d", codes.OK))
	w.Header().Add("Trailer", "Grpc-Status")
	// w.Header().Add("Trailer", "Grpc-Message")
	// w.Header().Add("Trailer", "Grpc-Status-Details-Bin")

	respHeader := make([]byte, 5) // todo 参考协议文档进行解析
	respHeader[0] = 0

	binary.BigEndian.PutUint32(respHeader[1:], uint32(len(data)))
	if _, err := w.Write(append(respHeader, data...)); err != nil {
		fmt.Printf("响应返回错误:%s", err)
	}
	// w.Write(data)
}

func main() {
	// Create the main listener
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Start the gRPC server
	s := &myServer{}
	h2s := &http2.Server{}
	handler := h2c.NewHandler(s, h2s)
	http.Serve(lis, handler)
}
