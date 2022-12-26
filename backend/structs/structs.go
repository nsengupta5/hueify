package structs

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/zmb3/spotify"
)

type Index interface {
	uint16 | int
}

type RGBRanges struct {
	RedMax, RedMin     uint32
	GreenMax, GreenMin uint32
	BlueMax, BlueMin   uint32
}

type RelatedArtistInfo struct {
	Id         string
	Popularity int32
}

type AlbumRes struct {
	Artist             string                     `json:"artist"`
	ArtistId           spotify.ID                 `json:"artist_id"`
	AlbumImg           string                     `json:"album_image"`
	AlbumName          string                     `json:"album_name"`
	AlbumId            spotify.ID                 `json:"album_id"`
	ImageColors        []prominentcolor.ColorItem `json:"image_colors"`
	RelatedArtists     []string                   `json:"related_artists"`
	RelatedArtistsURIs []string                   `json:"related_artists_uri"`
	NewReq             bool                       `json:"new_request"`
}

type Info struct {
	Related    map[string]Info
	Popularity int
}

type ArtistRelations map[string]Info

type RecommendedAlbumReq struct {
	ColorScheme []prominentcolor.ColorItem `json:"colorScheme"`
	URI         string                     `json:"uri"`
}

type RecommendedAlbum struct {
	Type      string                     `json:"type"`
	Id        string                     `json:"id"`
	Name      string                     `json:"name"`
	Artists   string                     `json:"artists"`
	Image     string                     `json:"image"`
	Colors    []prominentcolor.ColorItem `json:"colors"`
	EndStream bool                       `json:"endStream"`
}
