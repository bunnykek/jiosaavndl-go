package main

import (
	"fmt"
	"os"
	"strings"
)

var logo = `
           /$$            /$$$$$$                                                         /$$ /$$
          |__/           /$$__  $$                                                       | $$| $$
       /$$ /$$  /$$$$$$ | $$  \__/  /$$$$$$   /$$$$$$  /$$    /$$ /$$$$$$$           /$$$$$$$| $$
      |__/| $$ /$$__  $$|  $$$$$$  |____  $$ |____  $$|  $$  /$$/| $$__  $$ /$$$$$$ /$$__  $$| $$
       /$$| $$| $$  \ $$ \____  $$  /$$$$$$$  /$$$$$$$ \  $$/$$/ | $$  \ $$|______/| $$  | $$| $$
      | $$| $$| $$  | $$ /$$  \ $$ /$$__  $$ /$$__  $$  \  $$$/  | $$  | $$        | $$  | $$| $$
      | $$| $$|  $$$$$$/|  $$$$$$/|  $$$$$$$|  $$$$$$$   \  $/   | $$  | $$        |  $$$$$$$| $$
      | $$|__/ \______/  \______/  \_______/ \_______/    \_/    |__/  |__/         \_______/|__/
 /$$  | $$                                                                                       
|  $$$$$$/                                                                             --by bunny
 \______/                                                                                        
`

func main() {
	jiosaavn := NewJiosaavnClient()
	println(logo)
	args := os.Args
	for _, url := range args {
		if strings.Contains(url, "/album/") {
			matches := Album_song_rx.FindStringSubmatch(url)
			id := matches[2]
			jiosaavn.ProcessAlbum(id)

		} else if strings.Contains(url, "/song/") {
			matches := Album_song_rx.FindStringSubmatch(url)
			id := matches[2]
			jiosaavn.ProcessTrack(id, "", 1, 1, "false")

		} else if strings.Contains(url, "/playlist/") {
			matches := Playlist_rx.FindStringSubmatch(url)
			id := matches[1]
			jiosaavn.ProcessPlaylist(id)
		}
	}
	fmt.Println("Done!")
}
