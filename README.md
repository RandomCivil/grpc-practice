go install google.golang.org/protobuf/cmd/protoc-gen-go
go install github.com/golang/protobuf/protoc-gen-go

protoc -I=search --go_out=search search/msg.proto
protoc -I=search --go_out=search search/*.proto

go get google.golang.org/grpc
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
protoc --go_out=. --go-grpc_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    search/*.proto
protoc -I=search --go_out=plugins=grpc:search search/*.proto
