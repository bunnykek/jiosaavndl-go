package main

type AlbumJson struct {
	Title          string `json:"title"`
	Year           string `json:"year"`
	PrimaryArtists string `json:"primary_artists"`
	Image          string `json:"image"`
	Songs          []struct {
		PermaURL string `json:"perma_url"`
	} `json:"songs"`
}

type SongJson struct {
	ID                string `json:"id"`
	ReleaseDate       string `json:"release_date"`
	Song              string `json:"song"`
	Album             string `json:"album"`
	Year              string `json:"year"`
	Music             string `json:"music"`
	PrimaryArtists    string `json:"primary_artists"`
	Singers           string `json:"singers"`
	Starring          string `json:"starring"`
	Image             string `json:"image"`
	Label             string `json:"label"`
	Language          string `json:"language"`
	CopyrightText     string `json:"copyright_text"`
	ExplicitContent   int    `json:"explicit_content"`
	HasLyrics         string `json:"has_lyrics"`
	EncryptedMediaURL string `json:"encrypted_media_url"`
	PermaURL          string `json:"perma_url"`
}

type PlaylistJson struct {
	Listname  string     `json:"listname"`
	Songs     []SongJson `json:"songs"`
	ListCount string     `json:"list_count"`
}
