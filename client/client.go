package client

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	pb "github.com/goPirateBay/greeter"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

type ServerCache struct {
	mu         sync.RWMutex
	data       []*mvccpb.KeyValue
	expiration time.Time
	ttl        time.Duration
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
}

func DownloadFile(sc *ServerCache, fileName string, hash string) error {
	log.Println("Downloading file", fileName)
	servers := sc.GetServices()

	log.Printf("%d servers found to check for file", len(servers))
	totalServers := len(servers)
	ipsFind := make(chan string, totalServers)
	doneChan := make(chan bool, 1)

	for _, serverIP := range servers {
		go checkFileInServer(hash, string(serverIP.Value), ipsFind)
	}

	go func() {
		for ip := range ipsFind {
			if ip != "" {
				log.Printf("Attempting to download file from server %s", ip)

				conn, err := grpc.Dial(ip, grpc.WithInsecure())
				if err != nil {
					log.Printf("Error connecting to server gRPC: %v", err)
					continue
				}
				defer conn.Close()

				client := pb.NewFileServiceClient(conn)
				log.Println("Starting download from server")

				err = downloadFile(client, hash, fileName, doneChan)
				if err != nil {
					doneChan <- false
					log.Printf("Error downloading file from server %s: %v", ip, err)
					continue
				}
				return
			} else {
				doneChan <- false
			}
		}

		log.Println("No server has the requested file.")
		doneChan <- false
	}()
	serversTeste := 0
	for {
		select {
		case success := <-doneChan:
			if success {
				log.Println("File downloaded successfully.")

			}
			if serversTeste == totalServers {
				log.Println("All servers have been tested. None have the file.")
				return fmt.Errorf("file not found on any server")
			}
			log.Println("Download failed or file not found.")
			return fmt.Errorf("download failed or no server has the file")
		}
	}
}
func downloadFile(client pb.FileServiceClient, hash string, fileName string, done chan<- bool) error {
	log.Printf("Starting file download for hash: %s", hash)

	req := &pb.FileDownloadRequest{Sha1Hash: hash}
	stream, err := client.Download(context.Background(), req)
	if err != nil {
		log.Printf("Error starting file download for hash %s: %v", hash, err)
		return fmt.Errorf("error starting download file: %v", err)
	}

	outFile, err := os.Create(constants.InitDirFiles + fileName)
	if err != nil {
		log.Printf("Error creating file %s: %v", fileName, err)
		return fmt.Errorf("error creating file for download: %v", err)
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			log.Printf("Error closing file %s: %v", fileName, err)
		}
	}(outFile)

	log.Printf("Receiving file chunks for %s...", fileName)

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			log.Printf("File download for %s completed.", fileName)
			break
		}
		if err != nil {
			log.Printf("Error receiving chunk for file %s: %v", fileName, err)
			return fmt.Errorf("error receiving chunk: %v", err)
		}

		_, err = outFile.Write(res.GetChunk())
		if err != nil {
			log.Printf("Error writing chunk to file %s: %v", fileName, err)
			return fmt.Errorf("error saving file: %v", err)
		}
	}
	done <- true
	log.Printf("File %s downloaded successfully.", fileName)
	return nil
}

func ListServerCotainsFile(sc *ServerCache, hashFile string) []string {
	servers := sc.GetServices()

	ipsFind := make(chan string, len(servers))

	for _, serverIP := range servers {
		go checkFileInServer(hashFile, string(serverIP.Value), ipsFind)
	}

	var serversCotains []string

	for i := 0; i < len(servers); i++ {
		serverIP := <-ipsFind
		if serverIP != "" {
			serversCotains = append(serversCotains, serverIP)
		}
	}
	return serversCotains
}

func checkFileInServer(hash string, serverIp string, done chan<- string) {

	conn, err := grpc.Dial(serverIp, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	c := pb.NewFileServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CheckExistsFile(ctx, &pb.FileExistsRequest{Sha1Hash: hash})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	if r.GetExists() {
		log.Printf("File %s exists to server %s.", hash, serverIp)
		done <- serverIp
	} else {
		log.Printf("File %s  not exists to server %s.", hash, serverIp)
		done <- ""
	}
}

func (sc *ServerCache) GetServices() []*mvccpb.KeyValue {
	sc.mu.RLock()
	if time.Now().Before(sc.expiration) && sc.data != nil {

		log.Println("Date to cache...")
		defer sc.mu.RUnlock()
		return sc.data
	}
	sc.mu.RUnlock()

	log.Println("Cache expired. Load information to etcd...")
	services := listServices()

	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data = services
	sc.expiration = time.Now().Add(sc.ttl)

	return sc.data
}

func listServices() []*mvccpb.KeyValue {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.IpEtcd},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error to connect Etcd: %v", err)
	}
	defer func(cli *clientv3.Client) {
		err := cli.Close()
		if err != nil {

		}
	}(cli)

	resp, err := cli.Get(context.Background(), constants.PrefixNameServerETCP, clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Error to connecct servers: %v", err)
	}

	if resp.Kvs == nil || len(resp.Kvs) == 0 {
		return nil
	}
	return resp.Kvs
}
