package server

import (
	"Shoka/internal/config"
	"Shoka/internal/filter"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"errors"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// TODO: Maybe filter/sort all function for when there is no page or pagesize given.

func (s *Server) getArchiveFilterHandler(c *gin.Context) {
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
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "title_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			} else {
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "title_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "page_count_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			} else {
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "page_count_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "created_at_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			} else {
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "created_at_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "updated_at_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			} else {
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "updated_at_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "release_date_asc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			} else {
				archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "release_date_desc")
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				c.JSON(http.StatusOK, archives)
				return
			}
		default:
			archives, err := filter.MatchAndGet(payload, s.repo, page, pageSize, "title_asc")
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
		switch sortby {
		case "title":
			if sortdir == "desc" {
				archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
					OrderBy: "title_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "created_at_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "created_at_asc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "updated_at_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "updated_at_asc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "release_date_asc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
					OrderBy: "release_date_desc",
					Limit:   int32(pageSize),
					Offset:  (int32(page) - 1) * int32(pageSize),
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
						Status:  "error",
						Message: err.Error(),
					})
					return
				}
				count, err := s.repo.CountArchives(ctx)
				if err != nil {
					c.JSON(http.StatusInternalServerError, &models.ResponseError{
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
				OrderBy: "title_asc",
				Limit:   int32(pageSize),
				Offset:  (int32(page) - 1) * int32(pageSize),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.ResponseError{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
			count, err := s.repo.CountArchives(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, &models.ResponseError{
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

func (s *Server) getAllFiltersHandler(c *gin.Context) {
	ctx := context.Background()

	tags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	artists, err := s.repo.GetAllArtists(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	characters, err := s.repo.GetAllCharacter(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	parodies, err := s.repo.GetAllParodies(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	languages, err := s.repo.GetAllLanguage(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	categories, err := s.repo.GetAllCategory(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":       tags,
		"artists":    artists,
		"characters": characters,
		"parodies":   parodies,
		"languages":  languages,
		"categories": categories,
	})
}
