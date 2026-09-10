# Postgres Queries

Use context deadlines around database work. Register the connection once at startup, then reuse the named connection.

```go
package examples

import (
	"context"
	"os"
	"time"

	"github.com/cshekharsharma/photon/storage/postgres"
)

type User struct {
	ID    int64
	Email string
}

func LoadUser(parent context.Context, email string) (*User, error) {
	if err := postgres.SetConnectionConfigE("primary", &postgres.ConnectionConfig{
		Host:                "localhost",
		Port:                "5432",
		UserName:            "app",
		Password:            os.Getenv("POSTGRES_PASSWORD"),
		DbName:              "appdb",
		SSLMode:             "require",
		MaxOpenConn:         25,
		MaxIdleConn:         10,
		ConnMaxLifetime:     30 * time.Minute,
		ConnMaxIdleTime:     5 * time.Minute,
		ConnectTimeout:      5 * time.Second,
		PingTimeout:         2 * time.Second,
		DefaultQueryTimeout: 5 * time.Second,
		PgBouncer:           postgres.PgBouncerTransaction,
		RuntimeParams: map[string]string{
			"application_name": "checkout-api",
		},
	}); err != nil {
		return nil, err
	}

	db, err := postgres.Connect(&postgres.PostgresDbConnector{}, "primary")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	user := &User{}
	err = db.QueryRowContext(ctx, `
		SELECT id, email
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}
```
