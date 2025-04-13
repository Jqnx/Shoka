package archive

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func Tag(c context.Context, q *repository.Queries, p *models.ArchivePayload, archive *repository.Archive) error {
	if err := q.RemoveTagFromArchive(c, archive.ID); err != nil {
		return err
	}
	for _, item := range p.Tags {
		i := strings.ToLower(item)
		tag, _ := q.GetTag(c, i)

		if tag.Tag == i {
			if err := q.AddTagToArchive(c, repository.AddTagToArchiveParams{
				ArchiveID: archive.ID,
				TagID:     tag.ID,
			}); err != nil {
				return err
			}
		} else {
			tag, err := q.CreateTag(c, i)
			if err != nil {
				return err
			}
			if err := q.AddTagToArchive(c, repository.AddTagToArchiveParams{
				ArchiveID: archive.ID,
				TagID:     tag.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func Character(c context.Context, q *repository.Queries, p *models.ArchivePayload, archive *repository.Archive) error {
	if err := q.RemoveCharacterFromArchive(c, archive.ID); err != nil {
		return err
	}
	for _, item := range p.Character {
		i := strings.ToLower(item)
		char, _ := q.GetCharacter(c, i)

		if char.Character == i {
			if err := q.AddCharacterToArchive(c, repository.AddCharacterToArchiveParams{
				ArchiveID:   archive.ID,
				CharacterID: char.ID,
			}); err != nil {
				return err
			}
		} else {
			char, err := q.CreateCharacter(c, i)
			if err != nil {
				return err
			}
			if err := q.AddCharacterToArchive(c, repository.AddCharacterToArchiveParams{
				ArchiveID:   archive.ID,
				CharacterID: char.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func Parody(c context.Context, q *repository.Queries, p *models.ArchivePayload, archive *repository.Archive) error {
	if err := q.RemoveParodyFromArchive(c, archive.ID); err != nil {
		return err
	}
	for _, item := range p.Parody {
		i := strings.ToLower(item)
		parody, _ := q.GetParody(c, i)

		if parody.Parody == i {
			if err := q.AddParodyToArchive(c, repository.AddParodyToArchiveParams{
				ArchiveID: archive.ID,
				ParodyID:  parody.ID,
			}); err != nil {
				return err
			}
		} else {
			parody, err := q.CreateParody(c, i)
			if err != nil {
				return err
			}
			if err := q.AddParodyToArchive(c, repository.AddParodyToArchiveParams{
				ArchiveID: archive.ID,
				ParodyID:  parody.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func Artist(c context.Context, q *repository.Queries, p *models.ArchivePayload, archive *repository.Archive) error {
	if err := q.RemoveArtistFromArchive(c, archive.ID); err != nil {
		return err
	}
	for _, item := range p.Artist {
		i := strings.ToLower(item)
		artist, _ := q.GetArtistByName(c, i)

		if artist.Name == i {
			if err := q.AddArtistToArchive(c, repository.AddArtistToArchiveParams{
				ArchiveID: archive.ID,
				ArtistID:  artist.ID,
			}); err != nil {
				return err
			}
		} else {
			artist, err := q.CreateArtist(c, repository.CreateArtistParams{
				Name:      i,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				return err
			}
			if err := q.AddArtistToArchive(c, repository.AddArtistToArchiveParams{
				ArchiveID: archive.ID,
				ArtistID:  artist.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func URL(c context.Context, q *repository.Queries, p *models.ArchivePayload, archive *repository.Archive) error {
	if err := q.RemoveArchiveUrl(c, archive.ID); err != nil {
		log.Err(err).Msg("remove url")
		return err
	}
	for _, i := range p.URL {
		exists, err := q.ArchiveUrlExists(c, i)
		if err != nil {
			return err
		}
		if exists.RowsAffected() == 0 {
			if err := q.CreateArchiveURL(c, repository.CreateArchiveURLParams{
				Url:       i,
				ArchiveID: archive.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
