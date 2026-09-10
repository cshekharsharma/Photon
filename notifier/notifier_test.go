package notifier

import (
	"testing"

	"github.com/cshekharsharma/photon/notifier/entity"
	"github.com/cshekharsharma/photon/notifier/slack"
	"github.com/cshekharsharma/photon/notifier/teams"
	"github.com/stretchr/testify/assert"
)

func TestNewNotifier_NilRequest(t *testing.T) {
	notifier, err := NewNotifier(nil)
	assert.Nil(t, notifier)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")
}

func TestNewNotifier_UnsupportedTarget(t *testing.T) {
	req := &entity.Request{
		Platform: "unknown",
	}
	notifier, err := NewNotifier(req)
	assert.Nil(t, notifier)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notifier type")
}

func TestNewNotifier_Slack(t *testing.T) {
	req := &entity.Request{
		Platform: entity.PlatformSlack,
	}
	notifier, err := NewNotifier(req)
	assert.NoError(t, err)
	assert.NotNil(t, notifier)

	// Type assertion if you want to verify the implementation
	_, ok := notifier.(*slack.SlackNotifier)
	assert.True(t, ok)
}

func TestNewNotifier_MSTeams(t *testing.T) {
	req := &entity.Request{
		Platform: entity.PlatformMSTeams,
	}
	notifier, err := NewNotifier(req)
	assert.NoError(t, err)
	assert.NotNil(t, notifier)

	// Type assertion if you want to verify the implementation
	_, ok := notifier.(*teams.TeamsNotifier)
	assert.True(t, ok)
}
