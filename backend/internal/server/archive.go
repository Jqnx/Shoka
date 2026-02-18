package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"reflect"
	"strconv"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/filter"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/sources"
	"Shoka/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

	var uid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		uid = util.GetUserFromRequest(c)
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("size"))
	if pageSize == 0 {
		pageSize = 10
	}

	order := util.GetSortOrderFromRequest(c)

	archives, err := s.repo.GetArchiveSortList(ctx, repository.GetArchiveSortListParams{
		Uid:     uid,
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

	count, err := s.repo.CountArchives(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.ArchiveListResponse[repository.GetArchiveSortListRow]{
		Archives: archives,
		Count:    int(count),
	})
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id := c.Param("id")

	var userid uuid.UUID
	header := c.Request.Header.Get("Authorization")
	if header != "" {
		userid = util.GetUserFromRequest(c)
	}

	result, err := archive.GetResponse(s.app, id, userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
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

	archives, err := s.repo.SearchArchiveList(ctx, repository.SearchArchiveListParams{
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
	count, err := s.repo.CountSearchArchive(ctx, query)
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

// TODO: REDO
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
		userid = util.GetUserFromRequest(c)
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
			tot, err := s.repo.CountUserFavoriteArchive(ctx, userid)
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
			archive, err := s.repo.GetUserFavoriteArchiveShuffle(ctx, repository.GetUserFavoriteArchiveShuffleParams{
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
	form, err := sources.NewSource(s.app.Cfg, config.SourceForm, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	form.Unmarshal(payload)
	meta, err := form.GetMetadata()
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Update Archive
	arch := archive.NewArchive(s.app)
	arch.Update(id, &meta[0])
	if err := arch.UpdateInDB(ctx, s.app); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	res, err := archive.GetResponse(s.app, arch.ID, uuid.Nil)
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
