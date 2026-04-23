package handlers

import (
	"net/http"
	"time"

	"etf-insight/config"
	"etf-insight/middleware"
	"etf-insight/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authMiddleware *middleware.AuthMiddleware
	users          map[string]*UserCredentials
	tokenExpiry    time.Duration
}

type UserCredentials struct {
	ID          string
	Username    string
	Password    string
	Role        string
	Permissions []string
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	User      UserInfo  `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func NewAuthHandler(cfg *config.JWTConfig) *AuthHandler {
	expiry := time.Duration(cfg.ExpiryHours) * time.Hour
	if expiry <= 0 {
		expiry = 24 * time.Hour
	}

	return &AuthHandler{
		authMiddleware: middleware.NewAuthMiddleware(cfg),
		tokenExpiry:    expiry,
		users: map[string]*UserCredentials{
			"admin": {
				ID:          "1",
				Username:    "admin",
				Password:    "admin123",
				Role:        "admin",
				Permissions: []string{"logs_view", "etf_view", "portfolio_view", "config_manage"},
			},
			"user": {
				ID:          "2",
				Username:    "user",
				Password:    "user123",
				Role:        "user",
				Permissions: []string{"logs_view", "etf_view", "portfolio_view"},
			},
		},
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误",
		})
		return
	}

	user, exists := h.users[req.Username]
	if !exists || user.Password != req.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户名或密码错误",
		})
		return
	}

	token, err := h.authMiddleware.GenerateToken(user.ID, user.Username, user.Role, user.Permissions)
	if err != nil {
		utils.Error("生成Token失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "登录失败，请稍后重试",
		})
		return
	}

	expiresAt := time.Now().Add(h.tokenExpiry)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": LoginResponse{
			Token: token,
			User: UserInfo{
				ID:          user.ID,
				Username:    user.Username,
				Role:        user.Role,
				Permissions: user.Permissions,
			},
			ExpiresAt: expiresAt,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "登出成功",
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "未登录",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户身份无效",
		})
		return
	}

	usernameValue, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户信息缺失",
		})
		return
	}
	username, ok := usernameValue.(string)
	if !ok || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "用户名无效",
		})
		return
	}

	roleValue, ok := c.Get("role")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "角色信息缺失",
		})
		return
	}
	role, ok := roleValue.(string)
	if !ok || role == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "角色信息无效",
		})
		return
	}

	permissionsValue, ok := c.Get("permissions")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "权限信息缺失",
		})
		return
	}
	permissions, ok := permissionsValue.([]string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "权限信息无效",
		})
		return
	}

	token, err := h.authMiddleware.GenerateToken(
		userIDStr,
		username,
		role,
		permissions,
	)
	if err != nil {
		utils.Error("刷新Token失败", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "刷新Token失败",
		})
		return
	}

	expiresAt := time.Now().Add(h.tokenExpiry)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":      token,
			"expires_at": expiresAt,
		},
	})
}
