package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/usecases/project"
)

type GetHandler struct {
	usecase *project.GetUsecase
}

func NewGetHandler(usecase *project.GetUsecase) *GetHandler {
	return &GetHandler{usecase: usecase}
}

func (h *GetHandler) Handle(c *gin.Context) {
	id := c.Param("id")

	project, err := h.usecase.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}
