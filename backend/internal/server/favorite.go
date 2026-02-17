package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"Shoka/internal/filter"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Server) favoriteArchiveHandler(c *gin.Context) {
	ctx := context.Background()

	archiveID := c.Param("id")
	userID, _ := uuid.Parse(c.GetString("userid"))

	archive, err := s.repo.GetArchiveByID(ctx, archiveID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	check, err := s.repo.ArchiveIsFavorited(ctx, repository.ArchiveIsFavoritedParams{
		ArchiveID: archive.ID,
		UserID:    userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if check.RowsAffected() == 0 {
		if err := s.repo.AddFavoriteArchive(ctx, repository.AddFavoriteArchiveParams{
			ArchiveID:   archive.ID,
			UserID:      userID,
			FavoritedAt: time.Now(),
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"archive":     archive,
			"user_id":     userID,
			"is_favorite": true,
		})
	} else {
		if err := s.repo.RemoveFavoriteArchive(ctx, repository.RemoveFavoriteArchiveParams{
			ArchiveID: archive.ID,
			UserID:    userID,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"archive":     archive,
			"user_id":     userID,
			"is_favorite": false,
		})
	}
}

func (s *Server) getFavoriteArchiveListHandler(c *gin.Context) {
	ctx := context.Background()
	filters, hasFilter := filter.GetFromQuery(c)

	uid, _ := uuid.Parse(c.GetString("userid"))

	order := util.GetSortOrderFromRequest(c)

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	// Query with filters if query is not empty
	if hasFilter {
		archives, err := filter.MatchAndGetFavorites(s.repo, filter.MatchAndGetParams{
			Filters:  *filters,
			Page:     page,
			PageSize: pageSize,
			Order:    order,
			UserID:   uid,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, archives)
		return
	} else {
		count, err := s.repo.CountUserFavoriteArchive(ctx, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
			UserID:  uid,
			OrderBy: order,
			Limit:   int32(pageSize),
			Offset:  (int32(page) - 1) * int32(pageSize),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetFavoriteArchiveSortListRow]{
			Archives: archives,
			Count:    int(count),
		})
		return
	}
}
