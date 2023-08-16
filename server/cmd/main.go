package main

import (
	"fmt"
	v1 "github.com/AndreySibrinin/grspSendingFiles/proto/v1"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/config"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/repo"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/services/fileservices"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

func main() {
	
	if err := run(); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func run() error {

	cnf := config.Init()

	grpcAddr := fmt.Sprintf("%s:%d", cnf.App.GrpcHost, cnf.App.GrpcPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	fileRepo := repo.NewFileRepo(cnf.Storage.FilesPath)

	fileService := fileservices.New(fileRepo)

	s := grpc.NewServer()

	v1.RegisterFileUploadServiceServer(s, fileService)

	log.Printf("server listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
