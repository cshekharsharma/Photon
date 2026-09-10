package appconfig

import (
	"time"

	"github.com/cshekharsharma/photon/core/logger"
)

type FetchConfigInput struct {
	ApplicationName string
	EnvironmentName string
	ConfigProfile   string
	ConfigVersion   string
}

type WatchConfigInput struct {
	UniqueConfigID string
	Application    string
	Environment    string
	ConfigProfile  string
	ClientID       string
	PollInterval   time.Duration
	Logger         logger.Logger
}
