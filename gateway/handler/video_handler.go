package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ourvideos/proto/video"
)

type VideoHandler struct {
	Client video.VideoServiceClient
}

func (h *VideoHandler) CreateVideo(c *gin.Context) {
	var req video.CreateVideoReq
	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先登录"})
	}
	userId := userIDVal.(uint)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	req.UserId = int32(userId)
	resp, err := h.Client.CreateVideo(c.Request.Context(), &req)
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"video": resp.Video}) //service和repo还没改
}

func (h *VideoHandler) UploadVideo(c *gin.Context) {
	var req video.UploadVideoReq
	_, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先登录"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Client.UploadVideo(c.Request.Context(), &req)
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"video": resp.Video})

}

func (h *VideoHandler) GetVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的视频 ID"})
		return
	}

	userIDVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	currentUserID := userIDVal.(uint)
	fmt.Println(currentUserID)
	resp, err := h.Client.GetVideo(c.Request.Context(), &video.GetVideoReq{Id: id, Uid: uint64(currentUserID)})
	if err != nil {
		HandleGRPCError(c, err) //get里面没放token
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"video":    resp.Video,
		"is_liked": resp.IsLiked,
	})
}

func (h *VideoHandler) ListSeriesEpisodes(c *gin.Context) {
	seriesID, _ := strconv.ParseUint(c.Param("series_id"), 10, 64)
	resp, err := h.Client.ListSeriesEpisodes(c.Request.Context(), &video.SeriesEpisodesReq{SeriesId: seriesID})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"episodes": resp.Episodes})
}

func (h *VideoHandler) ListVideo(c *gin.Context) {
	category := c.DefaultQuery("category", "")
	tag := c.DefaultQuery("tag", "")
	sortBy := parseSortBy(c.DefaultQuery("sort", "latest"))
	fmt.Println(sortBy)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.Client.ListVideo(c.Request.Context(), &video.ListVideoReq{
		Category: category,
		SortBy:   sortBy,
		Offset:   int32(offset),
		Limit:    int64(limit),
		Tag:      tag,
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"videos": resp.Videos, "total": resp.Total})
}

func (h *VideoHandler) SearchVideos(c *gin.Context) {
	query := c.DefaultQuery("q", "")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.Client.SearchVideo(c.Request.Context(), &video.SearchVideosReq{
		Query:  query,
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"videos": resp.Videos, "total": resp.Total})
}

func (h *VideoHandler) LikeVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	userIdVal, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	userID := userIdVal.(uint)
	resp, err := h.Client.LikeVideo(c.Request.Context(), &video.LikeVideoReq{
		VideoId:   id,
		Uid:       uint64(userID),
		CommentId: 0,
	})
	if err != nil {
		HandleGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"like_count": resp.LikeCount,
		"is_liked":   resp.IsLiked,
	})
}

func parseSortBy(s string) video.SortBy {
	switch s {
	case "popular":
		return video.SortBy_SORT_BY_POPULAR
	case "rating":
		return video.SortBy_SORT_BY_RATING
	default:
		return video.SortBy_SORT_BY_LATEST
	}
}
