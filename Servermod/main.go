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

//Constantes del servidor
const (
	hostname = "localhost"
	port     = "25565"
	protocol = "tcp"
)

type client struct {
	id         int
	cname      string
	address    string
	connection net.Conn
	reader     bufio.Reader
	sentfiles  int
}

var bigId int
var channels map[string]map[int]client

func main() {
	fmt.Println("Launching server @ " + hostname + ":" + port)
	//Se abre el socket
	socket, err := net.Listen(protocol, hostname+":"+port)
	//Errores de conexión
	if err != nil {
		fmt.Println("Error launching server: ", err.Error())
		os.Exit(1)
	}
	//Se manda a cerrar el socket de antemano
	defer socket.Close()
	//Inicializar mapa de canales
	channels = make(map[string]map[int]client)

	//Aceptar clientes
	for {
		//Se acepta la conexión
		c, err := socket.Accept()
		//Errores de conexión
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}
		fmt.Println("Client @" + c.RemoteAddr().String() + " connected.")
		//Registrar cliente
		go greetClient(c)
	}
}

func greetClient(c net.Conn) {
	//Recibir información del cliente
	bufferReader := bufio.NewReader(c)
	buffer, err := bufferReader.ReadBytes('\n')
	//Revisar si se desconectó el cliente
	if err != nil {
		fmt.Println("Client @" + c.RemoteAddr().String() + " left.")
		c.Close()
		return
	}
	//Se inicializa el objeto cliente
	newClient := client{id: bigId, cname: string(buffer[:len(buffer)-1]), address: c.RemoteAddr().String(), connection: c, reader: *bufferReader}
	bigId += 1
	go handleConnection(newClient)
}

func handleConnection(c client) {

	bufferReader := c.reader
	buffer, err := bufferReader.ReadBytes('\n')
	//Se revisa si se desconectó el usuario
	if err != nil {
		fmt.Println("Client left.")
		c.connection.Close()
		go unsubAll(c)
		return
	}
	bufferString := string(buffer[:len(buffer)-1])
	bufferString = strings.TrimSuffix(bufferString, "\n")
	bufferFields := strings.Fields(bufferString)
	log.Println(c.cname+" command:", bufferString)
	switch bufferFields[0] {
	case "sub":
		subscribeClient(c, bufferFields[1])
	case "unsub":
		unsubscribeClient(c, bufferFields[1])
	case "load":
		loadSize, _ := strconv.Atoi(bufferFields[3])
		sendFile(c, bufferFields[1], bufferFields[2], loadSize)
	}

	//c.connection.Write(buffer)
	//fmt.Println(channels)
	handleConnection(c)
}

func subscribeClient(c client, chanName string) {
	if _, ok := channels[chanName]; !ok {
		//Si no existe el canal, se crea
		channels[chanName] = make(map[int]client)
	}
	channels[chanName][c.id] = c
	c.connection.Write([]byte("sub succesful\n"))
	log.Println(c.cname + " subscribed to " + chanName + " channel.")
}

func unsubscribeClient(c client, chanName string) {
	delete(channels[chanName], c.id)
	c.connection.Write([]byte("unsub succesful\n"))
	log.Println(c.cname + " unsubscribed from " + chanName + " channel.")
}

func sendFile(c client, chanName string, fileName string, fileSize int) {
	if _, ok := channels[chanName]; !ok {
		//Si no existe el canal, se le notifica al cliente
		c.connection.Write([]byte("nochan\n"))
		log.Println("No channel " + chanName)
	} else {
		c.connection.Write([]byte("chan\n"))
		log.Println("Authorized client " + c.cname + " for loading.")
		//payload, _ := c.reader.ReadBytes('\n')
		payload := make([]byte, fileSize)
		c.connection.Read(payload)
		c.sentfiles += 1
		for _, y := range channels[chanName] {
			y.connection.Write([]byte("file " + fileName + " " + strconv.Itoa(fileSize) + "\n"))
			y.connection.Write(payload)
			log.Println("Sent " + fileName + " to client " + y.cname + " @" + y.address)
		}
	}
}

func unsubAll(c client) {
	for _, chann := range channels {
		delete(chann, c.id)
	}
	log.Println("Unsubscribed client " + c.cname + " from all channels.")
}
