package gm

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (s *Server) handleLogin(ctx iris.Context) {
	var req loginRequest
	if err := ctx.ReadJSON(&req); err != nil {
		s.fail(ctx, iris.StatusBadRequest, "invalid payload")
		return
	}
	user, err := s.store.GetGMUserByUsername(ctx.Request().Context(), req.Username)
	if err != nil {
		s.fail(ctx, iris.StatusUnauthorized, "invalid credentials")
		return
	}
	if user.Status != "" && user.Status != "active" {
		s.fail(ctx, iris.StatusForbidden, "account disabled")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		s.fail(ctx, iris.StatusUnauthorized, "invalid credentials")
		return
	}

	user.LastLoginAt = time.Now().Unix()
	_, _ = s.store.UpdateGMUser(ctx.Request().Context(), &GMUser{ID: user.ID, LastLoginAt: user.LastLoginAt}, nil)

	signed, expiresAt, err := s.buildToken(user)
	if err != nil {
		s.fail(ctx, iris.StatusInternalServerError, "token sign failed")
		return
	}

	s.ok(ctx, loginResponse{Token: signed, ExpiresAt: expiresAt})
}

func (s *Server) jwtMiddleware(ctx iris.Context) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		s.fail(ctx, iris.StatusUnauthorized, "missing token")
		return
	}
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		s.fail(ctx, iris.StatusUnauthorized, "invalid token")
		return
	}
	claims := &gmClaims{}
	_, err := jwt.ParseWithClaims(auth[len(prefix):], claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		s.fail(ctx, iris.StatusUnauthorized, "token invalid")
		return
	}
	ctx.Values().Set("gm_claims", claims)
	ctx.Next()
}

func (s *Server) handleProfile(ctx iris.Context) {
	claims, ok := s.claimsFromCtx(ctx)
	if !ok {
		s.fail(ctx, iris.StatusUnauthorized, "unauthorized")
		return
	}
	s.ok(ctx, iris.Map{
		"id":       claims.UID,
		"username": claims.Subject,
		"roles":    claims.Roles,
		"perms":    claims.Perms,
	})
}
