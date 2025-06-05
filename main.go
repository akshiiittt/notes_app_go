package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var users []User
var nextId int = 0

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
				return
			}
		}

		user.ID = nextId
		nextId++
		users = append(users, user)
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
			if v.Username == user.Username && v.Password == user.Password {
				c.JSON(200, gin.H{"success": "you are logged in"})
				return
			}
		}

		c.JSON(400, gin.H{"error": "user doesnot exists"})
	})

	auth.GET("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": users})
	})

	r.Run()
}
