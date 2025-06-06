package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Active   bool   `json:"active"`
}

var users []*User
var jwtKey = []byte("abcdefghijklmn")
var nextId int = 0

func GenerateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("could not parse claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", fmt.Errorf("username not found in token")
	}

	return username, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(400, gin.H{"error": "token missing"})
			c.Abort()
			return
		}

		username, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("username", username)
		c.Next()
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
				token, err := GenerateToken(v.Username)
				if err != nil {
					c.JSON(500, gin.H{"error": "could not generate token"})
					return
				}
				v.Active = true
				c.JSON(200, gin.H{"token": token})
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

	r.POST("/logout", AuthMiddleware(), func(c *gin.Context) {
		username := c.GetString("username")

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
