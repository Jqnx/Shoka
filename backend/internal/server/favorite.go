package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/filter"
	"Shoka/internal/models"
	"Shoka/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func (s *Server) favoriteArchiveHandler(c *gin.Context) {
	ctx := context.Background()

	archiveId := c.Param("id")
	userId, _ := uuid.Parse(c.GetString("userid"))

	archive, err := s.repo.GetArchiveByID(ctx, archiveId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	check, err := s.repo.ArchiveIsFavorited(ctx, repository.ArchiveIsFavoritedParams{
		ArchiveID: archive.ID,
		UserID:    userId,
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
			UserID:      userId,
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
			"user_id":     userId,
			"is_favorite": true,
		})
	} else {
		if err := s.repo.RemoveFavoriteArchive(ctx, repository.RemoveFavoriteArchiveParams{
			ArchiveID: archive.ID,
			UserID:    userId,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"archive":     archive,
			"user_id":     userId,
			"is_favorite": false,
		})
	}
}

// TODO: Add sorting and filters to get user favorite archives list
func (s *Server) getUserFavoriteArchives(c *gin.Context) {
	ctx := context.Background()
	userId, _ := uuid.Parse(c.GetString("userid"))

	p := c.Query("page")
	ps := c.Query("size")

	page, _ := strconv.Atoi(p)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ps)
	if pageSize == 0 {
		pageSize = 10
	}

	if p == "" && ps == "" {
		archives, err := s.repo.GetUserFavoriteArchiveAll(ctx, userId)
		if err != nil {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
		})
		return
	}

	archives, err := s.repo.GetUserFavoriteArchiveList(ctx, repository.GetUserFavoriteArchiveListParams{
		ID:     userId,
		Limit:  int32(pageSize),
		Offset: (int32(page) - 1) * int32(pageSize),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}
	count, err := s.repo.CountUserFavoriteArchive(ctx, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"archives": archives,
		"total":    count,
	})
}

func (s *Server) getFavoriteArchiveFilterHandler(c *gin.Context) {
	var payload models.ArchiveFilters
	if errPost := c.ShouldBind(&payload); errPost != nil {
		var verr validator.ValidationErrors
		if errors.As(errPost, &verr) {
			c.JSON(http.StatusBadRequest, &models.ResponseFail{
				Status: "fail",
				Data:   config.Validate(verr),
			})
			return
		}
		c.JSON(http.StatusBadRequest, &models.ResponseFail{
			Status: "fail",
			Data:   errPost.Error(),
		})
		return
	}

	ctx := context.Background()

	// URL Queries
	userid, _ := uuid.Parse(c.GetString("userid"))
	p := c.Query("page")
	ps := c.Query("size")
	sortby := c.Query("sortby")
	sortdir := c.Query("sortdir")

	// Set Page & Page Size
	page, _ := strconv.Atoi(p)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ps)
	if pageSize == 0 {
		pageSize = 10
	}

	// Query with filters if query is not empty
	if !reflect.DeepEqual(models.ArchiveFilters{
		Tags:       []string{},
		Artists:    []string{},
		Characters: []string{},
		Parodies:   []string{},
		Languages:  []string{},
		Categories: []string{},
	}, payload) {
		switch sortby {
		case "title":
			if sortdir == "desc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "title_desc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "title_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		case "page_count":
			if sortdir == "desc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "page_count_desc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "page_count_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		case "created_at":
			if sortdir == "desc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "created_at_desc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "created_at_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		case "updated_at":
			if sortdir == "desc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "updated_at_desc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "updated_at_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		case "release_date":
			if sortdir == "asc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "release_date_asc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "release_date_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					fmt.Println(err)
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		case "favorited_at":
			if sortdir == "asc" {
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "favorited_at_asc")
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
				archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "favorited_at_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		default:
			archives, err := filter.MatchAndGetFavorites(payload, s.repo, page, pageSize, userid, "favorited_at_desc")
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, archives)
			return
		}
	} else {
		// Otherwise only sort
		count, err := s.repo.CountUserFavoriteArchive(ctx, userid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		switch sortby {
		case "title":
			if sortdir == "desc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "title_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "title_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		case "page_count":
			if sortdir == "desc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "page_count_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "page_count_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		case "created_at":
			if sortdir == "desc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "created_at_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "created_at_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		case "updated_at":
			if sortdir == "desc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "updated_at_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "updated_at_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		case "release_date":
			if sortdir == "asc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "release_date_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "release_date_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		case "favorited_at":
			if sortdir == "asc" {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "favorited_at_asc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			} else {
				archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
					UserID:  userid,
					OrderBy: "favorited_at_desc",
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
				c.JSON(http.StatusOK, gin.H{
					"archives": archives,
					"total":    count,
				})
				return
			}
		default:
			archives, err := s.repo.GetFavoriteArchiveSortList(ctx, repository.GetFavoriteArchiveSortListParams{
				UserID:  userid,
				OrderBy: "favorited_at_desc",
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
			c.JSON(http.StatusOK, gin.H{
				"archives": archives,
				"total":    count,
			})
			return
		}
	}
}
