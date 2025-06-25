package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/anatolio-deb/picovpnd/api"
	"github.com/anatolio-deb/picovpnd/auth"
	"github.com/anatolio-deb/picovpnd/core"
	pb "github.com/anatolio-deb/picovpnd/grpc"
	"github.com/anatolio-deb/picovpnd/ip"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	certFile = "/etc/ssl/certs/cert.pem"
	keyFile  = "/etc/ssl/private/key.pem"
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
	ip, err := ip.GetPublicIP()
	if err != nil {
		log.Fatalf("failed to get public IP: %v", err)
	}
	err = auth.GenerateSelfSignedCert(certFile, keyFile, []string{ip})
	if err != nil {
		log.Fatal(err)
	}

	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		log.Fatalf("failed to lookup domain name for IP %s: %v", ip, err)
	} else {
		log.Printf("domain name for IP %s: %s", ip, names[0])
	}

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening on %s", lis.Addr().String())

	m := &autocert.Manager{
		Cache:      autocert.DirCache("certs"),
		Prompt:     autocert.AcceptTOS,
		Email:      os.Getenv("AUTOCERT_EMAIL"),
		HostPolicy: autocert.HostWhitelist(names[0]),
	}
	cert, err := m.GetCertificate(&tls.ClientHelloInfo{
		ServerName: names[0],
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create tls based credential.
	creds := credentials.NewServerTLSFromCert(cert)
	if err != nil {
		log.Fatalf("failed to create credentials: %v", err)
	}

	// s := grpc.NewServer(grpc.Creds(creds), grpc.UnaryInterceptor(auth.HMACAuthInterceptor))
	s := grpc.NewServer(grpc.Creds(creds))

	// Register EchoServer on the server.
	pb.RegisterOpenConnectServiceServer(s, &server{})

	// certPEM, err := os.ReadFile(certFile)
	// if err != nil {
	// 	log.Fatalf("failed to read cert file: %v", err)
	// }

	daemon := api.Daemon{
		Address: names[0],
		Port:    lis.Addr().(*net.TCPAddr).Port,
		CertPEM: cert.Certificate[0],
		// KeyPem:  key,
	}

	go api.RegisterSelf(daemon)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
