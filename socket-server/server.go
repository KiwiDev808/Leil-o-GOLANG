package main

import (
	"encoding/json"
	"fmt"
	"net"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type Cliente struct {
	nome  string
	email string
	role  string
}

var clientes []Cliente

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
	// handleAuthentication(connection)
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			handleError(err, "Error reading:")
			connection.Close()
			return
		}
		fmt.Println("Received: ", string(buffer[:mLen]))
		_, err = connection.Write([]byte("Thanks! Got your message:" + string(buffer[:mLen])))
		if err != nil {
			handleError(err, "Error writing:")
			connection.Close()
			return
		}
	}
}

func handleError(err error, message string) {
	if err != nil {
		fmt.Println(message, err.Error())
	}
}

func handleAuthentication(connection net.Conn) Cliente {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	fmt.Println("Cliente conectado: ", string(buffer[:mLen]))
	handleConnectionError(connection, err, "Error reading")

	var cliente Cliente
	json.Unmarshal(buffer[:mLen], &cliente)
	_, err = connection.Write([]byte(cliente.nome + "você já está conectado e já pode fazer leilões"))
	handleConnectionError(connection, err, "Error writing")
	dbCliente, exists := clienteExiste(cliente)
	if exists {
		return dbCliente.(Cliente)
	}
	clientes = append(clientes, cliente)
	return cliente
}

func handleConnectionError(connection net.Conn, err error, message string) {
	if err != nil {
		handleError(err, message)
		connection.Close()
		panic(err)
	}
}

func clienteExiste(cliente Cliente) (interface{}, bool) {
	for _, value := range clientes {
		if value == cliente {
			return cliente, true
		}
	}

	return nil, false
}
