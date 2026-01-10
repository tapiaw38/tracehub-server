package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/usecases/project"
)

type CreateHandler struct {
	usecase *project.CreateUsecase
}

func NewCreateHandler(usecase *project.CreateUsecase) *CreateHandler {
	return &CreateHandler{usecase: usecase}
}

func (h *CreateHandler) Handle(c *gin.Context) {
	var input project.CreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, output)
}
