package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Sorrow446/go-mp4tag"
	"github.com/mrz1836/go-sanitize"
)

type JiosaavnClient struct {
	client *http.Client
}

func NewJiosaavnClient() *JiosaavnClient {
	return &JiosaavnClient{
		client: &http.Client{},
	}
}

var songApi = "https://www.jiosaavn.com/api.php?__call=webapi.get&token=%v&type=song&_format=json"
var albumApi = "https://www.jiosaavn.com/api.php?__call=webapi.get&token=%v&type=album&_format=json"
var playlistApi = "https://www.jiosaavn.com/api.php?__call=webapi.get&token=%v&type=playlist&_format=json"
var lyricsApi = "https://www.jiosaavn.com/api.php?__call=lyrics.getLyrics&ctx=web6dot0&api_version=4&_format=json&_marker=0%3F_marker%3D0&lyrics_id=%s"
var Album_song_rx, _ = regexp.Compile(`https://www.jiosaavn\.com/(album|song)/.+?/(.+)`)
var Playlist_rx, _ = regexp.Compile(`https://www\.jiosaavn\.com/s/playlist/.+/(.+)`)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (client *JiosaavnClient) tagger(jsondata SongJson, songPath, albumArtist, albumPath string, pos, total int, coverPath string) {
	mp4, err := mp4tag.Open(songPath)
	if err != nil {
		panic(err)
	}
	defer mp4.Close()

	if albumArtist == "" {
		albumArtist = jsondata.PrimaryArtists
	}

	var rating mp4tag.ItunesAdvisory
	if jsondata.ExplicitContent == 0 {
		rating = mp4tag.ItunesAdvisoryClean
	} else {
		rating = mp4tag.ItunesAdvisoryExplicit
	}

	imageFile, _ := os.ReadFile(coverPath)
	imageTag := &mp4tag.MP4Picture{Data: imageFile}

	customTags := map[string]string{
		"Record Label": jsondata.Label,
		"Language":     strings.ToTitle(jsondata.Language),
	}

	// println(customTags["Record Label"])

	if len(jsondata.Singers) > 1 {
		customTags["Singers"] = jsondata.Singers
	}

	if len(jsondata.Starring) > 1 {
		customTags["Starring"] = jsondata.Starring
	}

	tags := &mp4tag.MP4Tags{
		Title:          sanitize.AlphaNumeric(html.UnescapeString(jsondata.Song), true),
		Album:          sanitize.AlphaNumeric(html.UnescapeString(jsondata.Album), true),
		Artist:         sanitize.AlphaNumeric(html.UnescapeString(jsondata.PrimaryArtists), true),
		Composer:       sanitize.AlphaNumeric(html.UnescapeString(jsondata.Music), true),
		AlbumArtist:    sanitize.AlphaNumeric(html.UnescapeString(albumArtist), true),
		Date:           jsondata.ReleaseDate,
		Custom:         customTags,
		Copyright:      jsondata.CopyrightText,
		ItunesAdvisory: rating,
		TrackNumber:    int16(pos),
		TrackTotal:     int16(total),
		Pictures:       []*mp4tag.MP4Picture{imageTag},
	}

	if jsondata.HasLyrics == "true" {
		tags.Lyrics = client.getLyrics(jsondata.ID)
	}

	err = mp4.Write(tags, []string{})
	if err != nil {
		panic(err)
	}
}

func (client *JiosaavnClient) getLyrics(trackId string) string {
	lyricUrl := fmt.Sprintf(lyricsApi, "", "", trackId)
	// fmt.Println(lyricUrl)
	response, err := client.client.Get(lyricUrl)
	handleErr(err)

	bodyBytes, err := io.ReadAll(response.Body)
	handleErr(err)

	// fmt.Println(string(bodyBytes))

	var lyricsJson map[string]string
	err = json.Unmarshal(bodyBytes, &lyricsJson)
	handleErr(err)

	var lyrics string = lyricsJson["lyrics"]
	lyrics = strings.ReplaceAll(lyrics, "<br>", "\n")
	return lyrics
}

func (client *JiosaavnClient) ProcessAlbum(albumId string) {

	albumUrl := fmt.Sprintf(albumApi, albumId)

	response, err := client.client.Get(albumUrl)
	handleErr(err)

	bodyBytes, err := io.ReadAll(response.Body)
	handleErr(err)

	var albumJson AlbumJson
	err = json.Unmarshal(bodyBytes, &albumJson)
	handleErr(err)

	albumInfo := fmt.Sprintf(`
	Album info:
	Album name       : %v
	Album artists    : %v
	Year             : %v
	Number of tracks : %v
	
	`, albumJson.Title, albumJson.PrimaryArtists, albumJson.Year, len(albumJson.Songs))

	fmt.Println(albumInfo)

	var wg sync.WaitGroup
	for i, songInfo := range albumJson.Songs {
		match := Album_song_rx.FindStringSubmatch(songInfo.PermaURL)
		songId := match[2]
		wg.Add(1)
		go func(pos int) {
			client.ProcessTrack(songId, albumJson.PrimaryArtists, pos, len(albumJson.Songs), "false")
			defer wg.Done()
		}(i + 1)
	}
	wg.Wait()
}

func (client *JiosaavnClient) ProcessTrack(songId, albumArtist string, songPos, totalTracks int, isPlaylist string) {
	response, err := client.client.Get(fmt.Sprintf(songApi, songId))
	handleErr(err)

	bodyBytes, err := io.ReadAll(response.Body)
	handleErr(err)

	var tempJson map[string]SongJson
	err = json.Unmarshal(bodyBytes, &tempJson)
	handleErr(err)

	var songJson SongJson
	for _, tempJson := range tempJson {
		songJson = tempJson
		break
	}

	var primaryArtist, trackName, album, year string
	if albumArtist == "" {
		primaryArtist = html.UnescapeString(songJson.PrimaryArtists)
	} else {
		primaryArtist = albumArtist
	}

	trackName = sanitize.AlphaNumeric(html.UnescapeString(songJson.Song), true)
	album = sanitize.AlphaNumeric(html.UnescapeString(songJson.Album), true)
	year = sanitize.AlphaNumeric(html.UnescapeString(songJson.Year), true)

	var folderName, songFileName string

	if isPlaylist != "false" {
		folderName = isPlaylist
	} else {
		var artists = "Various Artists"
		if strings.Count(primaryArtist, ",") < 2 {
			artists = primaryArtist
		}

		folderName = fmt.Sprintf(`%v - %v [%v]`, artists, album, year)
	}

	songFileName = fmt.Sprintf(`%02d. %v.m4a`, songPos, trackName)

	albumPath := filepath.Join("Downloads", folderName)
	songPath := filepath.Join(albumPath, songFileName)
	coverPath := filepath.Join(albumPath, "cover.jpg")

	if isPlaylist != "false" {
		coverPath = filepath.Join(albumPath, songId+".jpg")
	}
	// fmt.Println(songPath)

	_, err = os.Stat(albumPath)
	if err != nil {
		err = os.MkdirAll(albumPath, os.ModePerm)
		handleErr(err)
	}

	_, err = os.Stat(coverPath)
	if err != nil || isPlaylist != "false" {
		// fmt.Println("Downloading the cover...")
		client.download(strings.ReplaceAll(songJson.Image, "150", "500"), coverPath)
	}

	_, err = os.Stat(songPath)
	if err == nil {
		fmt.Printf("%s already downloaded.\n", songFileName)
		return
	}

	fmt.Printf("Downloading: %v...\n", songFileName)
	cdnUrl := client.getCdnUrl(songJson.EncryptedMediaURL)
	client.download(cdnUrl, songPath)
	// println("Tagging metadata...")
	client.tagger(songJson, songPath, albumArtist, albumPath, songPos, totalTracks, coverPath)
	// println("Done!")

	if isPlaylist != "false" {
		os.Remove(coverPath)
	}
}

func (client *JiosaavnClient) download(url, filePath string) {
	out, err := os.Create(filePath)
	handleErr(err)
	defer out.Close()

	resp, err := client.client.Get(url)
	handleErr(err)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	handleErr(err)
}

func (client *JiosaavnClient) getCdnUrl(encurl string) string {
	encurl = url.QueryEscape(encurl)
	cdnApi := fmt.Sprintf("https://www.jiosaavn.com/api.php?__call=song.generateAuthToken&url=%s&bitrate=320&api_version=4&_format=json&ctx=wap6dot0&_marker=0", encurl)
	response, err := client.client.Get(cdnApi)
	handleErr(err)

	bodyBytes, err := io.ReadAll(response.Body)
	handleErr(err)

	// println(string(bodyBytes))
	var cdnJson map[string]string
	err = json.Unmarshal(bodyBytes, &cdnJson)
	handleErr(err)

	return cdnJson["auth_url"]
}

func (client *JiosaavnClient) ProcessPlaylist(playlistId string) {
	// println(playlistId)
	playlistUrl := fmt.Sprintf(playlistApi, playlistId)

	response, err := client.client.Get(playlistUrl)
	handleErr(err)

	bodyBytes, err := io.ReadAll(response.Body)
	handleErr(err)

	var playlistJson PlaylistJson
	err = json.Unmarshal(bodyBytes, &playlistJson)
	handleErr(err)

	playlistName := playlistJson.Listname
	trackCount := playlistJson.ListCount
	playlistFolderName := fmt.Sprintf("Playlist - %s", playlistName)

	playlistInfo := fmt.Sprintf(`
		Playlist Info:
		Playlist name : %s
		Tracks count  : %s
		
		`, playlistName, trackCount)

	fmt.Println(playlistInfo)

	var wg sync.WaitGroup
	for i, songInfo := range playlistJson.Songs {
		match := Album_song_rx.FindStringSubmatch(songInfo.PermaURL)
		songId := match[2]
		wg.Add(1)
		go func(pos int) {
			client.ProcessTrack(songId, "", pos, len(playlistJson.Songs), playlistFolderName)
			defer wg.Done()
		}(i + 1)
	}
	wg.Wait()
}
