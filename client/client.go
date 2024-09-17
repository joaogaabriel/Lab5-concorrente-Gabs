package client

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	pb "github.com/goPirateBay/greeter"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"log"
	"sync"
	"time"
)

type ServerCache struct {
	mu         sync.RWMutex
	data       []*mvccpb.KeyValue
	expiration time.Time
	ttl        time.Duration
}

func ListServerCotainsFile(sc *ServerCache, hashFile string) []string {
	servers := sc.GetServices()

	done := make(chan string, len(servers))

	for _, serverIP := range servers {
		go checkFileInServer(hashFile, string(serverIP.Value), done)
	}
	if len(sc.data) == 0 {
		fmt.Println("No server has this file")
	}
	for i := 0; i < len(servers); i++ {
		serverIP := <-done
		if serverIP != "" {
			fmt.Printf("Found server: %s\n", serverIP)
		}
	}
	return nil
}

func checkFileInServer(hash string, serverIp string, done chan<- string) {

	conn, err := grpc.Dial(serverIp, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CheckExistsFile(ctx, &pb.FileExistsRequest{Sha1Hash: hash})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	if r.GetExists() {
		done <- serverIp
	} else {
		done <- ""
	}
}

func (sc *ServerCache) GetServices() []*mvccpb.KeyValue {
	sc.mu.RLock()
	if time.Now().Before(sc.expiration) && sc.data != nil {

		fmt.Println("Date to cache...")
		defer sc.mu.RUnlock()
		return sc.data
	}
	sc.mu.RUnlock()

	fmt.Println("Cache expired. Load information to etcd...")
	services := listServices()

	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.data = services
	sc.expiration = time.Now().Add(sc.ttl)

	return sc.data
}

func listServices() []*mvccpb.KeyValue {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.IP_ETCD},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error to connect Etcd: %v", err)
	}
	defer cli.Close()

	resp, err := cli.Get(context.Background(), constants.PrefixNameServerETCP, clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Error to connecct servers: %v", err)
	}

	if resp.Kvs == nil || len(resp.Kvs) == 0 {
		return nil
	}
	return resp.Kvs
}
