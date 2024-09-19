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

type FileServiceServer struct {
	pb.FileServiceServer
	filesCache *fileUtils.FileCache
}

func (s *FileServiceServer) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {

	log.Println("Starting upload file")

	fileFind, exists := s.filesCache.GetFile(req.GetSha1Hash())

	if !exists {
		return fmt.Errorf("file not found")
	}

	log.Println("opening file to send")
	fileOpen, err := os.Open(fileFind.Path)
	if err != nil {
		return fmt.Errorf("error open file: %v", err)
	}
	defer func(fileOpen *os.File) {
		err := fileOpen.Close()
		if err != nil {

		}
	}(fileOpen)
	log.Println("create channel buffer file to send")
	buffer := make([]byte, constants.BufferSize)
	for {
		n, err := fileOpen.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error read file: %v", err)
		}

		err = stream.Send(&pb.FileDownloadResponse{
			Chunk: buffer[:n],
		})
		if err != nil {
			return fmt.Errorf("error sending part of the file: %v", err)
		}
	}

	return nil
}

func (s *FileServiceServer) CheckExistsFile(_ context.Context, in *pb.FileExistsRequest) (*pb.FileExistsResponse, error) {

	log.Printf("Checking if fileUtils with SHA-1 hash %s exists", in.Sha1Hash)

	_, exists := s.filesCache.GetFile(in.Sha1Hash)
	log.Println("Checking finish")

	if exists {
		return &pb.FileExistsResponse{Exists: true}, nil
	}

	return &pb.FileExistsResponse{Exists: false}, nil
}

func StartServer(filesCache *fileUtils.FileCache) {

	go registerService()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterFileServiceServer(s, &FileServiceServer{filesCache: filesCache})

	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func registerService() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.IpEtcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to serve Etcd: %v", err)
	}

	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {

		}
	}(cli)

	leaseResp, err := cli.Grant(context.Background(), constants.TimeCheckServer)
	if err != nil {
		log.Fatalf("Failed to create lease: %v", err)
	}

	localIp, err := netUtils.GetLocalIP()

	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
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
