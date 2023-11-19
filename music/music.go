package music

type MusicPlayer interface {
	PlayMusic()
	PauseMusic()
}

type BeepMusicPlayer struct {
}

func (musicPlayer *BeepMusicPlayer) PlayMusic() {

}

func (MusicPlayer *BeepMusicPlayer) PauseMusic() {

}

func (musicPlayer *BeepMusicPlayer) SendVisualizationsToUI() {

}

func (musicPlayer *BeepMusicPlayer) NextSong() string {
	return ""
}
