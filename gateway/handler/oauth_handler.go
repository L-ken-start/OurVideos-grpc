package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"ourvideos/proto/user"
	"strings"
	"time"
)

type OAuthHandler struct {
	UserClient user.UserServiceClient
	RDB        *redis.Client
	ClientID   string
	Secret     string
	Callback   string
	Frontend   string
}

// GitHubAuthorize 第一步：重定向到 GitHub 授权页
func (h *OAuthHandler) GitHubAuthorize(c *gin.Context) {
	state := randomString(16)
	ctx := c.Request.Context()
	if err := h.RDB.Set(ctx, "oauth:state:"+state, "1", 5*time.Minute).Err(); err != nil {
		log.Printf("[OAuth] Set state失败: %v", err)
	}

	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email&state=%s",
		h.ClientID, url.QueryEscape(h.Callback), state,
	)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubCallback 第二步：GitHub 回调，code 换 token 换用户信息，签发 JWT
func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// 1. 验 state 防 CSRF
	ctx := c.Request.Context()
	log.Printf("[OAuth] state=%s", state)
	if err := h.RDB.GetDel(ctx, "oauth:state:"+state).Err(); err != nil {
		log.Printf("[OAuth] state验证失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
		return
	}

	// 2. code → access_token
	data := url.Values{
		"client_id":     {h.ClientID},
		"client_secret": {h.Secret},
		"code":          {code},
		"redirect_uri":  {h.Callback},
	}

	req, err := http.NewRequest("POST",
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 token 失败"})
		return
	}
	req.Header.Set("Accept", "application/json") // 要求 GitHub 返回 JSON 格式
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 token 失败"})
		return
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &tokenResp)
	log.Printf("[OAuth] access_token返回: status=%d body=%s", resp.StatusCode, string(body))

	// ========== 步骤 3：access_token → 用户信息 ==========
	req, _ = http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID        int64  `json:"id"`         // GitHub 用户唯一数字 ID
		Login     string `json:"login"`      // GitHub 用户名
		AvatarURL string `json:"avatar_url"` // 头像 URL
	}

	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &ghUser)
	log.Printf("[OAuth] 用户信息: id=%d login=%s", ghUser.ID, ghUser.Login)

	// ========== 步骤 4：拼 openID ==========
	// "gh_" 前缀区分来源，将来 QQ 是 "qq_xxx"，微信是 "wx_xxx"
	openID := fmt.Sprintf("gh_%d", ghUser.ID)

	// ========== 步骤 5：调 user-service 注册或登录 ==========
	// user-service 内部会查这个 openID 存在不存在
	// 不存在 → 自动创建用户
	// 存在 → 直接登录
	// 不管哪种情况，都返回 JWT token
	log.Printf("[OAuth] 调用user-service OauthLogin: openID=%s nickname=%s", openID, ghUser.Login)
	userResp, err := h.UserClient.OauthLogin(context.Background(), &user.OAuthLoginReq{
		Provider: "github",
		OpenId:   openID,
		Nickname: ghUser.Login,
		Avatar:   ghUser.AvatarURL,
	})
	if err != nil {
		log.Printf("[OAuth] OauthLogin gRPC失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	// ========== 步骤 6：重定向回前端 ==========
	// 把 JWT token 通过 URL 参数传回前端
	// 生产环境建议用 Set-Cookie 更安全
	c.Redirect(http.StatusTemporaryRedirect, h.Frontend+"/auth/callback?token="+userResp.Token)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)

}
