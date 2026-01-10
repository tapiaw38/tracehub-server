package trace

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/usecases/trace"
)

type IngestHandler struct {
	usecase *trace.IngestUsecase
}

func NewIngestHandler(usecase *trace.IngestUsecase) *IngestHandler {
	return &IngestHandler{usecase: usecase}
}

func (h *IngestHandler) Handle(c *gin.Context) {
	// Get project ID from context (set by middleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input trace.IngestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Execute(c.Request.Context(), projectID.(string), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "trace ingested"})
}

func (h *IngestHandler) HandleBatch(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input trace.IngestBatchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.ExecuteBatch(c.Request.Context(), projectID.(string), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "traces ingested",
		"count":  len(input.Traces),
	})
}
