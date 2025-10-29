package group

import (
	"Shoka/internal/logger"
	"Shoka/internal/repository"
	"context"
)

func Delete(c context.Context,
	q *repository.Queries,
	name string,
	log logger.Logger,
) error {
	if err := q.DeleteGroup(c, name); err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
