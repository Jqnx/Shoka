package server

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

// TODO:
// Print proper error responses instead of just printing errors directly

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
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}
	if result.RowsAffected() != 0 {
		c.JSON(http.StatusConflict, &models.ResponseError{
			Status:  "error",
			Message: config.ErrArchiveNoDuplicates.Error(),
		})
		return
	}

	archive, err := archive.CreateTransaction(ctx, s.db, s.repo, &payload, s.log)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, archive)
}

// Read

//func (s *Server) getLastIDHandler(c *gin.Context) {
//	lastid, err := s.repo.GetArchiveLastAID(c)
//	if err != nil {
//		s.log.Error().AnErr("getLastIDHandler", err).Send()
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	c.JSON(http.StatusOK, lastid)
//}

func (s *Server) getArchiveListHandler(c *gin.Context) {
	ctx := context.Background()
	p := c.Query("page")
	ps := c.Query("size")

	if p == "" && ps == "" {
		// archives, err := archive.GetAll(ctx, s.repo, s.log)
		archives, err := s.repo.GetAllArchives(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		if len(archives) == 0 {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, archives)
	} else {

		page, _ := strconv.Atoi(p)
		if page == 0 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(ps)
		if pageSize == 0 {
			pageSize = 10
		}

		archives, err := s.repo.GetArchiveList(ctx, repository.GetArchiveListParams{
			Limit:  int32(pageSize),
			Offset: (int32(page) - 1) * int32(pageSize),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		if len(archives) == 0 {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, archives)
	}
}

func (s *Server) getAllTagHandler(c *gin.Context) {
	ctx := context.Background()

	tags, err := s.repo.GetAllTags(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoTags.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, tags)
}

func (s *Server) getAllCharacterHandler(c *gin.Context) {
	ctx := context.Background()

	characters, err := s.repo.GetAllCharacter(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoCharacter.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, characters)
}

func (s *Server) getAllParodyHandler(c *gin.Context) {
	ctx := context.Background()

	parodies, err := s.repo.GetAllParodies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoParody.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, parodies)
}

func (s *Server) getArchiveHandler(c *gin.Context) {
	id := c.Param("id")

	ctx := context.Background()
	// archive, err := archive.Get(ctx, s.repo, id, s.log)

	arch, err := s.repo.GetArchiveByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	tags, err := s.repo.GetArchiveTags(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	characters, err := s.repo.GetArchiveCharacters(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	parodies, err := s.repo.GetArchiveParodies(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	urls, err := s.repo.GetArchiveURLs(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
	artists, err := s.repo.GetArchiveArtists(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
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

	//d, err := os.ReadDir(filepath.Join(p, "pages"))
	//if err != nil {
	//	if errors.Is(err, os.ErrNotExist) {
	//		w := workers.NewWorkers(s.app, false, ctx)
	//		go w.Thumbs(&arch)
	//	}
	//} else if len(d) < int(arch.PageCount) {
	//	w := workers.NewWorkers(s.app, false, ctx)
	//	go w.Thumbs(&arch)
	//}

	// NOTE: TEMP
	// pages := archive.GetPages(int(arch.PageCount))

	result := &models.ArchiveResponse{
		ArchiveID: id,
		Title:     arch.Title,
		Summary:   arch.Summary,
		Tags:      tags,
		Artist:    artists,
		Parody:    parodies,
		Character: characters,
		Language:  arch.Language,
		Category:  arch.Category,
		PageCount: arch.PageCount,
		Url:       urls,
		Hash:      arch.Hash,
		Pages:     pages,
		// ThumbsPath: archive.ThumbsPath,
		Type:      arch.Type,
		CreatedAt: arch.CreatedAt,
		UpdatedAt: arch.UpdatedAt,
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) getArchiveByTagHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	tag := strings.ToLower(c.Param("tag"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.TagExists(ctx, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrTagNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByTag(ctx, s.repo, tag, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, archives)
}

func (s *Server) getArchiveByCharacterHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	character := strings.ToLower(c.Param("character"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.CharacterExists(ctx, character)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrCharacterNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByCharacter(ctx, s.repo, character, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, archives)
}

func (s *Server) getArchiveByParodyHandler(c *gin.Context) {
	// Get tag name from url parameter, makes it lowercase
	parody := strings.ToLower(c.Param("parody"))

	// Creates context
	ctx := context.Background()

	// Checks if tag exists
	exists, err := s.repo.ParodyExists(ctx, parody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// If tag does not exists respond with 404 ErrTagNotFound
	if exists.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrParodyNotFound.Error(),
		})
		return
	}

	// Get archives
	archives, err := archive.GetByParody(ctx, s.repo, parody, s.log)
	if err != nil {
		// If no archives found respond with 404 ErrNoArchive
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrNoArchive.Error(),
			})
			return
			// Otherwise respond with 500 and error
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	// Respond with 200 Success
	c.JSON(http.StatusOK, archives)
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
		c.JSON(http.StatusNotFound, &models.ResponseError{
			Status:  "error",
			Message: config.ErrArchiveNotFound.Error(),
		})
		return
	}

	result, err := archive.UpdateTransaction(ctx,
		s.db,
		s.repo,
		&payload,
		id,
		s.log,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.ResponseError{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete

func (s *Server) deleteArchiveHandler(c *gin.Context) {
	id := c.Param("id")

	ctx := context.Background()
	if err := archive.DeleteTransaction(ctx, s.db, s.repo, id, s.log); err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, &models.ResponseError{
				Status:  "error",
				Message: config.ErrArchiveNotFound.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, &models.ResponseError{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
