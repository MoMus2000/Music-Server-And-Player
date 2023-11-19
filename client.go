package main

import (
	"flag"

	"github.com/music-server/transport"
	"github.com/music-server/ui"
)

func main() {
	var SongName string
	flag.StringVar(&SongName, "song", "chill", "The name of the song you want to play")
	flag.Parse()
	candidates := transport.SearchMusic(SongName)
	ui.InitUI(candidates)
}
