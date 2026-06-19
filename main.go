package main

import (
	"log"
	"net"
)

const PORT = ":3000"

func main() {
	listener, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Println(err)
		}
	}(listener)
	log.Println("GoStore started on port ", PORT)

	var store = NewStore()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection: ", err)
			continue
		}

		go HandleConnection(store, conn)
	}
}
