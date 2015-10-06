package ecs_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/remind101/12factor"
	"github.com/remind101/12factor/scheduler/ecs"
)

// app is our test application. This is a valid application that will be run
// with the docker daemon.
var app = twelvefactor.App{
	ID:      "acme",
	Name:    "acme",
	Image:   "remind101/acme-inc",
	Version: "v1",
	Env: map[string]string{
		"RAILS_ENV": "production",
	},
	Processes: []twelvefactor.Process{
		{
			Name:    "web",
			Command: []string{"acme-inc", "web"},
		},
	},
}

func TestScheduler_Run(t *testing.T) {
	//s := newScheduler(t)

	//if err := s.Run(app); err != nil {
	//t.Fatal(err)
	//}
}

func newScheduler(t testing.TB) *ecs.Scheduler {
	creds := &credentials.EnvProvider{}
	if _, err := creds.Retrieve(); err != nil {
		t.Skip("Skipping ECS test because AWS_ environment variables are not present.")
	}

	config := defaults.DefaultConfig.WithCredentials(credentials.NewCredentials(creds))
	return ecs.NewScheduler(config)
}
