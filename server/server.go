package server

import (
	"context"
	"fmt"
	"github.com/goPirateBay/file"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	pb "github.com/goPirateBay/greeter"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedGreeterServer
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer // Necessário para gRPC
}

func (s *FileServiceServer) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {
	filePath := filepath.Join("/tmp/goPirateBay", req.GetFileName())
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 1024) // Pedaços de 1KB
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break // Fim do arquivo
		}
		if err != nil {
			return fmt.Errorf("erro ao ler o arquivo: %v", err)
		}

		// Envia o pedaço atual ao cliente
		err = stream.Send(&pb.FileDownloadResponse{
			Chunk: buffer[:n], // Envia o pedaço lido
		})
		if err != nil {
			return fmt.Errorf("erro ao enviar o pedaço: %v", err)
		}
	}

	return nil
}

func (s *server) CheckExistsFile(ctx context.Context, in *pb.FileExistsRequest) (*pb.FileExistsResponse, error) {
	log.Printf("Checking if file with SHA-1 hash %s exists", in.Sha1Hash)

	resultChan := make(chan *file.FileInfo)

	go FindFileByHashAsync(in.Sha1Hash, resultChan)

	result := <-resultChan

	close(resultChan)

	if result == nil {
		return &pb.FileExistsResponse{Exists: false}, nil
	}

	return &pb.FileExistsResponse{Exists: true}, nil
}

func FindFileByHashAsync(hash string, resultChan chan<- *file.FileInfo) {
	directory := "/tmp/goPirateBay"

	fileIndex, err := file.ListFilesInDirectory(directory)
	if err != nil {
		log.Printf("Error listing files: %v", err)
		resultChan <- nil
		return
	}

	fileFound := file.FindFileByHash(fileIndex.Files, hash)
	resultChan <- fileFound
}

func StartServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{})
	pb.RegisterFileServiceServer(s, &FileServiceServer{})

	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
