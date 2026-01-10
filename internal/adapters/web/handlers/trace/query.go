package trace

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/usecases/trace"
)

type QueryHandler struct {
	usecase *trace.QueryUsecase
}

func NewQueryHandler(usecase *trace.QueryUsecase) *QueryHandler {
	return &QueryHandler{usecase: usecase}
}

func (h *QueryHandler) Handle(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input trace.QueryInput
	if err := c.ShouldBindQuery(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ProjectID = projectID.(string)

	output, err := h.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
