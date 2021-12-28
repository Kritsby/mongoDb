package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Testing []struct {
	Symbol           string  `json:"symbol" bson:"symbol"`
	Price_24h        float64 `json:"price_24h" bson:"price_24h,inline"`
	Volume_24h       float64 `json:"volume_24h" bson:"volume_24h,inline"`
	Last_trade_price float64 `json:"last_trade_price" bson:"last_trade_price,inline"`
}

func main() {
	fmt.Println("Заупск приложения")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	r := mux.NewRouter()
	r.HandleFunc("/test", GetRec).Methods("GET")
	go workWithDb(client)
	log.Fatal(http.ListenAndServe(":12345", r))
}

func GetRec(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Context-Type", "application/json")
	load := getAll()
	json.NewEncoder(response).Encode(load)
}

func getAll() []primitive.D {
	collection := client.Database("test").Collection("Testing")
	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		log.Fatal(err)
	}

	var results []primitive.D
	for cur.Next(context.Background()) {
		var result bson.D
		e := cur.Decode(&result)
		if e != nil {
			log.Fatal(e)
		}
		results = append(results, result)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	cur.Close(context.Background())
	return results
}

func workWithDb(client *mongo.Client) {
	defer time.Sleep(30 * time.Second)
	collection := client.Database("test").Collection("Testing")

	resp, err := http.Get("https://api.blockchain.com/v3/exchange/tickers")
	if err != nil {
		fmt.Println("No response from request")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	var f Testing
	err = json.Unmarshal(body, &f)
	checkErr(err)

	for _, rec := range f {
		opts := options.Update().SetUpsert(true)
		filter := bson.D{{"symbol", rec.Symbol}}
		update := bson.D{{"$set", bson.D{{"price_24h", rec.Price_24h}, {"volume_24h", rec.Volume_24h}, {"last_trade_price", rec.Last_trade_price}}}}
		_, err := collection.UpdateOne(context.TODO(), filter, update, opts)
		checkErr(err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
