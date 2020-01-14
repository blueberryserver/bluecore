package bluegrpc

import (
	"net"
	"sync"

	"google.golang.org/grpc"
)

// BLGRPCServer ...
type BLGRPCServer struct {
	_server    *grpc.Server
	_exitChan  chan struct{}   // goroutine exit channel
	_waitGroup *sync.WaitGroup // goroutine wait group
}

// NewGRPCServer ...
func NewGRPCServer() *BLGRPCServer {
	return &BLGRPCServer{
		_server:    grpc.NewServer(),
		_exitChan:  make(chan struct{}), // exit channel
		_waitGroup: &sync.WaitGroup{},   // goroutine wait group
	}
}

// Start ...
func (s *BLGRPCServer) Start(port string) error {
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	err = s._server.Serve(l)
	if err != nil {
		return err
	}
	return nil
}

// GetServer ..
func (s *BLGRPCServer) GetServer() *grpc.Server {
	return s._server
}
