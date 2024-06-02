package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	pb "github.com/RandomCivil/grpc-practice/helloworld"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
	"google.golang.org/protobuf/proto"
)

var clientPreface = []byte(http2.ClientPreface)

func handleConn(conn *net.TCPConn) {
	defer func() {
		_ = conn.Close()
	}()

	preface := make([]byte, len(clientPreface))
	if _, err := io.ReadFull(conn, preface); err != nil {
		fmt.Println("Error reading preface", err)
		return
	}

	framer := http2.NewFramer(conn, conn)
	frame, err := framer.ReadFrame()
	if err != nil {
		fmt.Println("Error reading frame", err)
		return
	}

	sf, ok := frame.(*http2.SettingsFrame)
	if !ok {
		fmt.Println("Expected settings frame")
		return
	}
	fmt.Println("first Settings frame", sf.String())

	isettings := []http2.Setting{{
		ID:  http2.SettingMaxFrameSize,
		Val: 16384,
	}}
	if err := framer.WriteSettings(isettings...); err != nil {
		fmt.Println("Error writing settings", err)
		return
	}

	for {
		frame, err := framer.ReadFrame()
		if err != nil {
			fmt.Println("Error reading frame", err)
			return
		}
		switch t := frame.(type) {
		case *http2.DataFrame:
			// This is a data frame
			data := t.Data()
			// Now you can parse the data
			fmt.Printf("Data frame: %v\n", string(data))

			reply := &pb.HelloReply{Message: "reply1 " + string(data)}

			data, err := proto.Marshal(reply)
			if err != nil {
				fmt.Println("Error marshaling message:", err)
				return
			}

			// Write the message to the connection
			err = writeGRPCMessage(framer, data)
			if err != nil {
				fmt.Println("Error writing message:", err)
				return
			}
		case *http2.SettingsFrame:
			fmt.Printf("frame loop settings frame: %v\n", t.String())
		case *http2.HeadersFrame:
			fmt.Printf("Header frame: %v\n", frame.Header().String())

		}
	}
}

func writeGRPCMessage(framer *http2.Framer, data []byte) error {
	// Create a gRPC message header
	header := make([]byte, 5)
	header[0] = 0 // compression flag: 0 for uncompressed
	binary.BigEndian.PutUint32(header[1:], uint32(len(data)))

	// Write the header and the message to the connection
	err := framer.WriteData(1, false, append(header, data...))
	if err != nil {
		return err
	}

	// Create the trailing headers
	var buf bytes.Buffer
	enc := hpack.NewEncoder(&buf)

	// Add grpc-status and grpc-message headers
	enc.WriteField(hpack.HeaderField{Name: ":status", Value: "200"})
	enc.WriteField(hpack.HeaderField{Name: "content-type", Value: "application/grpc"})

	enc.WriteField(hpack.HeaderField{Name: "grpc-status", Value: "0"})
	// enc.WriteField(hpack.HeaderField{Name: "grpc-message", Value: ""})

	headers := http2.HeadersFrameParam{
		StreamID:      1,
		EndStream:     true,
		EndHeaders:    true,
		BlockFragment: buf.Bytes(),
	}

	// Write the trailing headers
	err = framer.WriteHeaders(headers)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		conn, err := l.Accept()
		fmt.Println("Accepted connection")
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConn(conn.(*net.TCPConn))
	}
}
