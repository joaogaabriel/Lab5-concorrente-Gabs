package server

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/file"
	"github.com/goPirateBay/netUtils"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	pb "github.com/goPirateBay/greeter"
)

type server struct {
	pb.UnimplementedGreeterServer
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
}

const bitsTax = 1024

func (s *FileServiceServer) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {
	filePath := filepath.Join(constants.InitDirFiles, req.GetFileName())
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, bitsTax)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("erro ao ler o arquivo: %v", err)
		}

		err = stream.Send(&pb.FileDownloadResponse{
			Chunk: buffer[:n],
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

	directory := constants.InitDirFiles

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
	lis, err := net.Listen("tcp", constants.Localhost)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	registerService()

	pb.RegisterGreeterServer(grpcServer, &server{})
	pb.RegisterFileServiceServer(grpcServer, &FileServiceServer{})

	fmt.Println("Server is running on port " + constants.BroadcastPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func registerService() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.IP_ETCD},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to serve Etcd: %v", err)
	}
	defer cli.Close()

	leaseResp, err := cli.Grant(context.Background(), 60) // Tempo de expiração do lease
	if err != nil {
		log.Fatalf("Failed to create lease: %v", err)
	}

	localIp, err := netUtils.GetLocalIP()

	if err != nil {
		fmt.Println(err)
	}

	_, err = cli.Put(context.Background(), "services/"+localIp, localIp, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	log.Println("Registered service with lease successfully!")

	ch, err := cli.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		log.Fatalf("Failed to keep lease alive: %v", err)
	}

	for {
		<-ch
		log.Println("Lease renewed")
	}
}
