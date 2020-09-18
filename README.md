go install google.golang.org/protobuf/cmd/protoc-gen-go
go install github.com/golang/protobuf/protoc-gen-go

protoc -I=search --go_out=search search/msg.proto
protoc -I=search --go_out=search search/*.proto

go get google.golang.org/grpc
protoc -I=search --go_out=plugins=grpc:search search/*.proto
