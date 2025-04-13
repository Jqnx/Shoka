package group

import (
	"Shoka/internal/repository"
	"context"
	"log/slog"
)

func Delete(c context.Context,
	q *repository.Queries,
	name string,
	log *slog.Logger,
) error {
	if err := q.DeleteGroup(c, name); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
