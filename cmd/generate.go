package cmd

//go:generate protoc -I ../proto --go_out ../internal/grpc --go-grpc_out ../internal/grpc ../proto/stream.proto
