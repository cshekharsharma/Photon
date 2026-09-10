# Storage

Register each datastore once during startup, then reuse the named connection throughout the process.

## Mongo

```go
package examples

import (
	"context"
	"os"
	"time"

	"github.com/cshekharsharma/photon/storage/mongo"
)

func MongoExample(ctx context.Context) error {
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

	return client.Ping(ctx)
}
```

## Postgres

```go
package examples

import (
	"context"
	"os"
	"time"

	"github.com/cshekharsharma/photon/storage/postgres"
)

func PostgresExample(ctx context.Context) error {
	err := postgres.SetConnectionConfigE("primary", &postgres.ConnectionConfig{
		Host:                "localhost",
		Port:                "5432",
		UserName:            "app",
		Password:            os.Getenv("POSTGRES_PASSWORD"),
		DbName:              "appdb",
		SSLMode:             "require",
		MaxOpenConn:         25,
		MaxIdleConn:         10,
		ConnectTimeout:      5 * time.Second,
		PingTimeout:         2 * time.Second,
		DefaultQueryTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}

	db, err := postgres.Connect(&postgres.PostgresDbConnector{}, "primary")
	if err != nil {
		return err
	}

	return db.PingContext(ctx)
}
```
