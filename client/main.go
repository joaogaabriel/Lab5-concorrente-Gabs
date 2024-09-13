package main

import (
	"context"
	"log"
	"os"
	"time"

	pb "github.com/goPirateBay/greeter"
	"google.golang.org/grpc"
)

func main() {
	// Conecta ao servidor gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	// Pega o nome do usuário ou usa "world" como padrão
	hash := "6bb151ddb2b8d27fbf8d7bbd5e975119ea527fd9"
	if len(os.Args) > 1 {
		hash = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Envia a requisição SayHello ao servidor
	r, err := c.CheckExistsFile(ctx, &pb.FileExistsRequest{Sha1Hash: hash})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	log.Printf("Greeting: %s", r.Exists)
}
