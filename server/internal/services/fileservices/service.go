package fileservices

import (
	"context"
	"errors"
	v1 "github.com/AndreySibrinin/grspSendingFiles/proto/v1"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/models"
	"github.com/djherbis/times"
	"log"
	"path/filepath"
	"time"
)

type Service struct {
	v1.UnimplementedFileUploadServiceServer
	repo FileRepo
}

func New(fr FileRepo) *Service {
	return &Service{
		repo: fr,
	}
}

func (s *Service) UploadFile(ctx context.Context, req *v1.FileUploadRequest) (*v1.FileUploadResponse, error) {
	file := &models.File{
		Name:    req.GetFileName(),
		Content: req.GetFileContent(),
	}

	if err := s.repo.UploadFile(file); err != nil {
		log.Printf("Failed to upload file '%s': %v", file.Name, err)
		return nil, errors.New("failed to upload file: " + err.Error())
	}

	log.Printf("File '%s' uploaded successfully", file.Name)
	return &v1.FileUploadResponse{Message: "File uploaded successfully."}, nil
}

func (s *Service) GetListFiles(req *v1.ListFilesRequest, stream v1.FileUploadService_GetListFilesServer) error {
	files, err := s.repo.GetListFiles()
	if err != nil {
		log.Printf("Failed to get file list: %v", err)
		return errors.New("failed to get file list: " + err.Error())
	}

	for _, file := range files {
		t, err := times.Stat(file)
		if err != nil {
			log.Printf("Failed to get file times for '%s': %v", file, err)
			return errors.New("failed to get file times: " + err.Error())
		}

		var modTime, createTime time.Time
		modTime = t.ModTime()
		createTime = t.ModTime()

		// In some Unix-like systems, it is not possible to directly obtain the file creation date.
		if t.HasBirthTime() {
			createTime = t.BirthTime()
		}

		res := &v1.ListFilesResponse{
			FileName:   filepath.Base(file),
			DateCreate: modTime.Format("02.01.2006 15:04:05"),
			DateChange: createTime.Format("02.01.2006 15:04:05"),
		}

		if err := stream.Send(res); err != nil {
			log.Printf("Failed to send file list for '%s': %v", file, err)
			return errors.New("failed to send file list: " + err.Error())
		}
	}

	log.Printf("File list sent successfully")
	return nil
}

func (s *Service) DownloadFile(ctx context.Context, req *v1.FileDownloadRequest) (*v1.FileDownloadResponse, error) {
	file, err := s.repo.DownloadFile(req.GetFileName())
	if err != nil {
		log.Printf("Failed to download file '%s': %v", req.GetFileName(), err)
		return nil, errors.New("failed to download file: " + err.Error())
	}

	log.Printf("File '%s' downloaded successfully", file.Name)
	return &v1.FileDownloadResponse{FileContent: file.Content}, nil
}
