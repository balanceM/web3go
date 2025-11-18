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
	Password string `form:"password" binding:"required"`
	Email    string `form:"email"`
}

type LoginUser struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
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
		// 先调用 ParseMultipartForm 解析, 否则可能无法正确获取字段
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.JSON(500, gin.H{
				"error": "Parse form failed!",
			})
			c.Abort()
			return
		}
		// 获取密码字段
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
		// 重新设置password
		// c.Request.PostForm.Set("password", string(hashedPassword)) //这种方式，
		// 修改请求里的form数据无用，因为后续方法读取的form值来自于原始请求

		//存储加密后密码
		c.Set("hashedPassword", string(hashedPassword))
		c.Next()
	}
}

// 用户注册
func registerHandler(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	// 从上下文获取加密后的密码
	hashedPassword, exists := c.Get("hashedPassword")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password not encrypted"})
		return
	}
	user.Password = hashedPassword.(string)
	// 创建
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}
	// 生成token
	token, err := GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token generate failed"})
		return
	}
	// 返回
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
	})
}

// 登录
func loginHandler(c *gin.Context) {
	var user User
	username := c.PostForm("username")
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not exist"})
		return
	}
	// 比较密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.PostForm("password")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is not correct"})
		return
	}
	// 生成token
	token, err := GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token generate failed"})
		return
	}
	// 返回
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func main() {
	r := gin.Default()
	r.POST("/register", PasswordEncrypt(), registerHandler)
	r.POST("/login", loginHandler)

	r.Run(":8080")
}
