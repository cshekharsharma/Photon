# Config

Load configuration through the Koanf-backed provider. Prefer `LoadE()` in service startup code so configuration failures can be returned cleanly.

```go
package examples

import (
	"fmt"

	"github.com/cshekharsharma/photon/core/config"
)

func ConfigExample() error {
	err := config.Init(config.ConfigProviderKoanf, &config.Options{
		Source:    config.SourceRawBytes,
		Format:    config.FormatJson,
		Content:   []byte(`{"server":{"port":8080}}`),
		Delimiter: config.DefaultConfigPathDelimiter,
	})
	if err != nil {
		return err
	}

	cfg, err := config.LoadE()
	if err != nil {
		return err
	}

	fmt.Println(cfg.GetInt64("server.port"))
	return nil
}
```
