package gm

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kataras/iris/v12"
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
		ctx.StopWithJSON(iris.StatusBadRequest, iris.Map{"error": "invalid payload"})
		return
	}
	if req.Username != s.cfg.DefaultAdminUser || req.Password != s.cfg.DefaultAdminPass {
		ctx.StopWithJSON(iris.StatusUnauthorized, iris.Map{"error": "invalid credentials"})
		return
	}

	expires := time.Now().Add(time.Duration(s.cfg.TokenTTLMinutes) * time.Minute)
	claims := jwt.MapClaims{
		"sub":   req.Username,
		"roles": []string{"super_admin"},
		"exp":   expires.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		ctx.StopWithJSON(iris.StatusInternalServerError, iris.Map{"error": "token sign failed"})
		return
	}

	ctx.JSON(loginResponse{Token: signed, ExpiresAt: expires.Unix()})
}

func (s *Server) jwtMiddleware(ctx iris.Context) {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		ctx.StopWithJSON(iris.StatusUnauthorized, iris.Map{"error": "missing token"})
		return
	}
	const prefix = "Bearer "
	if len(auth) <= len(prefix) || auth[:len(prefix)] != prefix {
		ctx.StopWithJSON(iris.StatusUnauthorized, iris.Map{"error": "invalid token"})
		return
	}
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(auth[len(prefix):], claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		ctx.StopWithJSON(iris.StatusUnauthorized, iris.Map{"error": "token invalid"})
		return
	}
	ctx.Values().Set("gm_claims", claims)
	ctx.Next()
}
