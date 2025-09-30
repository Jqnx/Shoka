package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/filter"
	"Shoka/internal/fsutil"
	"Shoka/internal/metadata"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TODO: Print proper error responses instead of just printing errors directly
// TODO: Update handlers to use Metadata builder instead

// Create
func (s *Server) createArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload
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
	result, err := s.repo.ArchiveExists(ctx, payload.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNoDuplicates.Error(),
		})
		return
	}

	// TODO: Update handlers to use Metadata builder instead
	//archive, err := archive.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, &models.ResponseError{
	//		Status:  "error",
	//		Message: err.Error(),
	//	})
	//	return
	//}

	c.JSON(http.StatusCreated, nil)
}

// Read

func (s *Server) getArchiveListHandler(c *gin.Context) {
	ctx := context.Background()
	p := c.Query("page")
	ps := c.Query("size")
	sortby := c.Query("sortby")
	sortdir := c.Query("sortdir")

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		token := util.GetAuthTokenFromHeader(header)
		user, err := s.repo.GetUserByToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}
		uid = user.ID
	}

	if p == "" && ps == "" && sortby == "" && sortdir == "" {
		archives, err := s.repo.GetAllArchives(ctx, uid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		if len(archives) == 0 {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"archives": archives,
		})
	} else if p == "" && ps == "" {
		switch sortby {
		case "title":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "title_desc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "title_asc",
					Uid:     uid,
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
			}
		case "page_count":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "page_count_desc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "page_count_asc",
					Uid:     uid,
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
			}
		case "created_at":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "created_at_desc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "created_at_asc",
					Uid:     uid,
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
			}
		case "updated_at":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "updated_at_desc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "updated_at_asc",
					Uid:     uid,
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
			}
		case "release_date":
			if sortdir == "asc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "release_date_asc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "release_date_desc",
					Uid:     uid,
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
			}
		case "last_read":
			if sortdir == "asc" {
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "last_read_asc",
					Uid:     uid,
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
				archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
					OrderBy: "last_read_desc",
					Uid:     uid,
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
			}
		default:
			archives, err := s.repo.GetArchiveSort(ctx, repository.GetArchiveSortParams{
				OrderBy: "title_asc",
				Uid:     uid,
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
		}
	} else {
		page, _ := strconv.Atoi(p)
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ps)
		if pageSize == 0 {
			pageSize = 10
		}

		switch sortby {
		case "title":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					OrderBy: "title_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
					Uid:     uid,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					OrderBy: "title_asc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
					Uid:     uid,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					OrderBy: "page_count_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
					Uid:     uid,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					OrderBy: "page_count_asc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
					Uid:     uid,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.Response{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
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
				count, err := s.repo.CountArchives(ctx)
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
		case "last_read":
			if sortdir == "asc" {
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
					OrderBy: "last_read_asc",
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
				count, err := s.repo.CountArchives(ctx)
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
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					Uid:     uid,
					OrderBy: "last_read_desc",
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
				count, err := s.repo.CountArchives(ctx)
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
			archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
				Uid:     uid,
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
			count, err := s.repo.CountArchives(ctx)
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

func (s *Server) getArchiveHandler(c *gin.Context) {
	ctx := context.Background()
	id := c.Param("id")

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	tags, err := s.repo.GetArchiveTags(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	characters, err := s.repo.GetArchiveCharacters(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	parodies, err := s.repo.GetArchiveParodies(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	urls, err := s.repo.GetArchiveURLs(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	artists, err := s.repo.GetArchiveArtists(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.Response{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	var pages int
	p, _ := filepath.Abs(*arch.ThumbsPath)
	d, err := os.ReadDir(filepath.Join(p, "pages"))
	if err != nil {
		pages = 0
	}
	for _, i := range d {
		if fsutil.MatchExtension(i.Name(), config.ImageExtensions) {
			pages++
		}
	}

	var userid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		token := util.GetAuthTokenFromHeader(header)
		user, err := s.repo.GetUserByToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}
		userid = user.ID
	}

	if userid != uuid.Nil {
		fav := false
		check, err := s.repo.ArchiveIsFavorited(ctx, repository.ArchiveIsFavoritedParams{
			UserID:    userid,
			ArchiveID: arch.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		if check.RowsAffected() != 0 {
			fav = true
		}

		read := &ReadingProgress{}
		rp, err := s.repo.GetUserReadingProgress(ctx, repository.GetUserReadingProgressParams{
			UserID:    userid,
			ArchiveID: arch.ID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				read.ReadingState = "unread"
				read.Progress = 0
				read.LastRead = nil
			} else {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		} else {
			read.ReadingState = rp.State
			read.Progress = rp.Page
			read.LastRead = &rp.LastRead
		}

		result := &models.ArchiveResponseFavorite{
			ID:           arch.ID,
			Title:        arch.Title,
			Summary:      arch.Summary,
			Tags:         tags,
			Artist:       artists,
			Parody:       parodies,
			Character:    characters,
			Language:     arch.Language,
			Category:     arch.Category,
			PageCount:    arch.PageCount,
			Url:          urls,
			Hash:         arch.Hash,
			Pages:        pages,
			Type:         arch.Type,
			Progress:     read.Progress,
			ReadingState: read.ReadingState,
			LastRead:     read.LastRead,
			CreatedAt:    arch.CreatedAt,
			UpdatedAt:    arch.UpdatedAt,
			ReleaseDate:  arch.ReleaseDate,
			IsFavorite:   fav,
		}

		c.JSON(http.StatusOK, result)
	} else {
		result := &models.ArchiveResponse{
			ID:          arch.ID,
			Title:       arch.Title,
			Summary:     arch.Summary,
			Tags:        tags,
			Artist:      artists,
			Parody:      parodies,
			Character:   characters,
			Language:    arch.Language,
			Category:    arch.Category,
			PageCount:   arch.PageCount,
			Url:         urls,
			Hash:        arch.Hash,
			Pages:       pages,
			Type:        arch.Type,
			CreatedAt:   arch.CreatedAt,
			UpdatedAt:   arch.UpdatedAt,
			ReleaseDate: arch.ReleaseDate,
		}

		c.JSON(http.StatusOK, result)
	}
}

// TODO: Add search for artists, groups, tankoubons aswell
func (s *Server) searchArchiveHandler(c *gin.Context) {
	query := c.Query("q")
	p := c.Query("page")
	ps := c.Query("size")

	ctx := context.Background()

	page, _ := strconv.Atoi(p)
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(ps)
	if pageSize == 0 {
		pageSize = 10
	}

	archives, err := s.repo.SearchArchivesList(ctx, repository.SearchArchivesListParams{
		WebsearchToTsquery: query,
		Limit:              int32(pageSize),
		Offset:             (int32(page) - 1) * int32(pageSize),
	})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}
	count, err := s.repo.CountSearchArchives(ctx, query)
	if err != nil {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"archives": archives,
		"total":    count[0],
	})
}

func (s *Server) shuffleArchiveHandler(c *gin.Context) {
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

	favorite, err := strconv.ParseBool(c.Query("favorite"))
	if err != nil {
		favorite = false
	}

	var userid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		token := util.GetAuthTokenFromHeader(header)
		user, err := s.repo.GetUserByToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}
		userid = user.ID
	}

	countQuery := c.Query("c")
	var count int
	if countQuery != "" {
		cnt, err := strconv.Atoi(countQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, &models.Response{
				Status:  "error",
				Message: "invalid c value, must be an integer.",
			})
			return
		}
		count = cnt
	} else {
		var total int64
		if favorite {
			tot, err := s.repo.CountUserFavoriteArchives(ctx, userid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			total = tot
		} else {
			tot, err := s.repo.CountArchives(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			total = tot
		}
		count = int(math.Floor(float64(rand.Float32()) * float64(total)))
	}

	if !reflect.DeepEqual(models.ArchiveFilters{
		Tags:       []string{},
		Artists:    []string{},
		Characters: []string{},
		Parodies:   []string{},
		Languages:  []string{},
		Categories: []string{},
	}, payload) {
		archive, err := filter.MatchAndGetShuffle(payload, s.repo, count, userid, favorite)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, archive)
	} else {
		if favorite {
			archive, err := s.repo.GetUserFavoriteArchivesShuffle(ctx, repository.GetUserFavoriteArchivesShuffleParams{
				ID:     userid,
				Limit:  1,
				Offset: int32(count),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, archive)
		} else {
			archive, err := s.repo.GetArchiveShuffle(ctx, repository.GetArchiveShuffleParams{
				Limit:  1,
				Offset: int32(count),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.Response{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, archive)
		}
	}
}

// Update
func (s *Server) updateArchiveHandler(c *gin.Context) {
	var payload models.ArchivePayload
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

	id := c.Param("id")

	ctx := context.Background()
	exists, err := s.repo.ArchiveIDExists(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}

	// Fetch Metadata
	mb := metadata.GetBuilder("form")
	d := metadata.NewDirector(mb)
	meta, err := d.FetchMetadata(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Update Archive
	ab := archive.GetBuilder(s.app)
	archive := ab.UpdateArchive(id, &meta[0])
	if err := archive.Update(ctx, s.app); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	res, err := archive.Get(ctx, s.app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Delete

//	func (s *Server) deleteArchiveHandler(c *gin.Context) {
//		id := c.Param("id")
//
//		ctx := context.Background()
//		if err := archive.DeleteTransaction(ctx, s.db, s.repo, id, s.log); err != nil {
//			if err == pgx.ErrNoRows {
//				c.JSON(http.StatusNotFound, &models.ResponseError{
//					Status:  "error",
//					Message: config.ErrArchiveNotFound.Error(),
//				})
//				return
//			} else {
//				c.JSON(http.StatusInternalServerError, &models.ResponseError{
//					Status:  "error",
//					Message: err.Error(),
//				})
//				return
//			}
//		}
//
//		c.JSON(http.StatusOK, gin.H{"status": "success"})
//	}
