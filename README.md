# Go Pirate Bay

O projeto é um sistema de comunicação peer-to-peer (P2P) em Go que utiliza Local Peer Discovery para encontrar e conectar peers em uma rede local, facilitando a transferência de arquivos armazenados no diretório /tmp/dataset. O cliente pode se conectar a múltiplos peers simultaneamente para otimizar o desempenho do download. O sistema inclui suporte a reconexão automática, balanceamento de carga, log de operações e verificação de integridade dos arquivos, proporcionando uma solução escalável e eficiente para o compartilhamento de arquivos.

# Passos iniciais
```
go mod init github.com/goPirateBay
```

```
protoc --go_out=. --go-grpc_out=. greeter.proto
```

```
go get google.golang.org/grpc
```
## Comando para executar o serviço do etcd
```
etcd --listen-client-urls=http://0.0.0.0:4001 --advertise-client-urls=http://<meu_ip>:4001
```

# Estrutura projeto previamente estipulada
```
goPirateBay/
├── client/
│   ├── client.go                # Implementação do cliente
│   ├── discovery.go             # Descoberta de servidores ou peers
│   └── main.go                  # Ponto de entrada para execução do cliente
│
├── server/
│   ├── server.go                # Implementação do servidor
│   ├── file_handler.go          # Lógica para buscar arquivos em /tmp/dataset
│   ├── discovery.go             # Descoberta de peers no modo P2P
│   └── main.go                  # Ponto de entrada para execução do servidor
│
├── p2p/
│   ├── peer.go                  # Lógica P2P para comunicação entre peers
│   ├── discovery.go             # Módulo de descoberta P2P (DHT ou multicast)
│   └── connection.go            # Gerenciamento de conexões P2P via sockets ou gRPC
│
├── proto/                       # Definição do Protobuf para gRPC (se necessário)
│   ├── file_service.proto       # Definição do serviço gRPC para busca de arquivos
│   └── peer_service.proto       # Definição do serviço gRPC para descoberta de peers
│
├── scripts/
│   ├── start_client.sh          # Script para iniciar o cliente
│   └── start_server.sh          # Script para iniciar o servidor
│
├── /tmp/dataset/                # Diretório onde os arquivos estão armazenados
│   └── ...                      # Arquivos do dataset buscados pelo cliente
│
├── go.mod                       # Módulo Go para gerenciamento de dependências
├── go.sum                       # Gerenciamento de dependências Go
└── README.md                    # Documentação do projeto
```

# Requisitos Funcionais do Projeto

- [ ] **Descoberta de Peers ou Servidores (Local Peer Discovery)**
  - O sistema deve ser capaz de descobrir automaticamente os peers ou servidores disponíveis na rede local.
  - Deve utilizar o Local Peer Discovery via multicast ou broadcast para encontrar outros peers.

- [ ] **Conexão a Múltiplos Peers**
  - O cliente deve ser capaz de se conectar a múltiplos peers simultaneamente para realizar operações de download ou troca de arquivos.

- [ ] **Transferência de Arquivos entre Peers**
  - O sistema deve permitir a solicitação e transferência de arquivos entre peers.
  - Os arquivos compartilhados estarão no diretório `/tmp/dataset`.

- [ ] **Gerenciamento de Peers**
  - O sistema deve gerenciar a adição de novos peers e a remoção de peers que se desconectaram da rede.

- [ ] **Balanceamento de Carga entre Peers**
  - O sistema deve distribuir as solicitações de arquivos entre múltiplos peers para evitar sobrecarga de um único peer.

- [ ] **Log de Operações**
  - O sistema deve registrar operações importantes como descoberta de peers, conexões, transferências de arquivos e falhas.

- [ ] **Persistência Temporária de Arquivos**
  - Os arquivos devem ser armazenados temporariamente no diretório `/tmp/dataset` e o sistema deve refletir alterações de disponibilidade de arquivos.
