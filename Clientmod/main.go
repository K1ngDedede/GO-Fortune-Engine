package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	protocol = "tcp"
)

var channels map[string]struct{}

var synchro chan string

var fileRoute string

func main() {

	//Recibir input del usuario
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Type client name: ")
	cname, _ := reader.ReadString('\n')
	fileRoute = strings.TrimSuffix("files_"+cname, "\n")
	_ = os.Mkdir(fileRoute, 0755)
	fmt.Println("Type server address: ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSuffix(address, "\n")
	//Se conecta al socket del servidor
	fmt.Println("Connecting to " + address)
	socket, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Println("Error connecting to server: ", err.Error())
		os.Exit(1)
	}
	socket.Write([]byte(cname))
	fmt.Println("Connected succesfully to server @" + address)
	channels = make(map[string]struct{})
	synchro = make(chan string)
	//Mensaje inicial
	fmt.Println("Standing by. Type help to recieve a list of commands.")
	//Se escuchan los comandos del cliente
	go clientLoop(*reader, socket)
	//Se escucha al servidor
	socketReader := bufio.NewReader(socket)
	//go serverListener(*socketReader)
	for {
		message, _ := socketReader.ReadBytes('\n')
		serverRelay := strings.Fields(strings.TrimSuffix(string(message), "\n"))
		log.Print("Server relay: " + string(message))
		switch serverRelay[0] {
		case "nochan":
			synchro <- string(serverRelay[0])
		case "chan":
			synchro <- string(serverRelay[0])
		case "file":
			fmt.Println("Downloading file " + serverRelay[1] + " from server.")
			loadSize, _ := strconv.Atoi(serverRelay[2])
			payload := make([]byte, loadSize)
			socket.Read(payload)
			err := os.WriteFile(fileRoute+"/"+serverRelay[1], payload, os.ModePerm)
			if err != nil {
				fmt.Println("Error writing file: " + err.Error())
			}
		}

	}
}

//loop de lógica del cliente
func clientLoop(reader bufio.Reader, socket net.Conn) {
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")
	inputFields := strings.Fields(input)
	switch inputFields[0] {
	//Suscribirse a un canal
	case "sub":
		//Se revisa si ya se está suscrito al canal
		if len(inputFields) == 1 {
			fmt.Println("Not enough arguments.")
		} else if _, ok := channels[inputFields[1]]; ok {
			fmt.Println("Already subscribed to " + inputFields[1] + ".")
		} else {
			socket.Write([]byte(input + "\n"))
			channels[inputFields[1]] = struct{}{}
			fmt.Println("Subscribed to " + inputFields[1] + " succesfully.")
		}
	//Retirar suscripción a un canal
	case "unsub":
		//Se revisa si se está suscrito al canal
		if len(inputFields) == 1 {
			fmt.Println("Not enough arguments.")
		} else if _, ok := channels[inputFields[1]]; ok {
			socket.Write([]byte(input + "\n"))
			delete(channels, inputFields[1])
			fmt.Println("Unsubscribed from " + inputFields[1] + " succesfully.")
		} else {
			fmt.Println("Not subscribed to " + inputFields[1] + ".")
		}
	//Enviar un archivo
	case "load":
		if len(inputFields) <= 2 {
			fmt.Println("Not enough arguments.")
		} else {
			payload, err := os.ReadFile(fileRoute + "/" + inputFields[2])
			if err != nil {
				fmt.Println("Error loading file: " + err.Error())
			} else {
				socket.Write([]byte(input + " " + strconv.Itoa(len(payload)) + "\n"))
				log.Println("Load order for " + inputFields[2] + " sent.")

				if relay := <-synchro; relay == "nochan" {
					fmt.Println("Error: No such channel exists.")
				} else {
					socket.Write(payload)
					log.Println("File " + inputFields[2] + " sent.")
					fmt.Println("File loaded to channel " + inputFields[1])
				}

			}
		}
	//Ver lista de comandos
	case "help":
		fmt.Println("COMMAND LIST \n sub chan - Subscribes to \"chan\" channel.\n unsub chan - Unubscribes from \"chan\" channel.\n load chan file - Loads \"file\" from user's subfolder into \"chan\" channel.\n exit - Exits the application.\n help - Displays information on available commands.")
	case "exit":
		socket.Close()
		os.Exit(0)
		return
	default:
		fmt.Println("Unknown command. Type help to recieve a list of commands.")
	}

	clientLoop(reader, socket)
}
