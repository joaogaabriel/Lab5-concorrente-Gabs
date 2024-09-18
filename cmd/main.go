package main

import (
	"bufio"
	"fmt"
	"github.com/goPirateBay/client"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/fileUtils"
	"github.com/goPirateBay/server"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {

	directory := constants.InitDirFiles
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		createFilesTest(directory)
	}

	cacheServers := &client.ServerCache{}
	cacheFiles := fileUtils.NewFileCache(time.Minute)

	if cacheFiles == nil {
		log.Fatal("fileCache is not initialized")
	}

	cacheFiles.StartPeriodicCacheUpdate(constants.InitDirFiles, 2*time.Minute)

	//go server.StartServer(cacheFiles)
	go server.StartServerr(cacheFiles)
	cacheFiles.GetAllFiles()
	time.Sleep(time.Second / 2)
	for _, file := range cacheFiles.GetAllFiles() {
		fmt.Printf("File: %s, Size: %d bytes HASH: %s\n", file.Name, file.Size, file.SHA1Hash)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	reader := bufio.NewReader(os.Stdin)
	for {
		showMenu()

		fmt.Print("Escolha uma opção: ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		if option == "1" {
			fmt.Print("Digite o hash do arquivo para listar ips dos servidores que possuem: ")
			hash, _ := reader.ReadString('\n')
			hash = strings.TrimSpace(hash)
			validateMachines(hash, cacheServers)

		} else if option == "2" {
			fmt.Print("Digite o nome do arquivo para download: ")
			fileName, _ := reader.ReadString('\n')
			fileName = strings.TrimSpace(fileName)

			fmt.Print("Digite a máquina (IP) para realizar o download: ")
			machine, _ := reader.ReadString('\n')
			machine = strings.TrimSpace(machine)

			downloadFile(machine, fileName)

		} else if option == "3" {
			fmt.Println("Encerrando o servidor gRPC e saindo...")

			//grpcServer.GracefulStop()
			fmt.Println("Servidor gRPC encerrado.")
			break

		} else {
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}

func showMenu() {
	fmt.Println("===== Menu =====")
	fmt.Println("1. Validar quais máquinas possuem o arquivo")
	fmt.Println("2. Realizar download do arquivo")
	fmt.Println("3. Encerrar")
	fmt.Println("================")
}

func validateMachines(hash string, serverChave *client.ServerCache) {
	log.Print("searches servers to file hash: %s", hash)

	listServers := client.ListServerCotainsFile(serverChave, hash)
	if len(listServers) > 0 {
		for _, server := range listServers {
			fmt.Println(server)
		}
	} else {
		fmt.Println("No servers found to contains file hash")
	}
}

func downloadFile(machine string, fileName string) {
	fmt.Printf("Realizando download do arquivo '%s' da máquina '%s'...\n", fileName, machine)
	// Simulação de download do arquivo
	time.Sleep(2 * time.Second)
	fmt.Println("Download concluído!")
}

func createFilesTest(dirPath string) {

	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Println("Directory created:", dirPath)

	files := []string{"file1.txt", "file2.txt", "file3.txt"}

	for _, fileName := range files {
		filePath := filepath.Join(dirPath, fileName)

		file, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("Failed to create fileUtils: %v", err)
		}

		_, err = file.WriteString("This is some netUtils content for " + fileName)
		if err != nil {
			log.Fatalf("Failed to write to fileUtils: %v", err)
		}

		err = file.Close()
		if err != nil {
			log.Fatalf("Failed to close fileUtils: %v", err)
		}

		fmt.Println("File created:", filePath)
	}

	fmt.Println("All files created successfully!")
}
