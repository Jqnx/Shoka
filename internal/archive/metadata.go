package archive

import (
	"Shoka/internal/repository"
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	Qtx     *repository.Queries
	Archive *Archive
	ID      int64
}

func NewMetadata(qtx *repository.Queries, archive *Archive, id int64) *Metadata {
	return &Metadata{
		Qtx:     qtx,
		Archive: archive,
		ID:      id,
	}
}

func (m *Metadata) Tag(c context.Context) error {
	if err := m.Qtx.RemoveTagFromArchive(c, m.ID); err != nil {
		return err
	}
	for _, item := range *m.Archive.Tags {
		i := strings.ToLower(item.Tag)
		tag, _ := m.Qtx.GetTag(c, i)

		if tag.Tag == i {
			if err := m.Qtx.AddTagToArchive(c, repository.AddTagToArchiveParams{
				ArchiveID: m.ID,
				TagID:     tag.ID,
			}); err != nil {
				return err
			}
		} else {
			tag, err := m.Qtx.CreateTag(c, i)
			if err != nil {
				return err
			}
			if err := m.Qtx.AddTagToArchive(c, repository.AddTagToArchiveParams{
				ArchiveID: m.ID,
				TagID:     tag.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Metadata) Character(c context.Context) error {
	if err := m.Qtx.RemoveCharacterFromArchive(c, m.ID); err != nil {
		return err
	}
	for _, item := range *m.Archive.Character {
		i := strings.ToLower(item.Character)
		char, _ := m.Qtx.GetCharacter(c, i)

		if char.Character == i {
			if err := m.Qtx.AddCharacterToArchive(c, repository.AddCharacterToArchiveParams{
				ArchiveID:   m.ID,
				CharacterID: char.ID,
			}); err != nil {
				return err
			}
		} else {
			char, err := m.Qtx.CreateCharacter(c, i)
			if err != nil {
				return err
			}
			if err := m.Qtx.AddCharacterToArchive(c, repository.AddCharacterToArchiveParams{
				ArchiveID:   m.ID,
				CharacterID: char.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Metadata) Parody(c context.Context) error {
	if err := m.Qtx.RemoveParodyFromArchive(c, m.ID); err != nil {
		return err
	}
	for _, item := range *m.Archive.Parody {
		i := strings.ToLower(item.Parody)
		parody, _ := m.Qtx.GetParody(c, i)

		if parody.Parody == i {
			if err := m.Qtx.AddParodyToArchive(c, repository.AddParodyToArchiveParams{
				ArchiveID: m.ID,
				ParodyID:  parody.ID,
			}); err != nil {
				return err
			}
		} else {
			parody, err := m.Qtx.CreateParody(c, i)
			if err != nil {
				return err
			}
			if err := m.Qtx.AddParodyToArchive(c, repository.AddParodyToArchiveParams{
				ArchiveID: m.ID,
				ParodyID:  parody.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Metadata) Artist(c context.Context) error {
	if err := m.Qtx.RemoveArtistFromArchive(c, m.ID); err != nil {
		return err
	}
	for _, item := range *m.Archive.Artist {
		i := strings.ToLower(item.Artist)
		artist, _ := m.Qtx.GetArtistByName(c, i)

		if artist.Name == i {
			if err := m.Qtx.AddArtistToArchive(c, repository.AddArtistToArchiveParams{
				ArchiveID: m.ID,
				ArtistID:  artist.ID,
			}); err != nil {
				return err
			}
		} else {
			artist, err := m.Qtx.CreateArtist(c, repository.CreateArtistParams{
				Name:      i,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
			if err != nil {
				return err
			}
			if err := m.Qtx.AddArtistToArchive(c, repository.AddArtistToArchiveParams{
				ArchiveID: m.ID,
				ArtistID:  artist.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Metadata) URL(c context.Context) error {
	if err := m.Qtx.RemoveArchiveUrl(c, m.ID); err != nil {
		return err
	}
	for _, i := range *m.Archive.URL {
		exists, err := m.Qtx.ArchiveUrlExists(c, i.URL)
		if err != nil {
			return err
		}
		if exists.RowsAffected() == 0 {
			if err := m.Qtx.CreateArchiveURL(c, repository.CreateArchiveURLParams{
				Url:       i.URL,
				ArchiveID: m.ID,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewArchiveID() string {
	newUuid := uuid.New()
	buf := newUuid[0:4]
	return hex.EncodeToString(buf)
}
