// Package notifier provides a unified interface for sending messages to different
// notification platforms such as Slack, Microsoft Teams, etc.
//
// It defines an extensible factory pattern to instantiate notifier clients
// based on type, allowing consumers to work with a consistent interface regardless
// of the underlying messaging system.
package notifier

import (
	"fmt"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/cshekharsharma/photon/notifier/slack"
	"github.com/cshekharsharma/photon/notifier/teams"
)

// Notifier defines an interface for sending notifications.
// Implementations of this interface are responsible for delivering
// messages to their intended recipients.
//
// The Send method takes an entity.Message as input and returns an error
// if the message could not be sent successfully.
type Notifier interface {

	// Send sends a notification message using the underlying provider's implementation.
	//
	// The `message` parameter must be an instance of a struct that implements the `entity.Message` interface.
	// For example:
	// - To send a message via Slack, provide a `*entity.SlackMessage` struct that conforms to `entity.Message`.
	// - To send a message via Microsoft Teams, provide a `*entity.TeamsMessage` struct that conforms to `entity.Message`.
	//
	// This design allows the `Notifier` interface to remain generic and extensible for different providers.
	//
	// Returns an error if the message could not be sent due to issues such as network failures,
	// invalid message formatting, or provider-specific errors.
	Send(message entity.Message) error
}

// NewNotifier creates a new Notifier instance based on the platform specified in the given request.
// It returns the appropriate Notifier implementation or an error if the request is nil or the platform is unsupported.
//
// Parameters:
//   - request: A pointer to an entity.Request object containing the platform and other necessary details.
//
// Returns:
//   - Notifier: An implementation of the Notifier interface corresponding to the specified platform.
//   - error: An error if the request is nil or the platform is not supported.
//
// Supported Platforms:
//   - entity.PlatformSlack: Creates a Slack notifier.
//   - entity.PlatformMSTeams: Creates a Microsoft Teams notifier.
//
// Errors:
//   - Returns an error if the request is nil.
//   - Returns an error if the platofrm specified in the request is unsupported.
func NewNotifier(request *entity.Request) (Notifier, error) {
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	switch request.Platform {

	case entity.PlatformSlack:
		return slack.NewSlackNotifier(request), nil
	case entity.PlatformMSTeams:
		return teams.NewTeamsNotifier(request), nil

	default:
		return nil, fmt.Errorf("unsupported notifier type: %s", request.Platform)
	}
}
