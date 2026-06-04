package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

var jwtSecret = []byte("ourvideos-secret-key")

// JWTAuth中间件:验证请求头中的bearer token并将user_id存入上下文
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未查询到token信息",
			})
			c.Abort() //阻止后续处理
			return

		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return

		}
		tokenString := parts[1]

		//解析token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			//验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token无效或过期"})
			c.Abort()
			return
		}

		//提取claims中的user_id
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token载荷无效"})
			c.Abort()
			return
		}

		userID, exists := claims["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token缺少用户标识"})
			c.Abort()
			return
		}
		//转换为uint,gin上下文中设置的值可以后续用get获取
		switch v := userID.(type) {
		case float64:
			c.Set("userID", uint(v))
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token格式异常"})
			c.Abort()
			return
		}

		//执行后续handler
		c.Next()
	}
}
