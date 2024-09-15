package main

import (
	"context"
	"fmt"
	pb "github.com/goPirateBay/greeter"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
)

const (
	serverAddr = "localhost:50051" // Endereço do servidor gRPC
)

func downloadFile(client pb.FileServiceClient, fileName string) error {
	// Solicita o arquivo ao servidor
	req := &pb.FileDownloadRequest{FileName: fileName}
	stream, err := client.Download(context.Background(), req)
	if err != nil {
		return fmt.Errorf("erro ao iniciar o download: %v", err)
	}

	// Cria o arquivo local para gravar os dados
	outFile, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("erro ao criar o arquivo: %v", err)
	}
	defer outFile.Close()

	// Recebe os pedaços do arquivo e grava no disco
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break // Fim do arquivo
		}
		if err != nil {
			return fmt.Errorf("erro ao receber pedaço: %v", err)
		}

		// Escreve o pedaço no arquivo local
		_, err = outFile.Write(res.GetChunk())
		if err != nil {
			return fmt.Errorf("erro ao gravar o arquivo: %v", err)
		}
	}

	fmt.Printf("Download do arquivo %s concluído com sucesso!\n", fileName)
	return nil
}

func main() {
	// Conecta ao servidor gRPC
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Falha ao conectar ao servidor gRPC: %v", err)
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	// Solicita o download de um arquivo
	fileName := "file1.txt" // Altere para o nome do arquivo que você deseja baixar
	err = downloadFile(client, fileName)
	if err != nil {
		log.Fatalf("Erro ao baixar o arquivo: %v", err)
	}
}
