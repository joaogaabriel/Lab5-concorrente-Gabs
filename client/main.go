package main

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"time"
)

func listServices() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{constants.IP_ETCD},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao Etcd: %v", err)
	}
	defer cli.Close()

	resp, err := cli.Get(context.Background(), "services/", clientv3.WithPrefix())
	if err != nil {
		log.Fatalf("Erro ao listar serviços: %v", err)
	}

	fmt.Println("Servidores disponíveis:")
	for _, ev := range resp.Kvs {
		fmt.Printf("Serviço: %s - Endereço: %s\n", ev.Key, ev.Value)
	}
}

func main() {
	listServices()
}
