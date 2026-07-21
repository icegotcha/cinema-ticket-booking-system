package user

import "github.com/gin-gonic/gin"

type UserHandler struct{}

func NewHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	api := rg.Group("/users")
	api.Use(CheckAuth())

}
