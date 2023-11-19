package transport

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
)

// Resource https://www.youtube.com/watch?v=82oFmY-Qeok && ChatGpt (Go through History)
// Problem is that I dont want to read the whole file into memory, since that will just eat my available RAM
// Instead I want to send the data over in small chunks
// I can have the client read those chunks as they come in
// The first thing the server does is send the size of the file
// After it sends the file, I start to send over the rest of the bytes

var CHUNK_SIZE int = 2000

type TCPFileStreamingServer struct {
	Listener net.Listener
	Address  string
	Users    map[*net.Conn][]string // Mapping of connections to the songs
	quitch   chan struct{}
}

func NewTCPFileStreamingServer(Address string) *TCPFileStreamingServer {
	return &TCPFileStreamingServer{
		Address: Address,
	}
}

func (t *TCPFileStreamingServer) StartServer() error {
	cert, err := tls.LoadX509KeyPair("./certs/cert.pem", "./certs/key.pem")
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	listener, err := net.Listen("tcp", t.Address)
	if err != nil {
		fmt.Println(err)
		return err
	}
	t.Listener = listener

	go t.acceptConnection(config)

	<-t.quitch
	return nil
}

func (t *TCPFileStreamingServer) acceptConnection(config *tls.Config) {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
		}
		// tls handshake
		tlsConn := tls.Server(conn, config)
		songName := make([]byte, CHUNK_SIZE)
		n, err := tlsConn.Read(songName)
		if err != nil {
			fmt.Println(err)
			tlsConn.Close()
		}
		fileName := string(songName[:n])
		fmt.Printf("Got Song Name %s ...\n", fileName)
		go t.streamMusic(tlsConn, fileName)
	}
}

// Stream it to the client
func (t *TCPFileStreamingServer) streamMusic(conn net.Conn, fileName string) {
	var totalStream float64
	defer conn.Close()
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening the mp3")
	}
	defer file.Close()
	byteArr := make([]byte, CHUNK_SIZE)
	fmt.Println("Streaming Over Mp3 Music ..")
	for {
		bytesRead, err := file.Read(byteArr)
		if err != nil {
			fmt.Println("Stream Complete ...")
			return
		}
		n, err := conn.Write(byteArr[:bytesRead])
		totalStream += float64(n)
		// fmt.Println("Total Mega Bytes Sent : ", totalStream/1000000)
		if err != nil {
			conn.Close()
		}
	}
}
