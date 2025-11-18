package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	Content string
	UserID  uint
	User    User
	PostID  uint
	Post    Post
}

type CreateCommentReq struct {
	Content string
	PostID  uint
}

func CreateCommentHandler(c *gin.Context) {
	pidStr := c.Param("post_id")
	if pidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post id is null"})
		return
	}
	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id format is not correct"})
		return
	}

	uid, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	comment := &Comment{
		Content: c.PostForm("content"),
		UserID:  uid,
		PostID:  uint(pid),
	}
	if err := db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"comment": gin.H{
			"content": comment.Content,
			"post_id": comment.PostID,
			"user_id": comment.UserID,
		},
	})
}

func GetCommentsByPostID(c *gin.Context) {
	pidStr := c.Param("post_id")
	if pidStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post id is null"})
		return
	}
	pid, err := strconv.ParseUint(pidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "post_id format is not correct"})
		return
	}

	var comments []Comment
	if err := db.Preload("Post").Where("post_id = ?", pid).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"comments": comments,
	})
}
