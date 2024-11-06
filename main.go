package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NestedObject struct {
	Order int                  `bson:"order"`
	Data  string               `bson:"data"`
	Bool  bool                 `bson:"bool"`
	IDs   []primitive.ObjectID `bson:"ids"`
}

type TestDocument struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Objects       []NestedObject     `bson:"objects"`
	SizeInBytes   int                `bson:"sizeInBytes"`
	InsertionTime string             `bson:"insertionTime"` // 挿入時間
	RetrievalTime int64              `bson:"retrievalTime"` // 取得時間 (ミリ秒)
}

func main() {
	// .envファイルをロード
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// 環境変数からMongoDBのURIを取得
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatalf("MONGODB_URI must be set")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collectionName := generateUniqueCollectionName()
	collection := client.Database("testdb").Collection(collectionName)
	objectCounts := generateFibonacciUpTo(10000)

	// ヘッダーを表示
	fmt.Println("ObjectCount,SizeInBytes,InsertionTime,RetrievalTime(ms)")

	for _, count := range objectCounts {
		start := time.Now()
		doc := createTestDocument(count)

		// ドキュメントサイズの計測
		docSize, err := calculateDocumentSize(doc)
		if err != nil {
			log.Fatalf("Failed to calculate document size: %v", err)
		}
		doc.SizeInBytes = docSize
		doc.InsertionTime = start.Format("2006-01-02 15:04:05")

		// ドキュメント挿入
		insertedResult, err := collection.InsertOne(context.TODO(), doc)
		if err != nil {
			log.Fatalf("Failed to insert document: %v", err)
		}

		// ドキュメント取得時間の計測
		retrievalStart := time.Now()
		retrievedDoc, err := retrieveDocument(collection, insertedResult.InsertedID)
		if err != nil {
			log.Fatalf("Failed to retrieve document: %v", err)
		}
		retrievedDoc.RetrievalTime = time.Since(retrievalStart).Milliseconds() // 取得時間（ミリ秒）を設定

		// 更新して取得時間を保存
		_, err = collection.UpdateOne(
			context.TODO(),
			bson.M{"_id": insertedResult.InsertedID},
			bson.M{"$set": bson.M{"retrievalTime": retrievedDoc.RetrievalTime}},
		)
		if err != nil {
			log.Fatalf("Failed to update document with retrieval time: %v", err)
		}

		// CSV形式でコンソールに出力
		fmt.Printf("%d,%d,%s,%d\n", count, doc.SizeInBytes, doc.InsertionTime, retrievedDoc.RetrievalTime)
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
func generateRandomObjectIDs(count int) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, count)
	for i := range count {
		ids[i] = primitive.NewObjectID()
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

// retrieveDocument は、指定されたIDのドキュメントを取得します
func retrieveDocument(collection *mongo.Collection, id interface{}) (TestDocument, error) {
	var result TestDocument
	err := collection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&result)
	return result, err
}
