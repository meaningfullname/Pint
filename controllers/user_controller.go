package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"your-proj/models"
	"your-project/database"
)

func RegisterUser(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Check if user exists
	var existingUser models.User
	err := database.UsersCollection.FindOne(c, bson.M{"email": input.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Already have an account with this email"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error hashing password"})
		return
	}

	// Create user
	user := models.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := database.UsersCollection.InsertOne(c, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating user"})
		return
	}

	// Generate token
	generateToken(c, result.InsertedID.(primitive.ObjectID))

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "User Registered",
	})
}

func LoginUser(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	var user models.User
	err := database.UsersCollection.FindOne(c, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No user with this email"})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Wrong password"})
		return
	}

	generateToken(c, user.ID)

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "Logged in",
	})
}

func generateToken(c *gin.Context, userId primitive.ObjectID) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userId.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SEC")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating token"})
		return
	}

	c.SetCookie("token", tokenString, 15*24*60*60*1000, "/", "", false, true)
} 