# Mongo CRUD

Use `SetConnectionConfigE` during startup, then work through the returned Photon client interface.

```go
package examples

import (
	"context"
	"os"
	"time"

	"github.com/cshekharsharma/photon/storage/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Account struct {
	ID     string `bson:"_id"`
	Email  string `bson:"email"`
	Status string `bson:"status"`
}

func MongoCRUD(parent context.Context) error {
	err := mongo.SetConnectionConfigE("primary", &mongo.ConnectionConfig{
		Hosts:              []string{"localhost:27017"},
		Username:           "app",
		Password:           os.Getenv("MONGO_PASSWORD"),
		ConnectionTimeout:  5 * time.Second,
		ConnectionPoolsize: 25,
	})
	if err != nil {
		return err
	}

	client, err := mongo.Connect(&mongo.MongoDbConnector{}, "primary")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	accounts := client.Database("appdb").Collection("accounts")

	_, err = accounts.InsertOne(ctx, Account{
		ID:     "acct_123",
		Email:  "user@example.com",
		Status: "active",
	})
	if err != nil {
		return err
	}

	var account Account
	if err := accounts.FindOne(ctx, bson.M{"_id": "acct_123"}).Decode(&account); err != nil {
		return err
	}

	_, err = accounts.UpdateOne(
		ctx,
		bson.M{"_id": account.ID},
		bson.M{"$set": bson.M{"status": "disabled"}},
	)
	if err != nil {
		return err
	}

	_, err = accounts.DeleteOne(ctx, bson.M{"_id": account.ID})
	return err
}
```

