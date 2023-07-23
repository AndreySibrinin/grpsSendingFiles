package main

import (
	"bufio"
	"context"
	"fmt"
	v1 "github.com/AndreySibrinin/grspSendingFiles/proto/v1" // импорт сгенерированного кода
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {

	conn, err := grpc.Dial("localhost:50050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := v1.NewFileUploadServiceClient(conn)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Взаимодействие с пользователем

		fmt.Println("Choose action:")
		fmt.Println("1 - Upload file")
		fmt.Println("2 - Download file")
		fmt.Println("3 - Get list files")
		fmt.Println("0 - Exit")

		scanner.Scan()
		switch scanner.Text() {
		case "1":
			uploadFile(client, scanner)
		case "2":
			downloadFile(client, scanner)
		case "3":
			getListFiles(client)
		case "0":
			return

		default:
			fmt.Println("Unknown command")
		}
	}
}

// / Функция для загрузки файла
func uploadFile(client v1.FileUploadServiceClient, scanner *bufio.Scanner) {

	fmt.Println("Enter path to file:")
	//ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*40)
	//defer cancel()
	scanner.Scan()
	path := scanner.Text()

	fileContent, err := os.ReadFile(path)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fileName := filepath.Base(path)

	_, err = client.UploadFile(context.TODO(), &v1.FileUploadRequest{FileContent: fileContent, FileName: fileName})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("UPLOAD DONE")
}

// Функция для скачивания файла
func downloadFile(client v1.FileUploadServiceClient, scanner *bufio.Scanner) {

	fmt.Println("Enter file name:")
	scanner.Scan()
	path := scanner.Text()
	response, err := client.DownloadFile(context.TODO(), &v1.FileDownloadRequest{FileName: path})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fileContent := response.GetFileContent()

	path = filepath.Join("client", path)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		fmt.Println("Error:", err)
	}

	defer file.Close()

	_, err = file.Write(fileContent)

	if err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("Successfully wrote to file")

}

// Функция для получения списка файлов
func getListFiles(client v1.FileUploadServiceClient) {
	log.Printf("Getting the list of files started")

	stream, err := client.GetListFiles(context.TODO(), &v1.ListFilesRequest{})

	if err != nil {
		log.Fatalf("Could not send names: %v", err)
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while streaming %v", err)
		}

		log.Println(message)
	}

	log.Printf("Getting the list of files completed")
}
