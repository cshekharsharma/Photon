# Cloud

Photon exposes cloud services through provider-neutral interfaces. Configure the vendor and credentials once during startup, then request the service interface you need.

```go
package examples

import (
	"context"
	"os"

	"github.com/cshekharsharma/photon/cloud"
	"github.com/cshekharsharma/photon/cloud/entity"
)

func CloudExample() error {
	ctx := context.Background()
	client, err := cloud.NewClient(cloud.Config{
		Vendor: cloud.CloudVendorAws,
		AuthArguments: &entity.CloudAuthArguments{
			AuthMode:  entity.AuthTypeAccessKey,
			AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			Region:    os.Getenv("AWS_REGION"),
		},
	})
	if err != nil {
		return err
	}

	objectStore, err := client.GetObjectStorage(ctx)
	if err != nil {
		return err
	}

	_, err = objectStore.GetObject(ctx, "my-bucket", "path/to/object.json")
	return err
}
```
