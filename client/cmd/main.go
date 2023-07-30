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
	"strconv"
)

func main() {

	grpcPort, err := strconv.Atoi(getVar("GRPC_PORT_CLIENT", "50050"))

	if err != nil {
		log.Fatalf("GRPC_PORT doesn't look like an integer: %s", err)
	}

	grpcHost := getVar("GRPC_HOST_CLIENT", "0.0.0.0")

	grpcAddr := fmt.Sprintf("%s:%d", grpcHost, grpcPort)

	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

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

	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fileName := filepath.Base(path)

	optimalChunkSize := 512 * 1024

	stream, err := client.UploadFile(context.TODO())

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for i := 0; i < len(content); i += optimalChunkSize {
		end := i + optimalChunkSize
		if end > len(content) {
			end = len(content)
		}

		chunk := content[i:end]

		if err := stream.Send(&v1.FileUploadRequest{FileChunk: chunk, FileName: fileName}); err != nil {
			fmt.Println("Error:", err)
		}
	}

	reply, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}

	log.Printf("Route summary: %v", reply)
}

// Функция для скачивания файла
func downloadFile(client v1.FileUploadServiceClient, scanner *bufio.Scanner) {

	fmt.Println("Enter file name:")
	scanner.Scan()
	fileName := scanner.Text()

	stream, err := client.DownloadFile(context.TODO(), &v1.FileDownloadRequest{FileName: fileName})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fileBytes := make([]byte, 0)

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while streaming %v", err)
		}

		fileBytes = append(fileBytes, resp.GetFileContent()...)
	}

	path := filepath.Join("client", fileName)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0777)

	if err != nil {
		fmt.Println("Error:", err)
	}

	defer file.Close()

	_, err = file.Write(fileBytes)

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

func getVar(key string, fallback string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
