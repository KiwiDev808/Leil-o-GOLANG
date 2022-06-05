package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/manifoldco/promptui"
)

type Vendedor struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ItemLeilao struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Valor     string `json:"valor"`
}

type Message struct {
	Operacao string `json:"operacao"`
	Message  []byte `json:"message"`
}

type ItemLeilaoCliente struct {
	Id            string `json:"id"`
	Nome          string `json:"nome"`
	Descricao     string `json:"descricao"`
	ValorInicial int `json:"valorInicial"`
	ApostaVigente Aposta `json:"apostaVigente"`
}

type Aposta struct {
	EmailApostador string `json:"emailApostador"`
	Valor          int `json:"valor"`
}

type MessageListaDeLeiloes struct {
	Leiloes []ItemLeilaoCliente `json:"leiloes"`
}
type MessageEncerrarLeilao struct {
	Id string `json:"id"`
}

type MessageDarLance struct {
	Id string `json:"id"`
	Valor int `json:"valor"`
}

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

func main() {
	connection, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		panic(err)
	}

	nome, email := promptCredentials()

	vendedor, _ := json.Marshal(&Vendedor{
		Nome:  nome,
		Email: email,
		Role:  "comprador",
	})

	_, err = connection.Write(vendedor)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
	}
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))

	defer connection.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Ocorreu um erro na conexão do cliente ou cliente se desconectou: ", err)
		}
	}()
	for {

		prompt := promptui.Select{
			Label: "Selecione a operação",
			Items: []string{"Listar Artigos", "Dar Lance", "Sair"},
		}
		_, result, err := prompt.Run()

		handleError(err, "Erro ao selecionar opção: %v\n")
		if(err != nil) {
			continue
		}
		handleUserResponse(result, connection)

		fmt.Printf("You choose %q\n", result)
	}
}

func handleUserResponse(response string, connection net.Conn) {
	switch response {
	case "Listar Artigos":
		messageListarLeiloes, _ := json.Marshal(&Message{
			Operacao: "LISTAR_LEILOES",
			Message:  make([]byte, 0),
		})
		sendMessageToServer(connection, messageListarLeiloes, "Erro ao listar leilões: %v\n")
		receivedLeiloesMessage := receiveMessageFromServer(connection)
		var jsonMsg MessageListaDeLeiloes
		json.Unmarshal([]byte(receivedLeiloesMessage), &jsonMsg)
		listaLeiloes := jsonMsg.Leiloes
		if len(listaLeiloes) == 0 {
			fmt.Println("Não há leilões disponíveis")
			return
		}

		for _, leilao := range listaLeiloes {
			var maiorLance string
			maiorLance = "Sem lances"
			if(leilao.ApostaVigente != Aposta{}) {
				maiorLance = strconv.Itoa(leilao.ApostaVigente.Valor)
			} 
			fmt.Printf("Id: %s - Nome: %s - Descricao: %s - Valor Inicial: %v - Maior Lance: %s\n", leilao.Id, leilao.Nome, leilao.Descricao, leilao.ValorInicial ,maiorLance)
		}
		return
	case "Dar Lance":
		messageListarLeiloes, _ := json.Marshal(&Message{
			Operacao: "LISTAR_LEILOES",
			Message:  make([]byte, 0),
		})
		sendMessageToServer(connection, messageListarLeiloes, "Erro ao listar leilões: %v\n")
		receivedLeiloesMessage := receiveMessageFromServer(connection)
		var jsonMsg MessageListaDeLeiloes
		json.Unmarshal([]byte(receivedLeiloesMessage), &jsonMsg)
		listaLeiloes := jsonMsg.Leiloes
		if len(listaLeiloes) == 0 {
			fmt.Println("Não há leilões disponíveis")
			return
		}
		prompt := promptui.Select{
			Label: "Selecione o leilão a dar o lance",
			Items: listaLeiloes,
		}
		i, _, err := prompt.Run()

		handleError(err, "Erro ao selecionar lance: %v\n")
		if(err != nil) {
			return
		}

		promptValorLance := promptui.Prompt{
			Label: "Lance",
		}
	
		valor, err1 := promptValorLance.Run()
		handleError(err1, "Erro ao colocar valor do lance: %v\n")
		if(err1 != nil) {
			return
		}

		valorConvertido, err2 :=strconv.Atoi(valor)
		handleError(err1, "Erro ao converter valor do lance: %v\n")

		if(err2 != nil) {
			return
		}

		lanceLeilao, _ := json.Marshal(&MessageDarLance{
			Valor: int(valorConvertido),
			Id: listaLeiloes[i].Id,
		})

		messageDarLance, _ := json.Marshal(&Message{
			Operacao: "DAR_LANCE",
			Message:  lanceLeilao,
		})
		sendMessageToServer(connection, messageDarLance, "Erro ao dar lance: %v\n")

		receivedDarLanceMessage := receiveMessageFromServer(connection)
		fmt.Print(receivedDarLanceMessage + "\n")
		return
	case "Sair":
		messageEncerrarLeilao, _ := json.Marshal(&Message{
			Operacao: "SAIR",
			Message:  make([]byte, 0),
		})
		sendMessageToServer(connection, messageEncerrarLeilao, "Erro ao encerrar leilão: %v\n")
		os.Exit(0)
	}
}

func promptCredentials() (nome, email string) {
	promptNome := promptui.Prompt{
		Label: "Nome",
	}
	promptEmail := promptui.Prompt{
		Label: "Email",
	}

	nome, err1 := promptNome.Run()
	handleError(err1, "Error reading name: %v\n")

	email, err2 := promptEmail.Run()
	handleError(err2, "Error reading email: %v\n")

	return nome, email
}

func promptAuctionDetails() (nome, descricao string, valorInicial string) {
	promptNome := promptui.Prompt{
		Label: "Name",
	}
	promptDescricao := promptui.Prompt{
		Label: "Descricao",
	}
	promptValor := promptui.Prompt{
		Label: "Valor Inicial",
	}

	nome, err1 := promptNome.Run()
	handleError(err1, "Error reading nome: %v\n")

	email, err2 := promptDescricao.Run()
	handleError(err2, "Error reading email: %v\n")

	descricao, err3 := promptValor.Run()
	handleError(err3, "Error reading valor: %v\n")

	return nome, email, descricao
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Printf(message, err.Error())
		panic(err)
	}
}

func handleConnectionError(connection net.Conn, err error, message string) {
	if err != nil {
		handleError(err, message)
		connection.Close()
		panic(err)
	}
}

func sendMessageToServer(connection net.Conn, message []byte, errorMessage string) {
	_, err := connection.Write(message)
	handleError(err, errorMessage)
}

func receiveMessageFromServer(connection net.Conn) string {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	handleConnectionError(connection, err, "Perdemos a conexão com o servidor")
	return string(buffer[:mLen])
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
