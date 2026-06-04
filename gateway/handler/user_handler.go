package handler

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"net/http"
	"ourvideos/proto/user"
	"strconv"
)

type UserHandler struct {
	Client user.UserServiceClient
}

func (h *UserHandler) Register(c *gin.Context) {
	var req user.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return
	}
	resp, err := h.Client.Register(c.Request.Context(), &req)
	if err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    resp.Id,
		"token": resp.Token,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req user.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return
	}
	resp, err := h.Client.Login(c.Request.Context(), &req)
	if err != nil {
		//提取grpc错误信息
		st := status.Convert(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":   resp.Token,
		"user_id": resp.UserId,
	})

}

func (h *UserHandler) GetUser(c *gin.Context) {
	//从上下文中取出userID
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	currentUserID := userIDVal.(uint)
	//从url路径中提取想要的Id
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return
	}
	if uint(id) != currentUserID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权查看其他用户信息",
		})
		return
	}

	req := &user.GetUserReq{Id: id}
	resp, err := h.Client.GetUser(c.Request.Context(), req)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       resp.Id,
		"username": resp.Username,
		"email":    resp.Email,
		"nickname": resp.Nickname,
		"avatar":   resp.Avatar,
		"bio":      resp.Bio,
		"gender":   resp.Gender,
		"phone":    resp.Phone,
	})

}

func (h *UserHandler) UserUpdate(c *gin.Context) {
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未登录"})
		return
	}
	currentUserID := userIDVal.(uint)

	var req user.UserUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return

	}

	req.Id = uint64(currentUserID)
	resp, err := h.Client.UserUpdate(c.Request.Context(), &req)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       resp.Id,
		"username": resp.Username,
		"email":    resp.Email,
		"avatar":   resp.Avatar,
		"bio":      resp.Bio,
		"gender":   resp.Gender,
		"phone":    resp.Phone,
	})

}

func (h *UserHandler) PswUpdate(c *gin.Context) {
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先登录"})
		return
	}
	currentUserID := userIDVal.(uint)
	var req user.PswUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		st := status.Convert(err)
		c.JSON(toHTTPStatusCode(st.Code()), gin.H{"error": st.Message()})
		return
	}
	req.Id = uint32(currentUserID)
	resp, err := h.Client.PswUpdate(c.Request.Context(), &req)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": resp.Id})
}
