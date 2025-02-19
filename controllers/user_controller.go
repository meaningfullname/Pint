package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"Pint/database"
	"Pint/models"
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already registered"})
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

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  result.InsertedID.(primitive.ObjectID).Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SEC")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating token"})
		return
	}

	c.SetCookie("token", tokenString, 15*24*60*60*1000, "/", "", false, true)

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

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Wrong password"})
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID.Hex(),
		"exp": time.Now().Add(time.Hour * 24 * 15).Unix(), // 15 days
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SEC")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating token"})
		return
	}

	c.SetCookie("token", tokenString, 15*24*60*60*1000, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "Logged in",
	})
}

func LogoutUser(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged Out Successfully"})
}

func MyProfile(c *gin.Context) {
	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	var user models.User
	err := database.UsersCollection.FindOne(c, bson.M{"_id": userObjId}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func UserProfile(c *gin.Context) {
	id := c.Param("id")
	userObjId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}

	var user models.User
	err = database.UsersCollection.FindOne(c, bson.M{"_id": userObjId}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func FollowUser(c *gin.Context) {
	targetId := c.Param("id")
	targetObjId, _ := primitive.ObjectIDFromHex(targetId)

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	if targetId == userId.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You can't follow yourself"})
		return
	}

	var currentUser models.User

	// Check if already following
	err := database.UsersCollection.FindOne(c, bson.M{
		"_id":       userObjId,
		"following": targetObjId,
	}).Decode(&currentUser)

	if err == nil {
		// Unfollow
		_, err = database.UsersCollection.UpdateOne(c,
			bson.M{"_id": userObjId},
			bson.M{"$pull": bson.M{"following": targetObjId}})

		_, err = database.UsersCollection.UpdateOne(c,
			bson.M{"_id": targetObjId},
			bson.M{"$pull": bson.M{"followers": userObjId}})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error unfollowing user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User Unfollowed"})
	} else {
		// Follow
		_, err = database.UsersCollection.UpdateOne(c,
			bson.M{"_id": userObjId},
			bson.M{"$push": bson.M{"following": targetObjId}})

		_, err = database.UsersCollection.UpdateOne(c,
			bson.M{"_id": targetObjId},
			bson.M{"$push": bson.M{"followers": userObjId}})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error following user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User Followed"})
	}
}
