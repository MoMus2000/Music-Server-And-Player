package ui

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/faiface/beep/speaker"
	"github.com/liamg/gobless"
	"github.com/music-server/music"
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

var currentCounter int = 0

func InitUI(listOfSongs []string) {

	listOfClients = InitClient(listOfSongs)

	framer := make(chan []float64, 0)
	msv := transport.MusicStreamVisualizer{Framer: framer}

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
The Poor Man's Spotify
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
	chart2.SetYScale(5)
	chart2.SetBarStyle(gobless.NewStyle(gobless.ColorRed, gobless.ColorWhite))
	chart2.SetBorderColor(gobless.ColorRed)

	rows := []gobless.Component{
		gobless.NewRow(
			gobless.GridSizeThreeQuarters,
			gobless.NewColumn(
				gobless.GridSizeThreeQuarters,
				chart2,
			),
			gobless.NewColumn(
				gobless.GridSizeOneQuarter,
				gobless.NewRow(
					gobless.GridSizeFiveSixths,
					gobless.NewColumn(
						gobless.GridSizeFull,
						// chart,
						helloTextbox,
					),
				),
			),
		), gobless.NewRow(
			gobless.GridSizeOneQuarter,
			gobless.NewColumn(
				gobless.GridSizeFull,
				quitTextbox,
				// helloTextbox,
			),
		),
	}

	gui.Render(rows...)

	// go checkVolumeRoutine(gui, chart, rows...)
	go visualizeMusic(gui, chart2, msv, rows...)

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
		// go func() {
		// 	fmt.Println("INCOMING DATA: ", len(<-msv.Framer))
		// }()
		if currentCounter != 0 && currentCounter != len(listOfSongs)-1 {
			if listOfClients[currentCounter-1].Complete == false {
				listOfClients[currentCounter-1].Done <- true
			} else if listOfClients[currentCounter-1].Complete == true {
				listOfClients[currentCounter-1].Complete = false
			}
			go listOfClients[currentCounter].Listen(&msv)
			helloTextbox.SetText(format(currentCounter, listOfSongs))
		} else if currentCounter == len(listOfSongs)-1 {
			if listOfClients[currentCounter-1].Complete == false {
				listOfClients[currentCounter-1].Done <- true
			} else if listOfClients[currentCounter-1].Complete == true {
				listOfClients[currentCounter-1].Complete = false
			}
			go listOfClients[currentCounter].Listen(&msv)
			helloTextbox.SetText(format(currentCounter, listOfSongs))
			currentCounter = 0
			return
		} else {
			go listOfClients[currentCounter].Listen(&msv)
		}
		currentCounter += 1
	})

	gui.HandleKeyPress(gobless.KeyLeft, func(event gobless.KeyPressEvent) {
		helloTextbox.SetText(format(currentCounter, listOfSongs))
		if currentCounter > 0 {
			listOfClients[currentCounter].Done <- true
			go listOfClients[currentCounter-1].Listen(&msv)
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

func visualizeMusic(gui *gobless.GUI, chart *gobless.BarChart, msv transport.MusicStreamVisualizer, rows ...gobless.Component) {
	for {
		select {
		case <-time.After(time.Nanosecond):
			samples := <-msv.Framer
			desiredBinCount := 64

			// Trim or zero-pad the input signal to the desired length
			if len(samples) > desiredBinCount {
				samples = samples[:desiredBinCount]
			} else if len(samples) < desiredBinCount {
				// Zero-pad if the input signal is shorter than the desired length
				zeroPads := make([]float64, desiredBinCount-len(samples))
				samples = append(samples, zeroPads...)
			}

			var amplitude []float64 = make([]float64, 0)
			y := make([]complex128, len(samples))
			if len(y) > 0 {
				music.Ditfft2(samples, y, len(samples), 1)
				for i := range y {
					window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(len(y)-1))
					y[i] *= complex(window, 0)
				}
				for _, complexNumber := range y {
					// strNumber := strconv.Itoa(i)
					amplitude = append(amplitude, cmplx.Abs(complexNumber)*10000)
					// chart.SetBar(strNumber, int(cmplx.Abs(complexNumber)*1000))
				}
			}
			var MaxVal float64
			for _, val := range amplitude {
				if val > (MaxVal) {
					MaxVal = val
				}
			}
			chart.SetYScale(int(MaxVal * 0.30))
			// sort.Sort(sort.Reverse(sort.Float64Slice(amplitude)))

			first20 := amplitude[:64]
			for i, value := range first20 {
				strNumber := strconv.Itoa(i)
				chart.SetBar("%F{black}"+strNumber, int(value))
			}

			colorWays := []gobless.Color{gobless.ColorRed, gobless.ColorBlueViolet, gobless.ColorRoyalBlue, gobless.ColorDarkRed, gobless.ColorOrangeRed}
			index := rand.Intn(len(colorWays))
			chart.SetBarStyle(gobless.NewStyle(colorWays[index], colorWays[index]))
			chart.SetStyle(gobless.NewStyle(gobless.ColorBlack, gobless.ColorBlack))
			chart.SetBarSpacing(true)
			// chart.SetYScale(rand.Intn(1001))
			// chart.SetBar("EU", len(<-msv.Framer))
			// chart.SetBar("NA", rand.Intn(101))
			// chart.SetBar("SA", rand.Intn(101))
			// chart.SetBar("AS", rand.Intn(101))
			// chart.SetBar("NB", rand.Intn(101))
			// chart.SetBar("SQ", rand.Intn(101))
			// chart.SetBar("NA1", rand.Intn(101))
			// chart.SetBar("SA2", rand.Intn(101))
			// chart.SetBar("AS3", rand.Intn(101))
			// chart.SetBar("NB4", rand.Intn(101))
			// chart.SetBar("SQ52", rand.Intn(101))
			// chart.SetBar("NA12", rand.Intn(101))
			// chart.SetBar("SA22", rand.Intn(101))
			// chart.SetBar("AS32", rand.Intn(101))
			// chart.SetBar("NB42", rand.Intn(101))
			// chart.SetBar("SQ52", rand.Intn(101))
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
The Poor Man's Spotify

` + formatString
}
