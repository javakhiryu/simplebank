package api

import (
	"errors"
	"net/http"
	db "simplebank/db/sqlc"
	"time"

	"github.com/gin-gonic/gin"
)

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
type refreshTokenResponse struct {
	AcccessToken         string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

// refreshToken godoc
//
//	@Summary		Refresh a token
//	@Description	Refresh a token
//	@Tags			token
//	@Accept			json
//	@Produce		json
//	@Param			request	body		refreshTokenRequest	true	"Refresh Token Request"
//	@Success		200		{object}	refreshTokenResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/refreshToken [post]
func (server *Server) refreshToken(ctx *gin.Context) {

	var req refreshTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == db.ErrNoRowsFound {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if session.RefreshToken != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("invalid refresh token")))
		return
	}

	if session.Username != refreshPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("incorrect session user")))
		return
	}

	if time.Now().After(refreshPayload.ExpiredAt) {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("session has expired")))
		return
	}

	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("session is blocked")))
		return
	}

	user, err := server.store.GetUser(ctx, session.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(session.Username, user.Role, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := refreshTokenResponse{
		AcccessToken:         accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, res)

}
