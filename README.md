# Go Pirate Bay

Sistema de compartilhamento de arquivos P2P usando Go. Cada peer funciona como cliente e servidor, compartilhando arquivos em uma rede local e utilizando Etcd 
para registrar seu IP, permitindo a descoberta de outros peers. O sistema utiliza sockets para comunicação e SHA1 para verificar a integridade dos arquivos.

# Passos iniciais

## Comando necessarios
```
go mod init github.com/goPirateBay
```

```
protoc --go_out=. --go-grpc_out=. greeter.proto
```

```
go get google.golang.org/grpc
```

## Executar Servico ETCD

```
docker-compose up
```

# Comando para executar servico peer

```
go run cmd/main.go
```

# Estrutura projeto 
```
goPirateBay/
├── client/
│   ├── client.go                # Implementação do cliente
├── server/
│   ├── server.go                # Implementação do servidor
├── cmd/
│   ├── main.go                  # Inicio do sistema, onde está a interface, e start para os servicos necessarios para funcionamento do peer.
├── constants/
│   ├── constants.go             # Principais informacoes para funcionanemnto do sistema
├── greeter/
│   ├── ...                      # Arquivos gerados pelo proto
├── netUtils/
│   ├── netUtils.go              # Pacote que possui funcionalidades uteis referentes configuracoes de rede
├── go.mod                       # Módulo Go para gerenciamento de dependências
├── go.sum                       # Gerenciamento de dependências Go
└── README.md                    # Documentação do projeto
```

# Requisitos Funcionais do Projeto
- Registro no Etcd: Cada peer registra seu IP no Etcd ao iniciar.
- Renovação de Registro: Peers renovam seu registro periodicamente para indicar que ainda estão ativos.
- Listagem de Peers: Os peers podem obter uma lista dos IPs de outros peers registrados no Etcd.
- Cache de Listagem: A listagem de IPs dos peers e dos arquivos disponíveis é mantida em cache para melhorar a performance.
- Download de Arquivos: Peers podem requisitar arquivos uns dos outros, utilizando o hash SHA1 para identificar e baixar os arquivos desejados.
- Verificação de Integridade: Cada arquivo compartilhado é verificado com um hash SHA1 para garantir a integridade.