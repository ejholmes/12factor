package docker_test

import (
	"testing"

	"github.com/remind101/12factor"
	"github.com/remind101/12factor/schedulers/docker"
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
	s := newScheduler(t)

	if err := s.Run(app); err != nil {
		t.Fatal(err)
	}
}

func newScheduler(t testing.TB) *docker.Scheduler {
	s, err := docker.NewSchedulerFromEnv()
	if err != nil {
		t.Fatalf("Could not build docker scheduler: %v", err)
	}
	return s
}
