package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NestedObject struct {
	Order int             `bson:"order"`
	Data  string          `bson:"data"`
	Bool  bool            `bson:"bool"`
	IDs   []bson.ObjectID `bson:"ids"`
}

type TestDocument struct {
	ID            bson.ObjectID  `bson:"_id,omitempty"`
	Objects       []NestedObject `bson:"objects"`
	SizeInBytes   int            `bson:"sizeInBytes"`
	InsertionTime string         `bson:"insertionTime"`
	RetrievalTime int64          `bson:"retrievalTime"`
}

// go run main.go で実行
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatalf("MONGODB_URI must be set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	// CSV形式のヘッダーを出力
	fmt.Println("ObjectCount,AvgSizeInBytes,RetrievalTime(ms)")

	objectCounts := generateFibonacciSequenceUpTo(10000)

	for _, n := range objectCounts {
		collectionName := fmt.Sprintf("testcollection_N=%d_%s", n, time.Now().Format("20060102_150405"))
		collection := client.Database("testdb").Collection(collectionName)

		var totalSize int64
		docCount := 35 // 35個のドキュメントを生成

		for range docCount {
			start := time.Now()
			doc := createTestDocument(n)

			docSize, err := calculateDocumentSize(doc)
			if err != nil {
				log.Fatalf("Failed to calculate document size: %v", err)
			}
			totalSize += int64(docSize)
			doc.SizeInBytes = docSize
			doc.InsertionTime = start.Format("2006-01-02 15:04:05")

			if _, err = collection.InsertOne(context.TODO(), doc); err != nil {
				log.Fatalf("Failed to insert document: %v", err)
			}
		}

		averageSize := totalSize / int64(docCount)
		retrievalStart := time.Now() // 検索時間を計測開始
		cursor, err := collection.Find(context.TODO(), bson.M{})
		if err != nil {
			log.Fatalf("Failed to retrieve documents: %v", err)
		}
		var retrievedDocs []TestDocument
		if err = cursor.All(context.TODO(), &retrievedDocs); err != nil {
			log.Fatalf("Failed to decode documents: %v", err)
		}
		retrievalTime := time.Since(retrievalStart).Milliseconds() // 検索時間を計測終了

		// CSV形式で結果を出力
		fmt.Printf("%d,%d,%d\n", n, averageSize, retrievalTime)
	}
}

// フィボナッチ数列を指定された最大値まで生成
func generateFibonacciSequenceUpTo(max int) []int {
	fibSequence := []int{1, 2}
	for {
		next := fibSequence[len(fibSequence)-1] + fibSequence[len(fibSequence)-2]
		if next > max {
			break
		}
		fibSequence = append(fibSequence, next)
	}
	return fibSequence
}

// createTestDocument は、指定された数の NestedObject を持つドキュメントを作成します
func createTestDocument(count int) TestDocument {
	objects := make([]NestedObject, count)
	for i := range count {
		objects[i] = NestedObject{
			Order: i,
			Data:  generateRandomString(12),
			Bool:  rand.IntN(2) == 0,
			IDs:   generateRandomObjectIDs(5 + rand.IntN(16)),
		}
	}
	return TestDocument{Objects: objects}
}

// generateRandomString は指定された長さのランダムな文字列を生成します
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}

// generateRandomObjectIDs は指定された数のランダムなObjectIDを生成します
func generateRandomObjectIDs(count int) []bson.ObjectID {
	ids := make([]bson.ObjectID, count)
	for i := range count {
		ids[i] = bson.NewObjectID()
	}
	return ids
}

// calculateDocumentSize はドキュメントをバイナリ形式に変換し、バイトサイズを計測します
func calculateDocumentSize(doc TestDocument) (int, error) {
	bsonData, err := bson.Marshal(doc)
	if err != nil {
		return 0, err
	}
	return len(bsonData), nil
}
