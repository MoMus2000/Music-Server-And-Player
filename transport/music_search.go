package transport

import (
	"database/sql"
	"log"
	"math/rand"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	_ "github.com/mattn/go-sqlite3"
)

func dbInit() *sql.DB {
	db, err := sql.Open("sqlite3", "./dummyFiles/db/music.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func SearchMusic(tag string) []string {
	db := dbInit()

	var resultSongs []string

	query := "SELECT songName, tags FROM music"

	results, err := db.Query(query)

	if err != nil {
		log.Fatal("ERROR: Reading from the database", err)
	}

	for results.Next() {
		var songName string
		var tags string
		results.Scan(&songName, &tags)

		splitTags := strings.Split(tags, ",")
		for _, split := range splitTags {
			if fuzzy.Match(tag, split) {
				resultSongs = append(resultSongs, "./dummyFiles/db/"+songName)
				break
			}
		}
	}
	return resultSongs
}

func shuffleMusic(arr []string) {
	n := len(arr)
	for i := n - 1; i > 0; i-- {
		// Generate a random index between 0 and i (inclusive)
		j := rand.Intn(i + 1)

		// Swap arr[i] and arr[j]
		arr[i], arr[j] = arr[j], arr[i]
	}
}
