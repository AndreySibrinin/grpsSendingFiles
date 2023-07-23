package repo

import (
	"github.com/AndreySibrinin/grspSendingFiles/server/internal/models"
	"os"
	"path/filepath"
	"sync"
)

type FileRepo struct {
	storage string
	mutex   sync.RWMutex
}

func NewFileRepo(storage string) *FileRepo {
	return &FileRepo{storage: storage}
}

func (f *FileRepo) DownloadFile(fileName string) (*models.File, error) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	filePath := filepath.Join(f.storage, fileName)
	fileContent, err := os.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	return &models.File{Name: fileName, Content: fileContent}, nil
}

func (f *FileRepo) UploadFile(file *models.File) error {

	f.mutex.Lock()
	defer f.mutex.Unlock()

	path := filepath.Join(f.storage, file.Name)

	fileInDir, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		return err
	}

	defer fileInDir.Close()

	_, err = fileInDir.Write(file.Content)

	if err != nil {
		return err
	}

	return nil
}

func (f *FileRepo) GetListFiles() ([]string, error) {

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	path := f.storage

	files, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for _, file := range files {
		if !file.IsDir() {
			result = append(result, filepath.Join(path, file.Name()))
		}
	}
	return result, nil
}
