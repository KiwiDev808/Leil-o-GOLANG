package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

const (
	LEILAO_ATIVO     = "ATIVO"
	LEILAO_ENCERRADO = "ENCERRADO"
)

type DbCliente struct {
	Nome  string
	Email string
	Role  string
	Id    string
}

type Message struct {
	Operacao string `json:"operacao"`
	Message  []byte `json:"message"`
}
type MessageCriarCliente struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type MessageEncerrarLeilao struct {
	Id string `json:"id"`
}

type MessageCriarLeilao struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Valor     string `json:"valor"`
}

type MessageRespostaListarLeiloes struct {
	Leiloes []ItemLeilaoCliente `json:"leiloes"`
}

type ItemLeilaoCliente struct {
	Id   string
	Nome string
}
type ItemLeilaoDB struct {
	Id             string
	IdVendedor     string
	Nome           string
	Descricao      string
	EmailApostador string
	ValorAposta    string
	Status         string
}

type Aposta struct {
	EmailApostador string
	Valor          string
}

var itensLeilaoDB []ItemLeilaoDB
var clientes []DbCliente

func main() {
	fmt.Println("Server Running...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	handleError(err, "Error listening:")

	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()

		if err != nil {
			handleError(err, "Error accepting: ")
			connection.Close()
			continue
		}

		fmt.Println("client connected")
		go processClient(connection)
	}
}
func processClient(connection net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Ocorreu um erro na conexão do cliente ou cliente se desconectou: ", err)
		}
	}()
	cliente := handleAuthentication(connection)
	if cliente.Role == "vendedor" {
		go handleVendedor(connection, cliente)
		fmt.Println("Vendedor")
	} else {
		go handleComprador(connection)
		fmt.Println("Comprador")
	}

}

func handleError(err error, message string) {
	if err != nil {
		fmt.Println(message, err.Error())
	}
}

func handleConnectionError(connection net.Conn, err error, message string) {
	if err != nil {
		handleError(err, message)
		connection.Close()
		panic(err)
	}
}

func handleAuthentication(connection net.Conn) DbCliente {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	fmt.Println("Cliente conectado: ", string(buffer[:mLen]))
	handleConnectionError(connection, err, "Error reading")

	var cliente MessageCriarCliente
	json.Unmarshal(buffer[:mLen], &cliente)
	_, err = connection.Write([]byte(cliente.Nome + " você já está conectado e já pode fazer leilões"))
	handleConnectionError(connection, err, "Error writing")
	dbCliente, exists := clienteExiste(cliente)
	if exists {
		fmt.Println("Cliente já existe")
		return dbCliente.(DbCliente)
	}
	fmt.Println("Novo cliente adicionado")

	novoCliente := &DbCliente{
		Nome:  cliente.Nome,
		Email: cliente.Email,
		Role:  cliente.Role,
		Id:    generateRandomId(),
	}
	clientes = append(clientes, *novoCliente)
	return *novoCliente
}

func clienteExiste(cliente MessageCriarCliente) (interface{}, bool) {
	for _, value := range clientes {

		if value.Email == cliente.Email && value.Nome == cliente.Nome {
			return value, true
		}
	}
	return nil, false
}

func handleComprador(connection net.Conn) {
	for {
		message := handleSocketMessage(connection)
		fmt.Println("Received: ", message)
	}
}

func handleVendedor(connection net.Conn, vendedor DbCliente) {
	for {
		message := handleSocketMessage(connection)
		var jsonMsg Message
		json.Unmarshal([]byte(message), &jsonMsg)
		switch jsonMsg.Operacao {
		case "CRIAR_LEILAO":
			var itemLeilao MessageCriarLeilao
			json.Unmarshal(jsonMsg.Message, &itemLeilao)
			itemLeilaoDB := ItemLeilaoDB{
				Id:             generateRandomId(),
				Nome:           itemLeilao.Nome,
				Descricao:      itemLeilao.Descricao,
				IdVendedor:     vendedor.Id,
				EmailApostador: "",
				ValorAposta:    itemLeilao.Valor,
				Status:         LEILAO_ATIVO,
			}
			itensLeilaoDB = append(itensLeilaoDB, itemLeilaoDB)
			message := vendedor.Nome + " o leilao com o item " + itemLeilao.Nome + " foi criado com sucesso"
			sendMessageToClient(connection, message)

		case "ENCERRAR_LEILAO":
			var idLeilao MessageEncerrarLeilao
			json.Unmarshal(jsonMsg.Message, &idLeilao)
			var exists = false
			var foundItem ItemLeilaoDB
			for i, value := range itensLeilaoDB {
				if value.Id == idLeilao.Id && value.IdVendedor == vendedor.Id {
					exists = true
					itensLeilaoDB[i].Status = LEILAO_ENCERRADO
					foundItem = value
				}
			}

			if !exists {
				message := "O leilao com o id " + idLeilao.Id + " não existe"
				sendMessageToClient(connection, message)
			}
			var message string

			if foundItem.EmailApostador != "" {
				message = vendedor.Nome + " o leilao com o item " + foundItem.Nome + " foi encerrado com sucesso e o vencedor foi " + foundItem.EmailApostador + " com o valor " + foundItem.ValorAposta
			} else {
				message = vendedor.Nome + " o leilao com o item " + foundItem.Nome + " foi encerrado com sucesso mas não teve lances"
			}

			sendMessageToClient(connection, message)
		case "LISTAR_LEILOES":
			var leiloesAEnviar []ItemLeilaoCliente
			for i, ild := range itensLeilaoDB {
				if ild.Status == LEILAO_ATIVO && ild.IdVendedor == vendedor.Id {
					leiloesAEnviar = append(leiloesAEnviar, ItemLeilaoCliente{
						Id:   itensLeilaoDB[i].Id,
						Nome: itensLeilaoDB[i].Nome,
					})
				}
			}

			message, _ := json.Marshal(&MessageRespostaListarLeiloes{
				Leiloes: leiloesAEnviar,
			})
			sendMessageToClient(connection, string(message))
		case "SAIR":
			connection.Close()
			return
		default:
			fmt.Println("Operacao nao reconhecida")
			message := "Operação não reconhecida"
			sendMessageToClient(connection, message)
		}
	}
}

func handleSocketMessage(connection net.Conn) string {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	handleConnectionError(connection, err, "Perdemos a conexão com o cliente:")
	fmt.Println("Received: ", string(buffer[:mLen]))
	return string(buffer[:mLen])
}

func generateRandomId() string {
	return strconv.Itoa((rand.Intn(1000000)))
}

func sendMessageToClient(connection net.Conn, message string) {
	_, err := connection.Write([]byte(message))
	handleConnectionError(connection, err, "Error writing")
}
