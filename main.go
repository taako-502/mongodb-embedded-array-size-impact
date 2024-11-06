package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NestedObject struct {
	Data string `bson:"data"`
}

type TestDocument struct {
	ID      string         `bson:"_id,omitempty"`
	Objects []NestedObject `bson:"objects"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("testdb").Collection("testcollection")

	objectCounts := generateFibonacciUpTo(100)

	for _, count := range objectCounts {
		doc := createTestDocument(count)

		insertedDoc, err := collection.InsertOne(context.TODO(), doc)
		if err != nil {
			log.Fatalf("Failed to insert document: %v", err)
		}

		start := time.Now()
		err = retrieveDocument(collection, insertedDoc.InsertedID)
		if err != nil {
			log.Fatalf("Failed to retrieve document: %v", err)
		}
		duration := time.Since(start)

		fmt.Printf("Objects count: %d, Retrieval time: %v\n", count, duration)

		_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": insertedDoc.InsertedID})
		if err != nil {
			log.Fatalf("Failed to delete document: %v", err)
		}
	}
}

// generateFibonacciUpTo は指定された最大値までのフィボナッチ数列を生成します
func generateFibonacciUpTo(max int) []int {
	fibSequence := []int{0, 1}
	for {
		next := fibSequence[len(fibSequence)-1] + fibSequence[len(fibSequence)-2]
		if next > max {
			break
		}
		fibSequence = append(fibSequence, next)
	}
	return fibSequence
}

// createTestDocument は、指定された数のネストされたオブジェクトを持つドキュメントを作成します
func createTestDocument(count int) TestDocument {
	objects := make([]NestedObject, count)
	for i := 0; i < count; i++ {
		objects[i] = NestedObject{Data: fmt.Sprintf("data%d", i)}
	}
	return TestDocument{Objects: objects}
}

// retrieveDocument は、指定されたIDのドキュメントを取得します
func retrieveDocument(collection *mongo.Collection, id interface{}) error {
	var result TestDocument
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	return err
}
