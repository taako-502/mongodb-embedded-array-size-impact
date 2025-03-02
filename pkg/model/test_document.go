package model

import "go.mongodb.org/mongo-driver/v2/bson"

type TestDocument struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	Objects       []NestedObject `bson:"objects"`
	SizeInBytes   int            `bson:"sizeInBytes"`
	InsertionTime string         `bson:"insertionTime"`
	RetrievalTime int64          `bson:"retrievalTime"`
}

type NestedObject struct {
	Order int             `bson:"order"`
	Data  string          `bson:"data"`
	Bool  bool            `bson:"bool"`
	IDs   []bson.ObjectID `bson:"ids"`
}
