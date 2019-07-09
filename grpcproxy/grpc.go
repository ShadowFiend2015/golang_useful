package grpcproxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"google.golang.org/grpc"
	log "github.com/sirupsen/logrus"
	
	pb "grpcproxy/proto/grpcproxypb"
)

// GRPCProxy grpc proxy
type GRPCProxy struct {
	proxyAddr string
	client    *http.Client
}

type server struct{}

// New 创建新的grpc proxy
func New(proxyAddr string) *GRPCProxy {
	gr := &GRPCProxy{
		proxyAddr: proxyAddr,
		client:    &http.Client{},
	}
	return gr
}

func (g *GRPCProxy) ListenAndServe(lis net.Listener) error {
	s := grpc.NewServer(
		grpc.CustomCodec(Codec()),
		grpc.UnknownServiceHandler(g.ServeGRPC))
	pb.RegisterGrpcproxyServer(s, &server{})
	// Register reflection service on gRPC server.
	return s.Serve(lis)
}

func RunServer(addr, proxyAddr string) {
	grpcserver := New(proxyAddr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	if err := grpcserver.ListenAndServe(listener); err != nil {
		log.Errorf("http listen error: %v", err)
	}
}
