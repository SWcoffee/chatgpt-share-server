package api

import (
	"backend/openai/api/auth"

	"github.com/gogf/gf/v2/frame/g"
)

func init() {

	s := g.Server()
	apiGroup := s.Group("/api")
	// 暂时不刷新rt
	// apiGroup.GET("/auth/session", auth.Session)
	apiGroup.GET("/auth/csrf", auth.Csrf)
	apiGroup.POST("/auth/signout", auth.SignOut)
}
