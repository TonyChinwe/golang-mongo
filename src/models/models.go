package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Title   string             `bson:"title,omitempty"`
	Authors []string           `bson:"authors,omitempty"`
	Isbn    string             `bson:"isbn,omitempty"`
}
