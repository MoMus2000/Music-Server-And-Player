package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type MusicClient struct {
	Address  string
	Listener net.Listener
	SongName string
	Pause    chan int
	Done     chan bool
	Complete bool
}

func NewMusicClient(Addr string, SongName string) *MusicClient {
	return &MusicClient{
		Address:  Addr,
		SongName: SongName,
		Pause:    make(chan int),
		Done:     make(chan bool),
		Complete: false,
	}
}

type CustomNet struct {
	conn net.Conn
}

var totalBytes int = 0

func (cn CustomNet) Read(b []byte) (int, error) {
	// fmt.Println("Total Length: ", len(b))
	n, err := cn.conn.Read(b)
	if err != nil {
		return 0, err
	}
	totalBytes += n
	// fmt.Println("Total MB Received: ", float64(totalBytes)/1000000)
	return n, nil
}

func (cn CustomNet) Close() error {
	return cn.conn.Close()
}

type BassBoostEffect struct {
	Streamer beep.Streamer
	Gain     float64
}

func (bb *BassBoostEffect) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = bb.Streamer.Stream(samples)
	if !ok {
		return n, false
	}

	for i := range samples {
		samples[i][0] *= (1 + bb.Gain)
		samples[i][1] *= (1 + bb.Gain)
	}

	return n, true
}

func (bb *BassBoostEffect) Err() error {
	return bb.Streamer.Err()
}

func (client *MusicClient) Listen() {
	conn, err := tls.Dial("tcp", client.Address, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Println("Unable to establish connection with server:", err)
		return
	}
	defer conn.Close()
	_, err = conn.Write([]byte(client.SongName))
	if err != nil {
		fmt.Println(err)
		conn.Close()
		return
	}

	customConn := CustomNet{conn}
	streamer, format, err := mp3.Decode(customConn)
	if err != nil {
		fmt.Println("Error decoding mp3:", err)
		return
	}
	defer streamer.Close()

	// Wrap the existing streamer with the bass boost effect
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	if err != nil {
		fmt.Println("Error initializing speaker:", err)
		return
	}

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		client.Done <- true
	})))

	// Block until done
	<-client.Done

	client.Complete = true
}
