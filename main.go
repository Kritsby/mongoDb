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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Testing []struct {
	Symbol           string  `json:symbol bson:symbol`
	Price_24h        float64 `json:price_24h bson:price_24h`
	Volume_24h       float64 `json:volume_24h bson:volume_24h`
	Last_trade_price float64 `json:last_trade_price bson:last_trade_price`
}

var client *mongo.Client

func main() {
	Connect()
}

func Connect() {
	fmt.Println("Заупск приложения")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	checkErr(err)
	err = client.Connect(context.TODO())
	checkErr(err)
	r := mux.NewRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}

func GetRec(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var recov []Testing
	collection := client.Database("test").Collection("Testing")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, _ := collection.Find(ctx, bson.M{})
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var testing Testing
		cursor.Decode(&testing)
		recov = append(recov, testing)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(recov)
}

func workWithDb() {

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
		_, err := collection.InsertOne(context.TODO(), rec)
		checkErr(err)
		time.Sleep(30 * time.Second)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
