package handler

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// uploadRule 每种资源的限制配置
type uploadRule struct {
	MaxSize   int64    // 字节
	AllowExts []string // 如 {".jpg", ".png", ".webp"}
	Dir       string   // static 下的子目录
}

var uploadRules = map[string]uploadRule{
	"avatar": {2 << 20, []string{".jpg", ".jpeg", ".png", ".webp"}, "static/avatars"},
	"poster": {5 << 20, []string{".jpg", ".jpeg", ".png", ".webp"}, "static/posters"},
	"cover":  {3 << 20, []string{".jpg", ".jpeg", ".png", ".webp"}, "static/covers"},
}

// UploadHandler 统一上传
type UploadHandler struct{}

// Upload 统一上传入口
// FormData: file=图片文件, type=avatar|poster|cover
func (h *UploadHandler) Upload(c *gin.Context) {
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint)

	resourceType := c.PostForm("type")
	rule, ok := uploadRules[resourceType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的上传类型"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件"})
		return
	}
	defer file.Close()

	if header.Size > rule.MaxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("文件大小不能超过 %dMB", rule.MaxSize>>20),
		})
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := false
	for _, e := range rule.AllowExts {
		if ext == e {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件格式"})
		return
	}

	// 校验是否为真实图片（防改后缀）
	buf, err := io.ReadAll(io.LimitReader(file, rule.MaxSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	if _, _, err := image.DecodeConfig(strings.NewReader(string(buf))); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不是有效的图片文件"})
		return
	}

	// 确保目录存在
	if err := os.MkdirAll(rule.Dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器创建目录失败"})
		return
	}

	// 文件名：用户ID_时间戳_原文件名
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(rule.Dir, filename)

	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	url := "/" + strings.ReplaceAll(filePath, "\\", "/")
	c.JSON(http.StatusOK, gin.H{"url": url})
}
