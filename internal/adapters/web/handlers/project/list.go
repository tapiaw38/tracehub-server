package project

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/usecases/project"
)

type ListHandler struct {
	usecase *project.ListUsecase
}

func NewListHandler(usecase *project.ListUsecase) *ListHandler {
	return &ListHandler{usecase: usecase}
}

func (h *ListHandler) Handle(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	output, err := h.usecase.Execute(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
