// Package proto provides hand-written gRPC service stubs for the TaskService.
// Since protoc is not available in the build environment, this package
// implements the gRPC codec using JSON encoding instead of protobuf binary.
// When protoc becomes available, regenerate task.pb.go and task_grpc.pb.go
// and remove this codec.go file.
package proto

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/encoding"
)

func init() {
	// Register our JSON codec as the default codec so gRPC uses JSON
	// serialisation for TaskService messages (which are plain Go structs,
	// not protobuf-generated types).
	encoding.RegisterCodec(JSONCodec{})
}

// JSONCodec is a gRPC codec that marshals/unmarshals messages as JSON.
// It satisfies encoding.Codec and is registered under the name "proto"
// so it acts as the default codec for the gRPC connection.
type JSONCodec struct{}

// Name returns "proto" so this codec replaces the default protobuf codec.
func (JSONCodec) Name() string { return "proto" }

// Marshal serialises v to JSON bytes.
func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json codec marshal: %w", err)
	}
	return b, nil
}

// Unmarshal deserialises data from JSON into v.
func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("json codec unmarshal: %w", err)
	}
	return nil
}
