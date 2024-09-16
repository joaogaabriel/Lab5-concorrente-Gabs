package main

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	pb "github.com/goPirateBay/greeter"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func downloadFile(client pb.FileServiceClient, fileName string) error {

	req := &pb.FileDownloadRequest{FileName: fileName}
	stream, err := client.Download(context.Background(), req)
	if err != nil {
		return fmt.Errorf("erro ao iniciar o download: %v", err)
	}

	outFile, err := os.Create(constants.DownloadDir + fileName)
	if err != nil {
		return fmt.Errorf("erro ao criar o arquivo: %v", err)
	}
	defer outFile.Close()

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

const (
	broadcastAddr = "0.0.0.0:8000"       // Endereço de broadcast
	bufferSize    = constants.BufferSize // Tamanho do buffer
	timeout       = 20 * time.Second     // Timeout para respostas
)

func discoverServers() []string {
	var servers []string

	// Resolve o endereço de broadcast
	addr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		log.Fatalf("Erro ao resolver o endereço de broadcast: %v", err)
	}

	// Cria a conexão UDP para enviar a mensagem de discovery
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("Erro ao conectar UDP: %v", err)
	}
	defer conn.Close()

	// Envia a mensagem de discovery
	_, err = conn.Write([]byte("DISCOVER"))
	if err != nil {
		log.Fatalf("Erro ao enviar a mensagem de discovery: %v", err)
	}
	fmt.Println("Mensagem de discovery enviada.")

	// Escuta as respostas dos servidores
	conn.SetReadDeadline(time.Now().Add(timeout))
	buffer := make([]byte, bufferSize)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			break // Timeout
		}
		response := string(buffer[:n])
		fmt.Printf("Resposta recebida de %s: %s\n", remoteAddr, response)
		servers = append(servers, remoteAddr.String())
	}

	return servers
}

func main() {
	servers := discoverServers()
	if len(servers) == 0 {
		fmt.Println("Nenhum servidor encontrado.")
	} else {
		fmt.Println("Servidores encontrados:")
		for _, server := range servers {
			fmt.Println(server)
		}
	}
}
