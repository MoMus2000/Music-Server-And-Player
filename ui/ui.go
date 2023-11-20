package ui

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep/speaker"
	"github.com/liamg/gobless"
	"github.com/music-server/transport"
)

var quitChan = make(chan bool)

var listOfClients []*transport.MusicClient
var currentSong string

var lock int = 0

func getVolumeInfo() int {
	cmd := exec.Command("osascript", "-e", "output volume of (get volume settings)")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return -1
	}

	// Process the output to extract the volume information
	volumeInfo := strings.TrimSpace(string(output))

	volumeInfoInt, err := strconv.Atoi(volumeInfo)

	if err != nil {
		return 80
	}

	return volumeInfoInt
}

func cleanSongName(listOfSongs string) string {
	splits := strings.Split(listOfSongs, "/")

	return splits[len(splits)-1]
}

func InitUI(listOfSongs []string) {
	var currentCounter int = 0
	listOfClients = InitClient(listOfSongs)

	gui := gobless.NewGUI()
	if err := gui.Init(); err != nil {
		panic(err)
	}
	defer gui.Close()

	helloTextbox := gobless.NewTextBox()
	var formatString string = "\n"
	for i, song := range listOfSongs {
		if i == 0 {
			formatString += "[X] " + cleanSongName(song) + "\n"
		} else {
			formatString += "[ ] " + cleanSongName(song) + "\n"
		}

	}
	helloTextbox.SetText(`
Sasta Player AKA - The Poor Man's Spotify
` + formatString)
	helloTextbox.SetBorderColor(gobless.ColorGreen)
	helloTextbox.SetTitle("Information")
	helloTextbox.SetTextWrap(true)

	quitTextbox := gobless.NewTextBox()
	quitTextbox.SetText(`Press Ctrl-q to exit.
Press -> to forward to the next song.
Press <- to rollback to the previous song.
Press ↑ ↓ or decrease music volume.
Press Enter to pause.
	`)
	quitTextbox.SetBorderColor(gobless.ColorDarkRed)

	chart := gobless.NewBarChart()
	chart.SetTitle("Volume Level")
	chart.SetYScale(100)
	chart.SetBar("", getVolumeInfo())
	chart.SetBarStyle(gobless.NewStyle(
		gobless.ColorOliveDrab,
		gobless.DefaultStyle.ForegroundColor,
	))

	chart2 := gobless.NewBarChart()
	chart2.SetTitle("Visuals")
	chart2.SetYScale(101)
	chart2.SetBarStyle(gobless.NewStyle(gobless.ColorRed, gobless.ColorWhite))
	chart2.SetBorderColor(gobless.ColorRed)

	rows := []gobless.Component{
		gobless.NewRow(
			gobless.GridSizeThreeQuarters,
			gobless.NewColumn(
				gobless.GridSizeTwoThirds,
				helloTextbox,
			),
			gobless.NewColumn(
				gobless.GridSizeOneThird,
				gobless.NewRow(
					gobless.GridSizeFull,
					gobless.NewColumn(
						gobless.GridSizeFiveSixths,
						chart,
						chart2,
					),
				),
			),
		), gobless.NewRow(
			gobless.GridSizeOneQuarter,
			gobless.NewColumn(
				gobless.GridSizeFull,
				quitTextbox,
			),
		),
	}
	gui.Render(rows...)

	// go checkVolumeRoutine(gui, chart, rows...)
	go visualizeMusic(gui, chart2, rows...)

	gui.HandleKeyPress(gobless.KeyCtrlQ, func(event gobless.KeyPressEvent) {
		gui.Close()
	})

	gui.HandleKeyPress(gobless.KeyEnter, func(event gobless.KeyPressEvent) {
		if lock == 0 {
			speaker.Lock()
			lock = 1
		} else {
			speaker.Unlock()
			lock = 0
		}

	})

	gui.HandleKeyPress(gobless.KeyRight, func(event gobless.KeyPressEvent) {
		helloTextbox.SetText(format(currentCounter, listOfSongs))
		if currentCounter != 0 && currentCounter != len(listOfSongs)-1 {
			if listOfClients[currentCounter-1].Complete == false {
				listOfClients[currentCounter-1].Done <- true
			} else if listOfClients[currentCounter-1].Complete == true {
				listOfClients[currentCounter-1].Complete = false
			}
			go listOfClients[currentCounter].Listen()
			helloTextbox.SetText(format(currentCounter, listOfSongs))
		} else if currentCounter == len(listOfSongs)-1 {
			if listOfClients[currentCounter-1].Complete == false {
				listOfClients[currentCounter-1].Done <- true
			} else if listOfClients[currentCounter-1].Complete == true {
				listOfClients[currentCounter-1].Complete = false
			}
			go listOfClients[currentCounter].Listen()
			helloTextbox.SetText(format(currentCounter, listOfSongs))
			currentCounter = 0
			return
		} else {
			go listOfClients[currentCounter].Listen()
		}
		currentCounter += 1
	})

	gui.HandleKeyPress(gobless.KeyLeft, func(event gobless.KeyPressEvent) {
		helloTextbox.SetText(format(currentCounter, listOfSongs))
		if currentCounter > 0 {
			listOfClients[currentCounter].Done <- true
			go listOfClients[currentCounter-1].Listen()
			helloTextbox.SetText(format(currentCounter, listOfSongs))
			currentCounter -= 1
		}
	})

	gui.HandleKeyPress(gobless.KeyDown, func(event gobless.KeyPressEvent) {
		cmd := exec.Command("osascript", "-e", "set volume output volume (output volume of (get volume settings) - 5)")
		// Run the command and check for errors
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		cmd = exec.Command("osascript", "-e", "output volume of (get volume settings)")
		output, err := cmd.CombinedOutput()

		// Process the output to extract the volume information
		volumeInfo := strings.TrimSpace(string(output))

		volumeInfoInt, err := strconv.Atoi(volumeInfo)

		chart.SetBar("", volumeInfoInt)

		gui.Render(rows...)

	})

	gui.HandleKeyPress(gobless.KeyUp, func(event gobless.KeyPressEvent) {
		cmd := exec.Command("osascript", "-e", "set volume output volume (output volume of (get volume settings) + 5)")
		// Run the command and check for errors
		_, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		cmd = exec.Command("osascript", "-e", "output volume of (get volume settings)")
		output, err := cmd.CombinedOutput()

		// Process the output to extract the volume information
		volumeInfo := strings.TrimSpace(string(output))

		volumeInfoInt, err := strconv.Atoi(volumeInfo)

		chart.SetBar("", volumeInfoInt)

		gui.Render(rows...)

	})

	gui.HandleResize(func(event gobless.ResizeEvent) {
		gui.Render(rows...)
	})

	gui.Loop()
}

// func checkVolumeRoutine(gui *gobless.GUI, chart *gobless.BarChart, rows ...gobless.Component) {
// 	for {
// 		select {
// 		case <-time.After(time.Second * 10):
// 			chart.SetBar("", getVolumeInfo())
// 			gui.Render(rows...)
// 		case <-quitChan:
// 			break
// 		}
// 	}
// }

func InitClient(songCandidates []string) []*transport.MusicClient {
	for i := 0; i < len(songCandidates); i++ {
		song := songCandidates[i]
		client := transport.NewMusicClient("localhost:4000", song)
		listOfClients = append(listOfClients, client)
	}
	return listOfClients
}

func visualizeMusic(gui *gobless.GUI, chart *gobless.BarChart, rows ...gobless.Component) {
	for {
		select {
		case <-time.After(time.Millisecond * 1000):
			chart.SetBar("EU", rand.Intn(101))
			chart.SetBar("NA", rand.Intn(101))
			chart.SetBar("SA", rand.Intn(101))
			chart.SetBar("AS", rand.Intn(101))
			chart.SetBar("NB", rand.Intn(101))
			chart.SetBar("SQ", rand.Intn(101))
			gui.Render(rows...)
		case <-quitChan:
			break
		}
	}
}

func format(currentSong int, listOfSongs []string) string {
	formatString := ""
	for i, song := range listOfSongs {
		if i == currentSong {
			formatString += "[X] " + cleanSongName(song) + "\n"
		} else {
			formatString += "[ ] " + cleanSongName(song) + "\n"
		}

	}
	return `
Sasta Player AKA - The Poor Man's Spotify

` + formatString
}
