package server

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/fileUtils"
	"github.com/goPirateBay/netUtils"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/goPirateBay/greeter"
)

type server struct {
	pb.UnimplementedGreeterServer
	filesCache *fileUtils.FileCache
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	filesCache *fileUtils.FileCache
}

func (s *FileServiceServer) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {

	fileFind, exists := s.filesCache.GetFile(req.GetSha1Hash())

	if !exists {
		return fmt.Errorf("File not found")
	}

	fileOpen, err := os.Open(fileFind.Path)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer fileOpen.Close()

	buffer := make([]byte, constants.BufferSize)
	for {
		n, err := fileOpen.Read(buffer)
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

	log.Printf("Checking if fileUtils with SHA-1 hash %s exists", in.Sha1Hash)

	_, exists := s.filesCache.GetFile(in.Sha1Hash)

	if exists {
		return &pb.FileExistsResponse{Exists: true}, nil
	}

	return &pb.FileExistsResponse{Exists: false}, nil
}

func StartServerr(filesCache *fileUtils.FileCache) {

	go registerService()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterGreeterServer(s, &server{filesCache: filesCache})
	pb.RegisterFileServiceServer(s, &FileServiceServer{filesCache: filesCache})

	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
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

	leaseResp, err := cli.Grant(context.Background(), constants.TimeCheckServer) // Tempo de expiração do lease
	if err != nil {
		log.Fatalf("Failed to create lease: %v", err)
	}

	localIp, err := netUtils.GetLocalIP()

	if err != nil {
		fmt.Println(err)
	}

	_, err = cli.Put(context.Background(), constants.PrefixNameServerETCP+localIp, localIp+":"+constants.BroadcastPort, clientv3.WithLease(leaseResp.ID))
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
