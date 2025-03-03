package server

import (
	"Shoka/internal/models"
	"strings"
)

// ToStructs
func (s *Server) tagsToStruct(list []string) []models.Tag {
	tags := []models.Tag{}
	for _, item := range list {
		tag := models.Tag{
			Name: strings.ToLower(item),
		}
		tags = append(tags, tag)
	}
	return tags
}

func (s *Server) artistToStruct(list []string) []models.Artist {
	artists := []models.Artist{}
	for _, item := range list {
		artist := models.Artist{
			Name: strings.ToLower(item),
		}
		artists = append(artists, artist)
	}
	return artists
}

func (s *Server) parodyToStruct(list []string) []models.Parody {
	parodies := []models.Parody{}
	for _, item := range list {
		parody := models.Parody{
			Name: strings.ToLower(item),
		}
		parodies = append(parodies, parody)
	}
	return parodies
}

func (s *Server) characterToStruct(list []string) []models.Character {
	chars := []models.Character{}
	for _, item := range list {
		char := models.Character{
			Name: strings.ToLower(item),
		}
		chars = append(chars, char)
	}
	return chars
}

func (s *Server) urlToStruct(list []string) []models.URL {
	urls := []models.URL{}
	for _, item := range list {
		url := models.URL{
			Url: strings.ToLower(item),
		}
		urls = append(urls, url)
	}
	return urls
}
