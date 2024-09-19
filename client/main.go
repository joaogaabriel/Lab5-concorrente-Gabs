package main

import (
	"context"
	"log"
	"os"
	"time"
	"fmt"
	"bufio"

	pb "github.com/goPirateBay/greeter"
	"google.golang.org/grpc"
	"github.com/goPirateBay/discovery"
)

func main() {
	// Estabelece a conexão gRPC
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	for {
		// Exibe o menu
		fmt.Println("\n--- Menu Go Pirate Bay ---")
		fmt.Println("1. Descobrir peers")
		fmt.Println("2. Conectar a peers")
		fmt.Println("3. Buscar arquivo")
		fmt.Println("4. Sair")
		fmt.Print("Escolha uma opção: ")

		var choice int
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			// Chama a função para descobrir peers
			fmt.Println("Descobrindo peers...")
			discovery.SendDiscoveryMessage()
			time.Sleep(2 * time.Second)
			go discovery.ListenForDiscovery()

		case 2:
			// Exemplo de conexão a múltiplos peers
			peers := []string{"localhost:50051"} // Exemplo: substituir pelos peers descobertos
			fmt.Println("Conectando a peers...")

			for _, peer := range peers {
				connectToPeer(peer, c)
			}

		case 3:
			// Solicitar arquivo de peers
			fmt.Print("Digite o nome do arquivo para buscar: ")
			reader := bufio.NewReader(os.Stdin)
			filename, _ := reader.ReadString('\n')
			filename = filename[:len(filename)-1] // Remove newline character
			fmt.Printf("Buscando arquivo: %s\n", filename)
			// Aqui você pode adicionar a lógica de busca de arquivo entre peers usando o gRPC

		case 4:
			fmt.Println("Saindo...")
			os.Exit(0)

		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}

// Função para conectar a um peer e realizar a comunicação
func connectToPeer(peer string, c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Envia a requisição SayHello ao servidor
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "ping"})
	if err != nil {
		log.Printf("could not greet peer %s: %v", peer, err)
		return
	}

	log.Printf("Greeting from peer %s: %s", peer, r.GetMessage())
}
