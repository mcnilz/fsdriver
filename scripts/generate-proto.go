package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// Check if protoc is available
	if _, err := exec.LookPath("protoc"); err != nil {
		fmt.Println("protoc not found. Please install Protocol Buffers compiler:")
		fmt.Println("1. Download from https://github.com/protocolbuffers/protobuf/releases")
		fmt.Println("2. Extract and add to PATH")
		fmt.Println("3. Or use: go run scripts/generate-proto.go")
		os.Exit(1)
	}

	// Generate Go code
	protoFile := "proto/fsdriver.proto"

	// Generate protobuf messages
	cmd := exec.Command("protoc",
		"--go_out=.",
		"--go_opt=paths=source_relative",
		protoFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating protobuf messages: %v\n", err)
		os.Exit(1)
	}

	// Generate gRPC service code
	cmd = exec.Command("protoc",
		"--go-grpc_out=.",
		"--go-grpc_opt=paths=source_relative",
		protoFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating gRPC service: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Proto files generated successfully!")
}
