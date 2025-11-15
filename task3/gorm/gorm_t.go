package gorm_t

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string `gorm:"size:100;uniqueIndex;not null"`
	Name      string `gorm:"size:50;uniqueIndex;not null"`
	Password  string `gorm:"size:100;not null"`
	Posts     []Post `gorm:"foreignkey:UserID"`
	PostCount uint   `gorm:"default:0;not null"`
}

type Post struct {
	gorm.Model
	Title         string    `gorm:"size:200;not null"`
	Content       string    `gorm:"type:text;not null"`
	UserID        uint      `gorm:"not null"`
	User          User      `gorm:"foreignkey:UserID"`
	Comments      []Comment `gorm:"foreignkey:PostID"`
	CommentStatus string    `gorm:"size:50"`
}

// 文章创建时自动更新用户的文章数量统计字段
func (post *Post) BeforeCreate(tx *gorm.DB) error {
	return tx.Model(&User{}).
		Where("id = ?", post.UserID).
		Update("post_count", gorm.Expr("post_count + ?", 1)).
		Error
}

type Comment struct {
	gorm.Model
	Content string `gorm:"type:text;not null"`
	PostID  uint   `gorm:"not null"`
	Post    Post   `gorm:"foreignkey:PostID"`
	UserID  uint   `gorm:"not null"`
	User    User   `gorm:"foreignkey:UserID"`
}

// 评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"
func (comment *Comment) AfterDelete(tx *gorm.DB) error {
	var commentCount int64
	if err := tx.Model(&Comment{}).
		Where("post_id = ?", comment.PostID).
		Count(&commentCount).Error; err != nil {
		return err
	}

	if commentCount == 0 {
		if err := tx.Model(&Post{}).
			Where("id = ?", comment.PostID).
			Update("comment_status", "无评论").Error; err != nil {
			return err
		}
	}
	return nil
}

// 查询某个用户发布的所有文章及其对应的评论信息。
func getUserPostComments(db *gorm.DB, name string) ([]User, error) {
	var users []User
	result := db.Preload("Posts").Preload("Posts.Comments").Where("name = ?", name).Find(&users)
	return users, result.Error
}

// 查询评论数量最多的文章信息
func getPostWithMostComments(db *gorm.DB) (Post, error) {
	var post Post
	commemtSubQuery := db.Model(&Comment{}).
		Select("post_id, COUNT(comments.id) as comment_count").
		Group("post_id")
	err := db.Model(&Post{}).
		Select("posts.*, IFNULL(sub.comment_count, 0) as comment_count").
		Joins("LEFT JOIN (?) as sub ON posts.id = sub.post_id", commemtSubQuery).
		Order("comment_count DESC").
		First(&post).Error
	return post, err
}

func Run() {
	dsn := "root:root@tcp(127.0.0.1:3306)/goprac?charset=utf8mb4&parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败：%v", err)
		return
	}
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{})
	if err != nil {
		log.Fatalf("创建数据表失败：%v", err)
	}

	users, _ := getUserPostComments(db, "张三")
	for _, user := range users {
		fmt.Println(user.Name)
		for _, post := range user.Posts {
			fmt.Println(post.Title)
			for _, comment := range post.Comments {
				fmt.Println(comment)
			}
		}
	}

	post, _ := getPostWithMostComments(db)
	fmt.Println(post)
}
