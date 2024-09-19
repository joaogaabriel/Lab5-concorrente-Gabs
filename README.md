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

1. O time precisará pensar em um modelo de comunicação. Sockets (como mostramos em sala de aula: https://github.com/thiagomanel/fpc/blob/master/go/clock.go) funcionam. Outros esquemas de comunicação podem ser até mais interessantes (https://grpc.io/docs/languages/go/quickstart/).
 - O modelo de comunicação utilizado funciona via sockets para comunicação, ou através de gRPC.

3. Você precisará pensar em um esquema de organização. Completamente P2P ou cliente-servidor?
 - O projeto utiliza a arquitetura cliente-servidor para a descoberta dos peers na rede, através do serviço de banco multivalorado Etcd. Cada peer atua como servidor e, ao iniciar, precisa se registrar no serviço Etcd, informando seu IP e realizando checagem constante para indicar seu status ativo.

3. Você precisa pensar em um esquema de descoberta. Quais as máquinas que fazem parte do sistema?
 - Com o serviço Etcd, conseguimos acessar todos os IPs dos peers disponíveis na rede.

4. Desempenho continua sendo importante. Pense em minimizar o tempo total, do ponto de vista de um cliente, para obter a lista de máquinas (seus IPs). Embora a lista de otimizações possíveis seja enorme, primeiro FAZ FUNCIONAR! Considere que os arquivos buscados estão em um diretório no /tmp, por exemplo, /tmp/dataset.
 - Foram implementados caches no fluxo de consulta de IPs no serviço Etcd e na consulta dos arquivos disponíveis na máquina. O serviço Etcd foi desenvolvido especificamente para sistemas distribuídos, proporcionando alta eficiência.

5. Você vai escanear todos os arquivos ao receber uma nova requisição? Não deveria manter estado sobre isso? E se, ao invés da lista de máquinas, o programa fizesse download do arquivo? E se, ao invés de baixar de uma única máquina, o programa baixasse partes do arquivo de máquinas diferentes?
 - Os arquivos ficarão em cache por um período de tempo. O sistema também disponibiliza a função de download de arquivos na rede.

6. Um sistema distribuído é um sistema que falha porque uma máquina que você não faz ideia que existe falhou. E se uma máquina com a qual você estiver se comunicando falhar durante a comunicação?
 - Os clientes se comunicam com os servidores que confirmaram ter o arquivo. O cliente tentará realizar o download de um desses servidores, priorizando o primeiro com sucesso. O download só será confirmado quando concluído. Se uma máquina falhar, mas houver outra que possua o arquivo, o download será concluído.

7. E se o sistema permitir entradas de novas máquinas dinamicamente (depois de estar funcionando)? Como descobrir que novas máquinas fazem parte do sistema?
 - Com o auxílio do serviço Etcd, cada máquina se registra com seu IP, permitindo a entrada de novas máquinas na rede sem problemas.