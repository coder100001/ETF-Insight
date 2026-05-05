package handlers

import (
	"net/http"

	"etf-insight/services/agent"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	client *agent.Client
}

func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		client: agent.NewClient(),
	}
}

func (h *AgentHandler) Health(c *gin.Context) {
	result, err := h.client.Health()
	if err != nil {
		utils.Error("Agent service health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Agent service unavailable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AgentHandler) Discover(c *gin.Context) {
	agents, err := h.client.Discover()
	if err != nil {
		utils.Error("Failed to discover agents", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to discover agents",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": agents})
}

func (h *AgentHandler) Run(c *gin.Context) {
	var req agent.AgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if req.AgentID == "" || req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "agent_id and query are required"})
		return
	}

	if req.LLMProvider == "" {
		req.LLMProvider = "openai"
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}

	result, err := h.client.Run(req)
	if err != nil {
		utils.Error("Agent execution failed", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "Agent execution failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

func (h *AgentHandler) RunTeam(c *gin.Context) {
	var req agent.AgentTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request format"})
		return
	}

	if len(req.AgentIDs) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "At least 2 agents required for team"})
		return
	}

	if req.LLMProvider == "" {
		req.LLMProvider = "openai"
	}
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	if req.Rounds == 0 {
		req.Rounds = 1
	}

	result, err := h.client.RunTeam(req)
	if err != nil {
		utils.Error("Team execution failed", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "Team execution failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
