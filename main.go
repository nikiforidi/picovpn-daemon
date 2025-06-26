package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/anatolio-deb/picovpnd/api"
	"github.com/anatolio-deb/picovpnd/core"
	pb "github.com/anatolio-deb/picovpnd/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.OpenConnectServiceServer
}

func (s *server) UserAdd(_ context.Context, req *pb.UserAddRequest) (*pb.Response, error) {
	r := &pb.Response{}
	err := core.UserAdd(req.Username, req.Password)
	if err != nil {
		r.Error = err.Error()
	}
	return r, err
}

func (s *server) UserLock(_ context.Context, req *pb.UserLockRequest) (*pb.Response, error) {
	r := &pb.Response{}
	err := core.UserLock(req.Username)
	if err != nil {
		r.Error = err.Error()
	}
	return r, err
}

func (s *server) UserUnlock(_ context.Context, req *pb.UserUnlockRequest) (*pb.Response, error) {
	r := &pb.Response{}
	err := core.UserUnlock(req.Username)
	if err != nil {
		r.Error = err.Error()
	}
	return r, err
}

func (s *server) UserDelete(context.Context, *pb.UserDeleteRequest) (*pb.Response, error) {
	return &pb.Response{
		Error: "Not implemented",
	}, fmt.Errorf("not implemented")
}

func (s *server) UserChangePassword(context.Context, *pb.UserChangePasswordRequest) (*pb.Response, error) {
	return &pb.Response{
		Error: "Not implemented",
	}, fmt.Errorf("not implemented")
}

// https://github.com/grpc/grpc-go/blob/master/examples/features/encryption/TLS/server/main.go
func main() {
	host := os.Getenv("DAEMON_HOST")
	cert := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", host)

	// Create tls based credential.
	creds, err := credentials.NewServerTLSFromFile(
		cert,
		fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem", host),
	)
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	// s := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(auth.HMACAuthInterceptor))
	s := grpc.NewServer(grpc.Creds(creds))

	// Register EchoServer on the server.
	pb.RegisterOpenConnectServiceServer(s, &server{})

	certPEM, err := os.ReadFile(cert)
	if err != nil {
		log.Fatalf("failed to read cert file: %v", err)
	}

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", lis.Addr().String())

	daemon := api.Daemon{
		Address: host,
		Port:    lis.Addr().(*net.TCPAddr).Port,
		CertPEM: certPEM,
		// KeyPem:  key,
	}

	go api.RegisterSelf(daemon)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
