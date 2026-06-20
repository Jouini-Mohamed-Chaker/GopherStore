package store

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

const (
	maxValueLength     = 10 << 20 // 10MB
	connectionDeadline = 30 * time.Second
)

const (
	SET  = 0x01
	GET  = 0x02
	DEL  = 0x03
	PING = 0x04
)

const (
	OK        = 0x00
	ERROR     = 0x01
	NOT_FOUND = 0x02
)

type RequestHeader struct {
	Opcode      byte
	KeyLength   uint16
	ValueLength uint32
}

type ResponseHeader struct {
	Status      byte
	ValueLength uint32
}

func HandleConnection(store *Store, conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}()

	err := conn.SetDeadline(time.Now().Add(connectionDeadline))
	if err != nil {
		LogError(err)
		return
	}

	log.Println("Client connected from: ", conn.RemoteAddr())

	header, err := readHeader(conn)
	if err != nil {
		LogError(err)
		return
	}
	log.Printf("opcode: %d\nkey length: %d\nvalue length: %d\n",
		header.Opcode, header.KeyLength, header.ValueLength)

	keyBuffer, valueBuffer, err := readFullKeyAndValue(conn, header)
	if err != nil {
		LogError(err)
		return
	}

	err = handleRequest(conn, store, header.Opcode, keyBuffer, valueBuffer)
	if err != nil {
		log.Println(err)
		return
	}
}

func LogError(err error) {
	var errValueTooLarge = errors.New("value too large")
	switch {
	case err == nil:
		return
	case errors.Is(err, io.EOF):
		log.Println("Client disconnected.")
	case errors.Is(err, io.ErrUnexpectedEOF):
		log.Println("Client disconnected mid-message.")
	case errors.Is(err, errValueTooLarge):
		log.Println("Client sent oversized value, rejecting.")
	default:
		log.Println(err)
	}
}

func readFullKeyAndValue(r io.Reader, header RequestHeader) ([]byte, []byte, error) {
	if header.ValueLength > maxValueLength {
		return nil, nil, errors.New("value too large")
	}
	keyBuffer := make([]byte, header.KeyLength)
	valueBuffer := make([]byte, header.ValueLength)
	_, err := io.ReadFull(r, keyBuffer)
	if err != nil {
		return nil, nil, err
	}

	_, err = io.ReadFull(r, valueBuffer)
	if err != nil {
		return nil, nil, err
	}

	return keyBuffer, valueBuffer, nil
}

func readHeader(r io.Reader) (RequestHeader, error) {
	var header RequestHeader
	err := binary.Read(r, binary.BigEndian, &header)
	return header, err
}

func handleRequest(w io.Writer, s *Store, opcode byte, key []byte, value []byte) error {
	if opcode == PING {
		return writePong(w)
	}

	keyStr := string(key)
	return handleStorageOperations(w, s, opcode, keyStr, value)
}

func handleStorageOperations(w io.Writer, s *Store, opcode byte, key string, value []byte) error {
	switch opcode {
	case SET:
		return handleSetKeyValue(w, s, key, value)
	case GET:
		return handleGetValue(w, s, key)
	case DEL:
		return handleDeleteValue(w, s, key)
	default:
		return handleUnsupportedOpcode(w)
	}
}

func handleSetKeyValue(w io.Writer, s *Store, key string, value []byte) error {
	s.Set(key, value)
	return writeResponseHeader(w, OK, 0)
}

func handleGetValue(w io.Writer, s *Store, key string) error {
	responseData, exists := s.Get(key)
	if !exists {
		return writeResponseHeader(w, NOT_FOUND, 0)
	}

	err := writeResponseHeader(w, OK, uint32(len(responseData)))
	if err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, responseData)
}

func handleDeleteValue(w io.Writer, s *Store, key string) error {
	_, exists := s.Delete(key)
	if !exists {
		return writeResponseHeader(w, NOT_FOUND, 0)
	}
	return writeResponseHeader(w, OK, 0)
}

func handleUnsupportedOpcode(w io.Writer) error {
	return writeResponseHeader(w, ERROR, 0)
}

func writePong(w io.Writer) error {
	err := writeResponseHeader(w, OK, uint32(len("PONG")))
	if err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, []byte("PONG"))
}

func writeResponseHeader(w io.Writer, status byte, dataLength uint32) error {
	responseHeader := &ResponseHeader{
		Status:      status,
		ValueLength: dataLength,
	}
	return binary.Write(w, binary.BigEndian, responseHeader)
}
