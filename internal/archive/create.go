package archive

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// TODO:
// Archive Metadata Queries
// Create Archive Transaction with metadata

func CreateTransaction(ctx context.Context,
	db *pgx.Conn,
	queries *repository.Queries,
	payload *models.ArchivePayload,
) (*repository.Archive, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	qtx := queries.WithTx(tx)
	lang := strings.ToLower(payload.Language)
	category := strings.ToLower(payload.Category)
	archive, err := qtx.CreateArchive(ctx, repository.CreateArchiveParams{
		Title:     payload.Title,
		Summary:   &payload.Summary,
		Lang:      &lang,
		Category:  &category,
		FilePath:  &payload.FilePath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &archive, tx.Commit(ctx)
}
