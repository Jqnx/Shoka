package archive

import (
	"context"
	"strings"
	"time"

	"Shoka/internal/repository"

	"github.com/google/uuid"
	"github.com/jxskiss/base62"
)

type Metadata struct {
	Qtx     *repository.Queries
	Archive *Archive
	ID      string
}

func NewMetadata(qtx *repository.Queries, archive *Archive, id string) *Metadata {
	return &Metadata{
		Qtx:     qtx,
		Archive: archive,
		ID:      id,
	}
}

func (m *Metadata) Tag(c context.Context) error {
	tags, err := m.Qtx.RemoveTagFromArchive(c, m.ID)
	if err != nil {
		return err
	}
	for _, item := range tags {
		m.Qtx.UpdateTagCount(c, repository.UpdateTagCountParams{
			ID:    item.TagID,
			Count: item.Count - 1,
		})
	}

	for _, item := range *m.Archive.Tags {
		if len(item.Tag) == 0 {
			return nil
		}

		i := strings.ToLower(item.Tag)
		tag, _ := m.Qtx.GetTag(c, i)

		if tag.Name == i {
			if err := m.Qtx.AddTagToArchive(c, repository.AddTagToArchiveParams{
				ArchiveID: m.ID,
				TagID:     tag.ID,
			}); err != nil {
				return err
			}
			if err := m.Qtx.UpdateTagCount(c, repository.UpdateTagCountParams{
				ID:    tag.ID,
				Count: tag.Count + 1,
			}); err != nil {
				return err
			}
		} else {
			tag, err := m.Qtx.CreateTag(c, repository.CreateTagParams{
				Name:  i,
				Count: 1,
			})
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
	characters, err := m.Qtx.RemoveCharacterFromArchive(c, m.ID)
	if err != nil {
		return err
	}
	for _, item := range characters {
		if err := m.Qtx.UpdateCharacterCount(c, repository.UpdateCharacterCountParams{
			ID:    item.CharacterID,
			Count: item.Count - 1,
		}); err != nil {
			return err
		}
	}

	for _, item := range *m.Archive.Character {
		if len(item.Character) == 0 {
			return nil
		}

		i := strings.ToLower(item.Character)
		char, _ := m.Qtx.GetCharacter(c, i)

		if char.Name == i {
			if err := m.Qtx.AddCharacterToArchive(c, repository.AddCharacterToArchiveParams{
				ArchiveID:   m.ID,
				CharacterID: char.ID,
			}); err != nil {
				return err
			}
			if err := m.Qtx.UpdateCharacterCount(c, repository.UpdateCharacterCountParams{
				ID:    char.ID,
				Count: char.Count + 1,
			}); err != nil {
				return err
			}
		} else {
			char, err := m.Qtx.CreateCharacter(c, repository.CreateCharacterParams{
				Name:  i,
				Count: 1,
			})
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
	parodies, err := m.Qtx.RemoveParodyFromArchive(c, m.ID)
	if err != nil {
		return err
	}

	for _, item := range parodies {
		if err := m.Qtx.UpdateParodyCount(c, repository.UpdateParodyCountParams{
			ID:    item.ParodyID,
			Count: item.Count - 1,
		}); err != nil {
			return err
		}
	}
	for _, item := range *m.Archive.Parody {
		if len(item.Parody) == 0 {
			return nil
		}

		i := strings.ToLower(item.Parody)
		parody, _ := m.Qtx.GetParody(c, i)

		if parody.Name == i {
			if err := m.Qtx.AddParodyToArchive(c, repository.AddParodyToArchiveParams{
				ArchiveID: m.ID,
				ParodyID:  parody.ID,
			}); err != nil {
				return err
			}
			if err := m.Qtx.UpdateParodyCount(c, repository.UpdateParodyCountParams{
				ID:    parody.ID,
				Count: parody.Count + 1,
			}); err != nil {
				return err
			}
		} else {
			parody, err := m.Qtx.CreateParody(c, repository.CreateParodyParams{
				Name:  i,
				Count: 1,
			})
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
	artists, err := m.Qtx.RemoveArtistFromArchive(c, m.ID)
	if err != nil {
		return err
	}

	for _, item := range artists {
		if err := m.Qtx.UpdateArtistCount(c, repository.UpdateArtistCountParams{
			ID:    item.ArtistID,
			Count: item.Count - 1,
		}); err != nil {
			return err
		}
	}
	for _, item := range *m.Archive.Artist {
		if len(item.Artist) == 0 {
			return nil
		}

		i := strings.ToLower(item.Artist)
		artist, _ := m.Qtx.GetArtistByName(c, i)

		if artist.Name == i {
			if err := m.Qtx.AddArtistToArchive(c, repository.AddArtistToArchiveParams{
				ArchiveID: m.ID,
				ArtistID:  artist.ID,
			}); err != nil {
				return err
			}
			if err := m.Qtx.UpdateArtistCount(c, repository.UpdateArtistCountParams{
				ID:    artist.ID,
				Count: artist.Count + 1,
			}); err != nil {
				return err
			}
		} else {
			artist, err := m.Qtx.CreateArtist(c, repository.CreateArtistParams{
				Name:      i,
				Count:     1,
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
	newUUID := uuid.New()
	test := base62.EncodeToString(newUUID[:])
	out := test[0:8]
	return out
}
