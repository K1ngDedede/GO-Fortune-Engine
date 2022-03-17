GO Fortune Engine
----------------------
A simple CLI file server written in Golang as a learning excercise. 
The Clientmod module represents a client application, and the Servermod handles client connections. 
The server runs on localhost:25565 by default.

FortuneUI is a Vue.js GUI to visualize server operations. It consumes the server's API, available at localhost:25566 while the server is running.

Client commands:

sub chan - Subscribe to 'chan' server channel

unsub chan - Unsubscribe from 'chan' server channel

load chan file - Loads 'file' file from the client's file directory into 'chan' channel. The server will send this file to all clients currently subscribed to 'chan'

help - Displays information on commands

exit - Take a wild guess

---------------------
Servidor de archivos de CLI escrito en Golang para aprender.
El módulo Clientmod representa la aplicación del cliente, y Servermod administra las conexiones a los clientes.
El servidor se ejecuta en localhost:25565 por defecto.

FortuneUI es una GUI en Vue.js para visualizar las operaciones del servidor. Consume el API del servidor desde localhost:25566 mientras este se encuentre en ejecución.

Comandos del cliente:

sub chan - Suscribirse al canal 'chan' del servidor

unsub chan - Retirar suscripción del canal 'chan' del servidor

load chan file - Carga el archivo 'file' del directorio de archivos del cliente al canal 'chan' del servidor. El servidor distribuye el archivo a todos los clientes suscritos a 'chan'

help - Mustra información sobre los comandos

exit - Adivine
