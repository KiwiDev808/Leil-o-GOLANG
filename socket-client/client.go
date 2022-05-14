package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/manifoldco/promptui"
)

type Vendedor struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

type ItemLeilao struct {
	Nome      string `json:"nome"`
	Descricao string `json:"descricao"`
	Valor     string `json:"valor"`
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
	for {
		prompt := promptui.Select{
			Label: "Selecione a operação",
			Items: []string{"Iniciar Leilao", "Encerrar Leilao", "Sair"},
		}
		_, result, err := prompt.Run()

		handleError(err, "Erro ao selecionar opção: %v\n")
		handleUserResponse(result, connection)

		fmt.Printf("You choose %q\n", result)
	}
}

func handleUserResponse(response string, connection net.Conn) {
	switch response {
	case "Iniciar Leilao":
		nome, descricao, valorInicial := promptAuctionDetails()
		item, _ := json.Marshal(&ItemLeilao{
			Nome:      nome,
			Descricao: descricao,
			Valor:     valorInicial,
		})
		_, err := connection.Write(item)
		handleError(err, "Erro ao criar leilão: %v\n")
		return
	case "Encerrar Leilao":
		prompt := promptui.Select{
			Label: "Selecione o leilão a encerrar",
			Items: []string{"Fiesta 2005", "NFT Macaco"},
		}
		_, result, err := prompt.Run()

		handleError(err, "Erro ao encerrar leilão: %v\n")
		fmt.Printf("O leilão do item %q foi encerrado\n", result)
		return
	case "Sair":
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
