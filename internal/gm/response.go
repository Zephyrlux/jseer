package gm

import "github.com/kataras/iris/v12"

type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *Server) ok(ctx iris.Context, data interface{}) {
	ctx.JSON(apiResponse{Code: 0, Data: data})
}

func (s *Server) fail(ctx iris.Context, status int, message string) {
	ctx.StopWithJSON(status, apiResponse{Code: status, Message: message})
}
