package server

import (
	"context"
	"fmt"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/file"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"

	pb "github.com/goPirateBay/greeter"
)

type server struct {
	pb.UnimplementedGreeterServer
}

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
}

const bitsTax = 1024

func (s *FileServiceServer) Download(req *pb.FileDownloadRequest, stream pb.FileService_DownloadServer) error {
	filePath := filepath.Join(constants.InitDirFiles, req.GetFileName())
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("erro ao abrir o arquivo: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, bitsTax)
	for {
		n, err := file.Read(buffer)
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
	log.Printf("Checking if file with SHA-1 hash %s exists", in.Sha1Hash)

	resultChan := make(chan *file.FileInfo)

	go FindFileByHashAsync(in.Sha1Hash, resultChan)

	result := <-resultChan

	close(resultChan)

	if result == nil {
		return &pb.FileExistsResponse{Exists: false}, nil
	}

	return &pb.FileExistsResponse{Exists: true}, nil
}

func FindFileByHashAsync(hash string, resultChan chan<- *file.FileInfo) {

	directory := constants.InitDirFiles

	fileIndex, err := file.ListFilesInDirectory(directory)
	if err != nil {
		log.Printf("Error listing files: %v", err)
		resultChan <- nil
		return
	}

	fileFound := file.FindFileByHash(fileIndex.Files, hash)
	resultChan <- fileFound
}

func StartServer() {
	/*addr := net.UDPAddr{
		Port: constants.BroadcastPort,
		IP:   net.IPv4zero, // Escuta em todas as interfaces (0.0.0.0)
	}*/

	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:8000")
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatalf("Erro ao escutar na porta UDP: %v", err)
	}
	defer conn.Close()

	fmt.Printf("Servidor escutando na porta %d para discovery...\n", constants.BroadcastPort)

	buffer := make([]byte, 1024)
	for {
		// Recebe mensagens de discovery
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Erro ao ler da conexão UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("Mensagem recebida: %s de %s\n", message, remoteAddr)

		// Se for uma mensagem de discovery, responde com o IP do servidor
		if message == "DISCOVER" {
			response := fmt.Sprintf("SERVIDOR:%s", remoteAddr.IP.String())
			_, err = conn.WriteToUDP([]byte(response), remoteAddr)
			if err != nil {
				log.Printf("Erro ao enviar resposta para %s: %v", remoteAddr, err)
			} else {
				fmt.Printf("Resposta enviada para %s\n", remoteAddr)
			}
		}
	}
}
