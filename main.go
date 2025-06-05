package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int 
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var users []User

func main() {
	r := gin.Default()

	auth := r.Group("/api/auth")

	auth.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		fmt.Println(user)

		for _, v := range users {
			if v.Username == user.Username {
				c.JSON(400, gin.H{"error": "user already exists"})
			}
		}

		users = append(users, user)
		user.ID++
		c.JSON(200, gin.H{"success": "user created"})

		fmt.Println(users)
	})

	auth.POST("/login", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		for _, v := range users {
			if v.Username != user.Username {
				c.JSON(400, gin.H{"error": "user doesnot exists"})
			} else {
				c.JSON(200, gin.H{"success": "you are logged in"})
			}
		}

	})

	auth.GET("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": users})
	})

	r.Run()
}
