# Test Client

Simple gRPC client for testing fsdriver server connectivity.

## Usage

```bash
# Build
go build -o test_client.exe test_client.go

# Test connection
./test_client.exe [address:port]

# Examples
./test_client.exe 127.0.0.1:50052
./test_client.exe 0.0.0.0:50052
```

## What it does

1. Connects to the fsdriver gRPC server
2. Calls `ReadDir` on the root path
3. Displays the first 5 entries found
4. Shows connection success/failure

This is useful for:
- Testing server connectivity from Windows
- Debugging network issues
- Verifying server is working before mounting with FUSE
