package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/goPirateBay/client"
	"github.com/goPirateBay/constants"
	"github.com/goPirateBay/fileUtils"
	"github.com/goPirateBay/server"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {

	logsActive := flag.String("logs", "nil", "Ativar logs")
	createFilesToTest := flag.String("create-files-to-test", "nil", "Criar arquivos temporarios para teste")

	flag.Parse()

	if *logsActive != "true" {
		log.SetOutput(ioutil.Discard)
	}

	directory := constants.InitDirFiles
	if *createFilesToTest != "true" {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			createFilesTest(directory)
		}
	}

	cacheServers := &client.ServerCache{}
	cacheFiles := fileUtils.NewFileCache(time.Minute)

	if cacheFiles == nil {
		log.Fatal("fileCache is not initialized")
	}

	go server.StartServer(cacheFiles)

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
			showListFiles(cacheFiles)
		} else if option == "3" {
			fmt.Print("Digite o nome do arquivo para download: ")
			fileName, _ := reader.ReadString('\n')

			fmt.Print("Digite o hash do arquivo para download: ")
			hash, _ := reader.ReadString('\n')

			downloadFile(cacheServers, strings.TrimSpace(hash), strings.TrimSpace(fileName))

		} else if option == "4" {
			fmt.Println("Encerrando o servidor gRPC e saindo...")
			break
		} else {
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}

func showMenu() {
	fmt.Println("===================== MENU =====================")
	fmt.Println("1. Validar quais máquinas possuem o arquivo")
	fmt.Println("2. Listar arquivos locais disponiveis para download")
	fmt.Println("3. Realizar download do arquivo")
	fmt.Println("4. Encerrar")
	fmt.Println("================================================")
}

func showListFiles(cache *fileUtils.FileCache) {
	for _, file := range cache.GetAllFiles() {
		fmt.Printf("File: %s, Size: %d bytes HASH: %s\n", file.Name, file.Size, file.SHA1Hash)
	}
}

func validateMachines(hash string, serverChave *client.ServerCache) {
	log.Printf("searches servers to file hash: %s", hash)

	listServers := client.ListServerCotainsFile(serverChave, hash)
	if len(listServers) > 0 {
		fmt.Println("SERVIDORES ENCONTRADOS PARA ARQUIVO PESQUISADO: ")
		for _, server := range listServers {
			fmt.Println(server)
		}
	} else {
		fmt.Printf("Nenhum servidor encontrado contém hash de arquivo: %s", hash)
	}
}

func downloadFile(sc *client.ServerCache, hash string, fileName string) {
	err := client.DownloadFile(sc, fileName, hash)
	if err != nil {
		fmt.Printf("Houve um erro ao realizar download do arquivo %s", err)
	} else {
		fmt.Println("Download realizado com sucesso!!")
	}

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
