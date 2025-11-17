package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique" form:"username" binding:"required"`
	Password string `form:"password" bingding:"required"`
	Email    string `form:"email"`
}

type Post struct {
	gorm.Model
	Title   string
	Context string
	UserID  uint
	User    User
}

type Comment struct {
	gorm.Model
	Content string
	UserID  uint
	User    User
	PostID  uint
	Post    Post
}

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

// 密码加密中间件
func PasswordEncrypt() gin.HandlerFunc {
	return func(c *gin.Context) {
		password := c.PostForm("password")
		if password == "" {
			c.Next()
			return
		}
		// 加密
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "password encrypted failed!",
			})
			c.Abort()
			return
		}
		// 先调用 ParseMultipartForm 解析, 否则可能无法正确获取字段
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(500, gin.H{
				"error": "Parse form failed!",
			})
			c.Abort()
			return
		}
		// 重新设置password
		c.Request.PostForm.Set("password", string(hashedPassword))
		c.Next()
	}
}

// 用户注册
func registerHandler(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusAccepted, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func main() {
	r := gin.Default()
	r.POST("/register", PasswordEncrypt(), registerHandler)

	r.Run(":8080")
}
