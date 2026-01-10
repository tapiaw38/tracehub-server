package ai

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/tapiaw38/tracehub-server/internal/config"
	"github.com/tapiaw38/tracehub-server/pkg/models"
)

// Client is a client for interacting with Claude AI
type Client struct {
	client  *anthropic.Client
	config  *config.ClaudeConfig
	context *ContextBuilder
}

// NewClient creates a new AI client
func NewClient(cfg *config.ClaudeConfig, repoPath string) *Client {
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)

	return &Client{
		client:  &client,
		config:  cfg,
		context: NewContextBuilder(repoPath),
	}
}

// AnalyzeError analyzes an error and returns a fix proposal
func (c *Client) AnalyzeError(ctx context.Context, err *models.DetectedError) (*models.FixProposal, error) {
	log.Printf("Analyzing error %s with Claude AI...", err.ID)

	// Build context
	codeSnippet, ctxErr := c.context.BuildContext(err)
	if ctxErr != nil {
		log.Printf("Warning: failed to build context: %v", ctxErr)
		codeSnippet = "// Code snippet not available"
	}

	// Generate prompt
	prompt := AnalysisPrompt(err, codeSnippet)

	// Call Claude API
	response, apiErr := c.callClaude(ctx, prompt)
	if apiErr != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", apiErr)
	}

	// Parse response
	claudeResp, parseErr := ParseClaudeResponse(response)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w", parseErr)
	}

	// Create fix proposal
	proposal := &models.FixProposal{
		ID:          generateProposalID(err),
		ErrorID:     err.ID,
		ServiceName: err.ServiceName,
		RootCause:   claudeResp.RootCause,
		Explanation: claudeResp.Explanation,
		Fix:         claudeResp.Fix,
		Tests:       claudeResp.Tests,
		Prevention:  claudeResp.Prevention,
		CreatedAt:   time.Now(),
		Status:      models.ProposalPending,
	}

	log.Printf("Successfully analyzed error %s", err.ID)
	return proposal, nil
}

// callClaude calls the Claude API with the given prompt
func (c *Client) callClaude(ctx context.Context, prompt string) (string, error) {
	maxTokens := int64(c.config.MaxTokens)

	systemPrompt := SystemPrompt()

	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(c.config.Model),
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		System: []anthropic.TextBlockParam{
			{
				Type: "text",
				Text: systemPrompt,
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("API call failed: %w", err)
	}

	// Extract text from response
	if len(message.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	// Get the text content
	content := message.Content[0]
	if content.Text != "" {
		return content.Text, nil
	}

	return "", fmt.Errorf("unexpected response format from Claude")
}

// generateProposalID generates a unique ID for a fix proposal
func generateProposalID(err *models.DetectedError) string {
	return fmt.Sprintf("fix-%s-%d", err.ID, time.Now().Unix())
}
