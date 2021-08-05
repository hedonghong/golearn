package learn1

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestMongo1(t *testing.T) {
	//1、建立链接
	goclient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://testdb:testdb@localhost:27017/testdb"))
	if err != nil {
		return
	}
	defer func() {
		goclient.Disconnect(context.TODO())
	}()
	//2、选择数据库
	collection := goclient.Database("testdb").Collection("testdb")
	//3、选择表collection
	res, err := collection.InsertOne(context.TODO(), bson.D{{"name", "pi"}})
	if err != nil {
		return
	}
	id := res.InsertedID
	fmt.Println(id)
}
