package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

		tokenString = strings.TrimPrefix(tokenString, "Bearer")

		if tokenString == "" {
			c.JSON(400, gin.H{"status": "error", "error": "token missing", "data": nil})
			c.Abort()
			return
		}

		username, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(400, gin.H{"status": "error", "error": err.Error(), "data": nil})
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
			c.JSON(400, gin.H{"status": "error", "error": err.Error(), "data": nil})
			return
		}
		fmt.Println(user)

		for _, v := range users {
			if v.Username == user.Username {
				c.JSON(400, gin.H{"status": "error", "error": "user already exists", "data": nil})
				return
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "error": "could not hash password", "data": nil})
			return
		}

		user.Password = string(hashedPassword)

		user.ID = nextId
		nextId++
		users = append(users, &user)
		c.JSON(200, gin.H{"status": "success", "data": gin.H{"message": "user created"}, "error": nil})

		fmt.Println(users)
	})

	auth.POST("/login", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(400, gin.H{"status": "error", "error": err.Error(), "data": nil})
			return
		}

		for _, v := range users {
			if v.Username == user.Username {
				if err := bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(user.Password)); err == nil {
					token, err := GenerateToken(v.Username)
					if err != nil {
						c.JSON(500, gin.H{"status": "error", "error": "could not generate token", "data": nil})
						return
					}
					v.Active = true
					c.JSON(200, gin.H{"status": "success", "data": gin.H{"token": token}, "error": nil})
					return
				}

			}
		}

		c.JSON(400, gin.H{"status": "error", "error": "user doesn't exist or invalid password", "data": nil})
	})

	auth.GET("/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success", "data": gin.H{"users": users}, "error": nil})
	})

	protected := r.Group("/api/protected")

	protected.GET("/notes", AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success", "data": gin.H{"notes": []string{"note1", "note2"}}, "error": nil})
	})

	r.POST("/logout", AuthMiddleware(), func(c *gin.Context) {
		username := c.GetString("username")

		if username == "" {
			c.JSON(400, gin.H{"status": "error", "error": "no username is provided", "data": nil})
			c.Abort()
			return
		}

		for _, v := range users {
			if v.Username == username {
				if v.Active {
					v.Active = false
					c.JSON(200, gin.H{"status": "success", "data": gin.H{"message": "user logged out"}, "error": nil})
					return
				} else {
					c.JSON(400, gin.H{"status": "error", "error": "user is already logged out", "data": nil})
					return
				}
			}
		}

		c.JSON(404, gin.H{"status": "error", "error": "user not found", "data": nil})

	})

	r.Run()
}
