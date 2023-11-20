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
	Samples  chan [][2]float64
}

func NewMusicClient(Addr string, SongName string) *MusicClient {
	return &MusicClient{
		Address:  Addr,
		SongName: SongName,
		Pause:    make(chan int),
		Done:     make(chan bool),
		Complete: false,
		Samples:  make(chan [][2]float64),
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

type MusicStreamVisualizer struct {
	Streamer beep.Streamer
	Samples  [][2]float64
	Client   *MusicClient
	Framer   chan []float64
}

func (msv MusicStreamVisualizer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = msv.Streamer.Stream(samples)
	if !ok {
		return n, false
	}

	framer := make([]float64, 0)

	for i := range samples {
		framer = append(framer, samples[i][1])
		// 	// leftEar := samples[i][0] // Left Ear
		// 	// fmt.Println(len(msv.Framer))
		// 	// rightEar := samples[i][1] // Right Ear
		// 	// appendFloat64ToFile("musicStream.json", samples[i][:])
	}

	go func() {
		msv.Framer <- framer
	}()

	return n, true
}

func (msv MusicStreamVisualizer) Err() error {
	return msv.Streamer.Err()
}

func (client *MusicClient) Listen(msv *MusicStreamVisualizer) {
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

	// Wrap visualizer over stream
	msv.Streamer = streamer

	speaker.Play(beep.Seq(msv, beep.Callback(func() {
		client.Done <- true
	})))

	// Block until done
	<-client.Done

	client.Complete = true
}
