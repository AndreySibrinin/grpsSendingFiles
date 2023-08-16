package fileservices

import (
	"github.com/AndreySibrinin/grspSendingFiles/proto/v1"
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/models"
	"github.com/djherbis/times"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
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

func (s *Service) UploadFile(stream v1.FileUploadService_UploadFileServer) error {

	var fileBytes []byte
	var fileName string
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {

			file := &models.File{
				Name:    fileName,
				Content: fileBytes,
			}

			if err = s.repo.UploadFile(file); err != nil {
				log.Printf("Failed to upload file '%s': %v", file.Name, err)
				return status.Errorf(codes.Internal, "failed to upload file: %v", err)
			}

			log.Printf("File '%s' uploaded successfully", "test.txt")
			return stream.SendAndClose(&v1.FileUploadResponse{
				Message: "File uploaded successfully.",
			})
		}

		if err != nil {
			return status.Errorf(codes.Internal, "error stream: %s", err)
		}

		fileBytes = append(fileBytes, chunk.GetFileChunk()...)
		fileName = chunk.GetFileName()
	}
}

func (s *Service) DownloadFile(req *v1.FileDownloadRequest, stream v1.FileUploadService_DownloadFileServer) error {

	file, err := s.repo.DownloadFile(req.GetFileName())
	if err != nil {
		log.Printf("Failed to download file '%s': %v", req.GetFileName(), err)
		return status.Errorf(codes.Internal, "failed to download file: %v", err)
	}
	optimalChunkSize := 512 * 1024

	for i := 0; i < len(file.Content); i += optimalChunkSize {
		end := i + optimalChunkSize
		if end > len(file.Content) {
			end = len(file.Content)
		}

		chunk := file.Content[i:end]

		if err := stream.Send(&v1.FileDownloadResponse{FileContent: chunk}); err != nil {
			return status.Errorf(codes.Internal, "Failed to send chunk: %v", err)
		}
	}

	log.Printf("File '%s' downloaded successfully", file.Name)
	return nil
}

func (s *Service) GetListFiles(req *v1.ListFilesRequest, stream v1.FileUploadService_GetListFilesServer) error {
	files, err := s.repo.GetListFiles()
	if err != nil {
		log.Printf("Failed to get file list: %v", err)
		return status.Errorf(codes.Internal, "Failed to get file list: %v", err)
	}

	for _, file := range files {
		t, err := times.Stat(file)
		if err != nil {
			log.Printf("Failed to get file times for '%s': %v", file, err)
			return status.Errorf(codes.Internal, "Failed to get file times for '%s': %v", file, err)
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
			return status.Errorf(codes.Internal, "Failed to send file list for '%s': %v", file, err)
		}
	}

	log.Printf("File list sent successfully")
	return nil
}
