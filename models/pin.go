package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Pin struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Pin       string             `bson:"pin" json:"pin"`
	Owner     primitive.ObjectID `bson:"owner" json:"owner"`
	Image     Image              `bson:"image" json:"image"`
	Comments  []Comment          `bson:"comments" json:"comments"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Image struct {
	ID  string `bson:"id" json:"id"`
	URL string `bson:"url" json:"url"`
}

type Comment struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	User    primitive.ObjectID `bson:"user" json:"user"`
	Name    string             `bson:"name" json:"name"`
	Comment string             `bson:"comment" json:"comment"`
}
