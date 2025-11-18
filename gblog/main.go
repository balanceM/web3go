package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 初始化数据库操作对象
func initDB() *gorm.DB {
	dsn := "root:liu123@tcp(127.0.0.1:3306)/gblog?charset=utf8mb4&parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Init db failed!")
	}
	db.AutoMigrate(&User{}, &Post{}, &Comment{})
	return db
}

var db = initDB()

func main() {
	r := gin.Default()
	r.POST("/register", PasswordEncrypt(), registerHandler)
	r.POST("/login", loginHandler)

	auth := r.Group("/auth")
	auth.Use(JwtAuthMiddleware())

	auth.POST("/post", CreatePostHandler)
	auth.PUT("/post/:id", UpdatePostHandler)
	auth.GET("/post/:id", GetPostHandler)
	auth.DELETE("/post/:id", DeletePostHandler)

	auth.POST("/post/:post_id/comment", CreateCommentHandler)
	auth.GET("/post/:post_id/comments", GetCommentsByPostID)

	r.Run(":8080")
}
