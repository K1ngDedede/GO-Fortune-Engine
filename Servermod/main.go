package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//Constantes del servidor
const (
	hostname = "localhost"
	port     = "25565"
	protocol = "tcp"
)

//client representa el estado de una conexión que tiene el servidor con un cliente
type client struct {
	id         int
	Cname      string
	Address    string
	connection net.Conn
	reader     bufio.Reader
}

//auxClient representa a un cliente
type auxClient struct {
	Sentfiles int
	Online    bool
}

//simpleFile representa los datos de un archivo enviado por un cliente
type simpleFile struct {
	Fname      string
	Fsize      int
	Channel    string
	Sender     string
	Recipients []string
	Tstamp     time.Time
}

//id de la conexión más reciente al servidor
var bigId int

//mapa de canales del servidor
var channels map[string]map[int]client

//estructura para almacenar datos de los archivos enviados
var fileLog []simpleFile

//estructura para almacenar datos de los clientes
var clientLog map[string]auxClient

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
	//Inicializar log de archivos
	fileLog = make([]simpleFile, 0)
	//Inicializar log de clientes
	clientLog = make(map[string]auxClient)
	//Servir endpoints
	go serve()
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

//Fucnión que registra una nueva conexión al servidor
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
	//Se inicializan los objetos cliente y conexión al cliente
	newClient := client{id: bigId, Cname: string(buffer[:len(buffer)-1]), Address: c.RemoteAddr().String(), connection: c, reader: *bufferReader}
	if aux, ok := clientLog[newClient.Cname]; !ok {
		clientLog[newClient.Cname] = auxClient{Online: true}
	} else {
		aux.Online = true
		clientLog[newClient.Cname] = aux
	}
	bigId += 1
	//Se esperan comandos del nuevo cliente
	go handleConnection(newClient)
}

//Función que se encarga de comunicarse con el cliente por medio de la conexión dada
func handleConnection(c client) {
	//Se obtiene un mensaje del cliente
	bufferReader := c.reader
	buffer, err := bufferReader.ReadBytes('\n')
	//Se revisa si se desconectó el cliente
	if err != nil {
		fmt.Println("Client left.")
		c.connection.Close()
		//Se retiran todas las suscripciones activas del cliente
		go unsubAll(c)
		//Se cierra el proceso de la conexión actual
		return
	}
	//Se procesa el mensaje
	bufferString := string(buffer[:len(buffer)-1])
	bufferString = strings.TrimSuffix(bufferString, "\n")
	bufferFields := strings.Fields(bufferString)
	log.Println(c.Cname+" command:", bufferString)
	//Se actua según el mensaje enviado
	switch bufferFields[0] {
	//suscribirse a un canal
	case "sub":
		subscribeClient(c, bufferFields[1])
	//retirar suscripción a un canal
	case "unsub":
		unsubscribeClient(c, bufferFields[1])
	//subir archivo a un canal
	case "load":
		loadSize, _ := strconv.Atoi(bufferFields[3])
		sendFile(c, bufferFields[1], bufferFields[2], loadSize)
	}
	//Se repite lo anterior hasta que se cierre la conexión
	handleConnection(c)
}

//Función que se encarga de suscribir un cliente a un canal
func subscribeClient(c client, chanName string) {
	if _, ok := channels[chanName]; !ok {
		//Si no existe el canal, se crea
		channels[chanName] = make(map[int]client)
	}
	channels[chanName][c.id] = c
	c.connection.Write([]byte("sub succesful\n"))
	log.Println(c.Cname + " subscribed to " + chanName + " channel.")
}

//Función que se encarga de retirar la suscrpción de un cliente a un canal
func unsubscribeClient(c client, chanName string) {
	delete(channels[chanName], c.id)
	c.connection.Write([]byte("unsub succesful\n"))
	log.Println(c.Cname + " unsubscribed from " + chanName + " channel.")
}

//Función que se encarga de recibir un archivo de un cliente, y enviarlo a todos los clientes suscritos al canal especificado
func sendFile(c client, chanName string, fileName string, fileSize int) {
	if _, ok := channels[chanName]; !ok {
		//Si no existe el canal, se le notifica al cliente
		c.connection.Write([]byte("nochan\n"))
		log.Println("No channel " + chanName)
	} else {
		//Se autoriza el cliente para cargar el archivo
		c.connection.Write([]byte("chan\n"))
		log.Println("Authorized client " + c.Cname + " for loading.")
		//Se lee el archivo enviado por el cliente
		payload := make([]byte, fileSize)
		c.connection.Read(payload)
		//Se guardan datos de la transferencia
		if aux, ok := clientLog[c.Cname]; ok {
			aux.Sentfiles += 1
			clientLog[c.Cname] = aux
		}
		sFile := simpleFile{Fname: fileName, Fsize: fileSize, Channel: chanName, Sender: c.Cname, Tstamp: time.Now(), Recipients: make([]string, 0)}
		//Se envia el archivo a todos los clientes suscritos al canal
		for _, y := range channels[chanName] {
			y.connection.Write([]byte("file " + fileName + " " + strconv.Itoa(fileSize) + "\n"))
			y.connection.Write(payload)
			sFile.Recipients = append(sFile.Recipients, y.Cname)
			log.Println("Sent " + fileName + " to client " + y.Cname + " @" + y.Address)
		}
		fileLog = append(fileLog, sFile)
	}
}

//Función encargada de retirar todas las suscripciones de un cliente
func unsubAll(c client) {
	for _, chann := range channels {
		delete(chann, c.id)
	}
	if aux, ok := clientLog[c.Cname]; ok {
		aux.Online = false
		clientLog[c.Cname] = aux
	}
	log.Println("Unsubscribed client " + c.Cname + " from all channels.")
}

//Función encargada de servir los endpoints del API para el GUI web
func serve() {
	http.HandleFunc("/lmao", hello)
	http.HandleFunc("/chan", chanpoint)
	http.HandleFunc("/files", filepoint)
	http.HandleFunc("/clients", clientpoint)
	http.ListenAndServe(":25566", nil)
}

//endpoint de pruebas
func hello(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		fmt.Fprintf(w, "lmao")
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

//endpoint GET de canales
func chanpoint(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		data, _ := json.Marshal(channels)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(data))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

//endpoint GET de información de archivos
func filepoint(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		data, _ := json.Marshal(fileLog)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(data))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

//endpoint GET de información de clientes
func clientpoint(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		data, _ := json.Marshal(clientLog)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprint(w, string(data))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
