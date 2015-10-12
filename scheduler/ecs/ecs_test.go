package ecs

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/remind101/12factor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduler_Run(t *testing.T) {
	b := new(mockStackBuilder)
	s := &Scheduler{
		stackBuilder: b,
	}

	app := twelvefactor.App{}
	b.On("Build", app).Return(nil)
	err := s.Run(app)
	assert.NoError(t, err)
}

func TestScheduler_Remove(t *testing.T) {
	b := new(mockStackBuilder)
	s := &Scheduler{
		stackBuilder: b,
	}

	b.On("Remove", "app").Return(nil)
	err := s.Remove("app")
	assert.NoError(t, err)
}

func TestScheduler_ScaleProcess(t *testing.T) {
	b := new(mockStackBuilder)
	c := new(mockECSClient)
	s := &Scheduler{
		Cluster:      "cluster",
		stackBuilder: b,
		ecs:          c,
	}

	b.On("Services", "app").Return(map[string]string{
		"web": "app--web",
	}, nil)
	c.On("UpdateService", &ecs.UpdateServiceInput{
		Cluster:      aws.String("cluster"),
		DesiredCount: aws.Int64(1),
		Service:      aws.String("app--web"),
	}).Return(&ecs.UpdateServiceOutput{}, nil)
	err := s.ScaleProcess("app", "web", 1)
	assert.NoError(t, err)
}

func TestScheduler_ScaleProcess_NotFound(t *testing.T) {
	b := new(mockStackBuilder)
	c := new(mockECSClient)
	s := &Scheduler{
		Cluster:      "cluster",
		stackBuilder: b,
		ecs:          c,
	}

	b.On("Services", "app").Return(map[string]string{}, nil)
	err := s.ScaleProcess("app", "web", 1)
	assert.Error(t, err, "web process not found")
}

func TestScheduler_Tasks(t *testing.T) {
	b := new(mockStackBuilder)
	c := new(mockECSClient)
	s := &Scheduler{
		Cluster:      "cluster",
		stackBuilder: b,
		ecs:          c,
	}

	b.On("Services", "app").Return(map[string]string{
		"web": "app--web",
	}, nil)
	c.On("ListTasks", &ecs.ListTasksInput{
		Cluster:     aws.String("cluster"),
		ServiceName: aws.String("app--web"),
	}).Return(&ecs.ListTasksOutput{
		TaskArns: []*string{
			aws.String("arn:aws:ecs:us-east-1:012345678910:task/0b69d5c0-d655-4695-98cd-5d2d526d9d5a"),
		},
	}, nil)
	c.On("DescribeTasks", &ecs.DescribeTasksInput{
		Cluster: aws.String("cluster"),
		Tasks: []*string{
			aws.String("arn:aws:ecs:us-east-1:012345678910:task/0b69d5c0-d655-4695-98cd-5d2d526d9d5a"),
		},
	}).Return(&ecs.DescribeTasksOutput{
		Tasks: []*ecs.Task{
			{
				TaskArn:    aws.String("arn:aws:ecs:us-east-1:012345678910:task/0b69d5c0-d655-4695-98cd-5d2d526d9d5a"),
				LastStatus: aws.String("RUNNING"),
			},
		},
	}, nil)
	tasks, err := s.Tasks("app")
	assert.NoError(t, err)
	assert.Equal(t, tasks, []twelvefactor.Task{
		{
			ID:    "0b69d5c0-d655-4695-98cd-5d2d526d9d5a",
			State: "RUNNING",
		},
	})
}

// mockECSClient is an implementation of the ecsClient interface for testing.
type mockECSClient struct {
	mock.Mock
}

func (c *mockECSClient) UpdateService(input *ecs.UpdateServiceInput) (*ecs.UpdateServiceOutput, error) {
	args := c.Called(input)
	return args.Get(0).(*ecs.UpdateServiceOutput), args.Error(1)
}

func (c *mockECSClient) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	args := c.Called(input)
	return args.Get(0).(*ecs.ListTasksOutput), args.Error(1)
}

func (c *mockECSClient) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	args := c.Called(input)
	return args.Get(0).(*ecs.DescribeTasksOutput), args.Error(1)
}

// mockStackBuilder is an implementation of the StackBuilder interface for
// testing.
type mockStackBuilder struct {
	mock.Mock
}

func (b *mockStackBuilder) Build(app twelvefactor.App, processes ...twelvefactor.Process) error {
	args := b.Called(app)
	return args.Error(0)
}

func (b *mockStackBuilder) Remove(app string) error {
	args := b.Called(app)
	return args.Error(0)
}

func (b *mockStackBuilder) Services(app string) (map[string]string, error) {
	args := b.Called(app)
	return args.Get(0).(map[string]string), args.Error(1)
}
