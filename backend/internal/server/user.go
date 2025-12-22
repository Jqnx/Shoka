package server

/*
func (s *Server) registerUser(c *gin.Context) {
	var payload models.UserPayload
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

	_, err := s.repo.GetUserByName(ctx, payload.Name)
	if err != nil {
		passHash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		_, err = s.repo.CreateUser(ctx, repository.CreateUserParams{
			ID:        uuid.New(),
			Name:      payload.Name,
			Password:  string(passHash),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			fmt.Println(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "User registered successfully!",
		})
	} else {
		c.JSON(http.StatusConflict, &models.Response{
			Status:  "error",
			Message: "Username already exists.",
		})
		return
	}
}

func (s *Server) signInUser(c *gin.Context) {
	var payload models.UserPayload
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

	user, err := s.repo.GetUserByName(ctx, payload.Name)
	if err != nil {
		c.JSON(http.StatusNotFound, &models.Response{
			Status:  "error",
			Message: "User not found.",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, &models.Response{
			Status:  "error",
			Message: "Invalid password.",
		})
		return
	}

	token, err := util.GenerateToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: "Failed to generate token.",
		})
		return
	}

	expiry := time.Now().Add(time.Hour * 24 * 3)
	useragent := c.Request.UserAgent()
	ip := util.GetIPFromGinContext(c)
	tokenHash := sha256.Sum256([]byte(token))
	thString := hex.EncodeToString(tokenHash[:])
	if err := s.repo.CreateSession(ctx, repository.CreateSessionParams{
		SessionID: uuid.New(),
		UserID:    user.ID,
		Token:     thString,
		ExpiresAt: expiry,
		CreatedAt: time.Now(),
		IpAddress: ip,
		UserAgent: &useragent,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"expiry": expiry,
	})
}

func (s *Server) getUserSession(c *gin.Context) {
	ctx := context.Background()
	username := c.GetString("username")

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
	})
}

func (s *Server) signOutUser(c *gin.Context) {
	ctx := context.Background()
	token := c.GetString("token")

	if err := s.repo.DeleteSession(ctx, token); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "Successfully logged out.",
	})
}

func (s *Server) updateUserHandler(c *gin.Context) {
	var payload models.UpdateUserPayload
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
	username := c.GetString("username")

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if payload.Password != nil {
		passHash, err := bcrypt.GenerateFromPassword([]byte(*payload.Password), 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		pass := string(passHash)

		_, err = s.repo.UpdateUser(ctx, repository.UpdateUserParams{
			ID:       user.ID,
			Name:     payload.Name,
			Password: &pass,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	} else {
		_, err = s.repo.UpdateUser(ctx, repository.UpdateUserParams{
			ID:   user.ID,
			Name: payload.Name,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
	}
}

func (s *Server) deleteUserHandler(c *gin.Context) {
	ctx := context.Background()
	username := c.GetString("username")

	user, err := s.repo.GetUserByName(ctx, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	if err := s.repo.DeleteUser(ctx, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, &models.Response{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "User sucessfully deleted.",
	})
}
*/
