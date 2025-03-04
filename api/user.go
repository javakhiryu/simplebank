package api

import (
	"database/sql"
	"errors"
	"net/http"
	db "simplebank/db/sqlc"
	"simplebank/token"
	"simplebank/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Cause   error  `swaggertype:"string"`
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}
type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// createUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user account
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		createUserRequest	true	"Create User Request"
//	@Success		200		{object}	userResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/createUser [post]
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	var res userResponse
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashedPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code.Name() {
		case "unique_violation":
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res = newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}

type getUserRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

// getUser godoc
//
//	@Summary		Get a user
//	@Description	Get a user by username
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			username	path		string	true	"Username"
//	@Success		200			{object}	userResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Router			/getUser/{username} [get]
//	@Security		Bearer
func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res := newUserResponse(user)
	ctx.JSON(http.StatusOK, res)
}

type UpdateUserHashedPasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// updateUserHashedPassword godoc
//
//	@Summary		Update user password
//	@Description	Update user password
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		UpdateUserHashedPasswordRequest	true	"Update User Hashed Password Request"
//	@Success		200		{object}	userResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/updateUserHashedPassword [patch]
//	@Security		Bearer
func (server *Server) updateUserHashedPassword(ctx *gin.Context) {
	var req UpdateUserHashedPasswordRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUser(ctx, authPayload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	err = util.CheckPassword(req.NewPassword, user.HashedPassword)
	if err == nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errors.New("new password cannot be the same as old password")))
		return
	}
	err = util.CheckPassword(req.OldPassword, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusForbidden, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashedPassword(req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	arg := db.UpdateUserHashedPasswordParams{
		HashedPassword: hashedPassword,
		Username:       authPayload.Username,
	}
	user, err = server.store.UpdateUserHashedPassword(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res := newUserResponse(user)
	ctx.JSON(http.StatusOK, res)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AcccessToken          string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

// loginUser godoc
//
//	@Summary		Login user
//	@Description	Login user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			request	body		loginUserRequest	true	"Login User Request"
//	@Success		200		{object}	loginUserResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/login [post]
func (server *Server) loginUser(ctx *gin.Context) {

	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(req.Username, server.config.RefreshTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	arg := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     req.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.GetHeader("User-Agent"),
		IpAddress:    ctx.ClientIP(),
		ExpiresAt:    refreshPayload.ExpiredAt,
		CreatedAt:    time.Now(),
		IsBlocked:    false,
	}
	session, err := server.store.CreateSession(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res := loginUserResponse{
		SessionID:             session.ID,
		AcccessToken:          accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, res)

}
