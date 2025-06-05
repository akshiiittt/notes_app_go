package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Active   bool   `json:"active"`
}

var users []*User
var nextId int = 0

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")

		if username == "" {
			c.JSON(400, gin.H{"error": "no username is provided"})
			c.Abort()
			return
		}

		for _, v := range users {
			if v.Username == username && v.Active {
				c.Next()
				return
			}
		}

		c.JSON(403, gin.H{"error": " username is not active"})
		c.Abort()
	}
}

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
		users = append(users, &user)
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
				v.Active = true
				c.JSON(200, gin.H{"success": "you are logged in"})
				return
			}
		}

		c.JSON(400, gin.H{"error": "user doesnot exists"})
	})

	auth.GET("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": users})
	})

	protected := r.Group("/api/protected")

	protected.GET("/notes", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"notes": []string{"note1", "note2"}})
	})

	r.POST("/logout", func(c *gin.Context) {
		username := c.GetHeader("Username")

		if username == "" {
			c.JSON(400, gin.H{"error": "no username is provided"})
			c.Abort()
			return
		}

		for _, v := range users {
			if v.Username == username {
				if v.Active {
					v.Active = false
					c.JSON(200, gin.H{"success": "user logged out"})
					return
				} else {
					c.JSON(400, gin.H{"success": "user is already logged out"})
					return
				}
			}
		}

		c.JSON(404, gin.H{"error": "user not found"})

	})

	r.Run()
}
