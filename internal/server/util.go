package server

import (
	"strings"

	"Shoka/internal/models"
	"Shoka/internal/repository"
)

// ToStructs
// Archives
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

// Artists
func (s *Server) aliasToStruct(list []string) []repository.ArtistAlias {
	aliases := []repository.ArtistAlias{}
	for _, item := range list {
		alias := repository.ArtistAlias{
			Alias: strings.ToLower(item),
		}
		aliases = append(aliases, alias)
	}
	return aliases
}

func (s *Server) groupToStruct(list []string) []models.Group {
	groups := []models.Group{}
	for _, item := range list {
		group := models.Group{
			Name: strings.ToLower(item),
		}
		groups = append(groups, group)
	}
	return groups
}

func (s *Server) linkToStruct(list []string) []models.ArtistLink {
	links := []models.ArtistLink{}
	for _, item := range list {
		link := models.ArtistLink{
			Link: strings.ToLower(item),
		}
		links = append(links, link)
	}
	return links
}
