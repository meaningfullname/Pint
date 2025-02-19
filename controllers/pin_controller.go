package controllers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"Pint/database"
	"Pint/models"
	"Pint/utils"
)

func CreatePin(c *gin.Context) {
	var input struct {
		Title string `json:"title"`
		Pin   string `json:"pin"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No file uploaded"})
		return
	}

	// Upload to Cloudinary
	cloudinaryResponse, err := utils.UploadToCloudinary(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error uploading file"})
		return
	}

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	pin := models.Pin{
		Title: input.Title,
		Pin:   input.Pin,
		Owner: userObjId,
		Image: models.Image{
			ID:  cloudinaryResponse.PublicID,
			URL: cloudinaryResponse.URL,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = database.PinsCollection.InsertOne(c, pin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating pin"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pin Created"})
}

func GetAllPins(c *gin.Context) {
	cursor, err := database.PinsCollection.Find(c, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching pins"})
		return
	}

	var pins []models.Pin
	if err = cursor.All(c, &pins); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error decoding pins"})
		return
	}

	c.JSON(http.StatusOK, pins)
}

func GetSinglePin(c *gin.Context) {
	id := c.Param("id")
	pinObjId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid pin ID"})
		return
	}

	var pin models.Pin
	err = database.PinsCollection.FindOne(c, bson.M{"_id": pinObjId}).Decode(&pin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pin not found"})
		return
	}

	c.JSON(http.StatusOK, pin)
}

func UpdatePin(c *gin.Context) {
	id := c.Param("id")
	pinObjId, _ := primitive.ObjectIDFromHex(id)

	var input struct {
		Title string `json:"title"`
		Pin   string `json:"pin"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	var pin models.Pin
	err := database.PinsCollection.FindOne(c, bson.M{"_id": pinObjId}).Decode(&pin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pin not found"})
		return
	}

	if pin.Owner != userObjId {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	_, err = database.PinsCollection.UpdateOne(c,
		bson.M{"_id": pinObjId},
		bson.M{"$set": bson.M{
			"title":     input.Title,
			"pin":       input.Pin,
			"updatedAt": time.Now(),
		}})

	c.JSON(http.StatusOK, gin.H{"message": "Pin updated"})
}

func DeletePin(c *gin.Context) {
	id := c.Param("id")
	pinObjId, _ := primitive.ObjectIDFromHex(id)

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	var pin models.Pin
	err := database.PinsCollection.FindOne(c, bson.M{"_id": pinObjId}).Decode(&pin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Pin not found"})
		return
	}

	if pin.Owner != userObjId {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized"})
		return
	}

	// Delete from Cloudinary
	if pin.Image.ID != "" {
		cld, err := cloudinary.NewFromParams(os.Getenv("CLOUD_NAME"), os.Getenv("CLOUD_API_KEY"), os.Getenv("CLOUD_API_SECRET"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error connecting to Cloudinary"})
			return
		}
		_, err = cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: pin.Image.ID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting image"})
			return
		}
	}

	_, err = database.PinsCollection.DeleteOne(c, bson.M{"_id": pinObjId})

	c.JSON(http.StatusOK, gin.H{"message": "Pin Deleted"})
}

func CommentOnPin(c *gin.Context) {
	id := c.Param("id")
	pinObjId, _ := primitive.ObjectIDFromHex(id)

	var input struct {
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	comment := models.Comment{
		ID:      primitive.NewObjectID(),
		User:    userObjId,
		Name:    c.MustGet("userName").(string),
		Comment: input.Comment,
	}

	_, err := database.PinsCollection.UpdateOne(c,
		bson.M{"_id": pinObjId},
		bson.M{"$push": bson.M{"comments": comment}})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error adding comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment Added"})
}

func DeleteComment(c *gin.Context) {
	pinId := c.Param("id")
	commentId := c.Query("commentId")

	pinObjId, _ := primitive.ObjectIDFromHex(pinId)
	commentObjId, _ := primitive.ObjectIDFromHex(commentId)

	userId, _ := c.Get("userId")
	userObjId, _ := primitive.ObjectIDFromHex(userId.(string))

	result, err := database.PinsCollection.UpdateOne(c,
		bson.M{
			"_id": pinObjId,
			"comments": bson.M{
				"$elemMatch": bson.M{
					"_id":  commentObjId,
					"user": userObjId,
				},
			},
		},
		bson.M{"$pull": bson.M{"comments": bson.M{"_id": commentObjId}}})

	if err != nil || result.ModifiedCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"message": "Unauthorized or comment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment Deleted"})
}
