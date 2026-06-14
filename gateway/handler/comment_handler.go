package handler

import (
	"net/http"
	"ourvideos/proto/user"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"ourvideos/proto/comment"
)

type CommentHandler struct {
	Client     comment.CommentServiceClient
	UserClient user.UserServiceClient
}

func (h CommentHandler) AddComments(c *gin.Context) {
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint)
	var req comment.AddCommentsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	//调 user-service 拿当前用户的 username 和 avatar
	userResp, err := h.UserClient.GetUser(c.Request.Context(), &user.GetUserReq{Id: uint64(userID)})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}

	resp, err := h.Client.AddComments(c.Request.Context(), &comment.AddCommentsReq{
		Uid:       uint64(userID),
		VideoId:   req.VideoId,
		Content:   req.Content,
		Username:  userResp.Username,
		Avatar:    userResp.Avatar,
		CreateAt:  time.Now().Format("2006-01-02 15:04:05"),
		LikeCount: 0,
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"new_comment": resp.Comment,
	})
}

func (h CommentHandler) ListComments(c *gin.Context) {
	videoId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的视频 ID"})
		return
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "latest")

	resp, err := h.Client.ListComments(c.Request.Context(), &comment.ListCommentsReq{
		VideoId: videoId,
		Offset:  int32(offset),
		Limit:   int32(limit),
		Sort:    sort,
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"comments": resp.Comment,
		"total":    resp.Total,
	})
}

func (h CommentHandler) LikeComment(c *gin.Context) {
	commentId, err := strconv.ParseUint(c.Param("id"), 10, 64)

	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID := userIDVal.(uint)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Client.LikeComment(c.Request.Context(), &comment.LikeCommentReq{
		VideoId:   0,
		Uid:       uint64(userID),
		CommentId: commentId,
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"like_count": resp.LikeCount,
	})
}
