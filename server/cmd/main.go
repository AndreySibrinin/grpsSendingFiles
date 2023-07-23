package main

import (
	"context"
	"fmt"
	v1 "github.com/AndreySibrinin/grspSendingFiles/proto/v1"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/config"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/repo"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/services/fileservices"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	maxUploadCons   = 10
	maxDownloadCons = 10
	maxListCons     = 100
)

var uploadLimiter = make(chan struct{}, maxUploadCons)
var downloadLimiter = make(chan struct{}, maxDownloadCons)
var listLimiter = make(chan struct{}, maxListCons)

func limitHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	switch info.FullMethod {

	case "/v1.FileUploadService/UploadFile":
		uploadLimiter <- struct{}{}
		defer func() { <-uploadLimiter }()

	case "/v1.FileUploadService/DownloadFile":
		downloadLimiter <- struct{}{}
		defer func() { <-downloadLimiter }()

	case "/v1.FileUploadService/GetListFiles":
		listLimiter <- struct{}{}
		defer func() { <-listLimiter }()
	}

	return handler(ctx, req)
}

func main() {

	cnf := config.Init()

	grpcAddr := fmt.Sprintf("%s:%d", cnf.App.GrpcHost, cnf.App.GrpcPort)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	fileRepo := repo.NewFileRepo(cnf.Storage.FilesPath)

	fileService := fileservices.New(fileRepo)

	s := grpc.NewServer(grpc.UnaryInterceptor(limitHandler))

	v1.RegisterFileUploadServiceServer(s, fileService)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
