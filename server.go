package main

import (
	"fmt"

	"github.com/music-server/transport"
)

// Things I can do, additional processing of songs
// Play specific songs

func main() {
	server := transport.NewTCPFileStreamingServer(":4000")
	fmt.Println("Starting TCP Streaming Server")
	server.StartServer()
}
