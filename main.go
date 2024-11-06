package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NestedObject struct {
	Data string `bson:"data"`
}

type TestDocument struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Objects     []NestedObject     `bson:"objects"`
	SizeInBytes int                `bson:"sizeInBytes"`
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collectionName := generateUniqueCollectionName()
	collection := client.Database("testdb").Collection(collectionName)
	objectCounts := generateFibonacciUpTo(100)

	for _, count := range objectCounts {
		doc := createTestDocument(count)
		docSize, err := calculateDocumentSize(doc)
		if err != nil {
			log.Fatalf("Failed to calculate document size: %v", err)
		}
		doc.SizeInBytes = docSize

		_, err = collection.InsertOne(context.TODO(), doc)
		if err != nil {
			log.Fatalf("Failed to insert document: %v", err)
		}

		fmt.Printf("Inserted document with %d objects, Size: %d bytes\n", count, doc.SizeInBytes)
	}
}

// generateUniqueCollectionName は、日付と時刻を基にユニークなコレクション名を生成します
func generateUniqueCollectionName() string {
	return fmt.Sprintf("testcollection_%s", time.Now().Format("20060102_150405"))
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

// calculateDocumentSize はドキュメントをバイナリ形式に変換し、バイトサイズを計測します
func calculateDocumentSize(doc TestDocument) (int, error) {
	bsonData, err := bson.Marshal(doc)
	if err != nil {
		return 0, err
	}
	return len(bsonData), nil
}
