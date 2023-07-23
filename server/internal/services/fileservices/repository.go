package fileservices

import "github.com/AndreySibrinin/grspSendingFiles/server/internal/models"

type FileRepo interface {
	DownloadFile(fileName string) (*models.File, error)
	UploadFile(*models.File) error
	GetListFiles() ([]string, error)
}
