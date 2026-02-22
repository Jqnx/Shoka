// Package metadata contains almost all logic related to
// adding, updating and removing archive metadata in the database.
package metadata

import (
	"context"

	"Shoka/internal/repository"

	"github.com/google/uuid"
	"github.com/jxskiss/base62"
)

type Metadata struct {
	Qtx        *repository.Queries
	ID         string
	Tags       []int32
	URLs       []string
	Parodies   []int32
	Characters []int32
	Artists    []int32
}

func GetCurrent(ctx context.Context, qtx *repository.Queries, id string) (*Metadata, error) {
	tags, err := qtx.GetArchiveTagIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	artists, err := qtx.GetArchiveArtistIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	characters, err := qtx.GetArchiveCharacterIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	parodies, err := qtx.GetArchiveParodyIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	urlsRow, err := qtx.GetArchiveUrls(ctx, id)
	if err != nil {
		return nil, err
	}

	var urls []string
	for _, i := range urlsRow {
		urls = append(urls, i.Url)
	}

	return &Metadata{
		Qtx:        qtx,
		ID:         id,
		Tags:       tags,
		Artists:    artists,
		Characters: characters,
		Parodies:   parodies,
		URLs:       urls,
	}, nil
}

//func (m *Metadata) URL(c context.Context) error {
//	if err := m.Qtx.RemoveArchiveUrl(c, m.ID); err != nil {
//		return err
//	}
//	for _, i := range *m.Archive.URL {
//		if len(i.URL) == 0 {
//			return nil
//		}
//
//		exists, err := m.Qtx.ArchiveUrlExists(c, i.URL)
//		if err != nil {
//			return err
//		}
//		if exists.RowsAffected() == 0 {
//			if err := m.Qtx.CreateArchiveURL(c, repository.CreateArchiveURLParams{
//				Url:       i.URL,
//				ArchiveID: m.ID,
//			}); err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}

func NewArchiveID() string {
	newUUID := uuid.New()
	test := base62.EncodeToString(newUUID[:])
	out := test[0:8]
	return out
}
