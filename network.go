package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}(conn)
	log.Println("Client connected from: ", conn.RemoteAddr())

	headerBuffer := make([]byte, 7)
	numberBytesRead, err := io.ReadFull(conn, headerBuffer)

	if errors.Is(err, io.EOF) {
		log.Println("Client disconnected.")
		return
	} else if errors.Is(err, io.ErrUnexpectedEOF) {
		log.Println("Client disconnected half way through sending the header!")
		return
	} else if err != nil {
		log.Println("Network error occurred: ", err)
	}

	log.Printf("Number of Bytes read: %d\n", numberBytesRead)
	log.Printf("Header bytes: %s\n", string(headerBuffer))

	opcode := headerBuffer[0]
	keyLength := binary.BigEndian.Uint16(headerBuffer[1:3])
	valueLength := binary.BigEndian.Uint32(headerBuffer[3:])

	log.Printf("opcode: %d\nkey length: %d\nvalue length: %d\n", opcode, keyLength, valueLength)
}
