package main

import (
	"context"
	"log"
	"net"

	pb "github.com/goPirateBay/greeter"

	"google.golang.org/grpc"
)

// Servidor que implementa o serviço Greeter
type server struct {
	pb.UnimplementedGreeterServer
}

// Implementação do método SayHello
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %s", in.GetName())
	return &pb.HelloReply{Message: "Pong"}, nil
}

func main() {
	// Inicia o servidor na porta 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})

	log.Println("Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
