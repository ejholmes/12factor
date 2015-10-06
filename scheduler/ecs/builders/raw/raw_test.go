package raw

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStackBuilder_Services(t *testing.T) {
	c := new(mockECSClient)
	b := &StackBuilder{
		Cluster: "cluster",
		ecs:     c,
	}

	c.On("ListServicesPages", &ecs.ListServicesInput{
		Cluster: aws.String("cluster"),
	}).Return(nil, []*ecs.ListServicesOutput{
		{
			ServiceArns: []*string{
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app--web"),
			},
		},
	})
	services, err := b.Services("app")
	assert.NoError(t, err)
	assert.Equal(t, services, map[string]string{
		"web": "app--web",
	})
}

func TestStackBuilder_Services_Pagination(t *testing.T) {
	c := new(mockECSClient)
	b := &StackBuilder{
		Cluster: "cluster",
		ecs:     c,
	}

	c.On("ListServicesPages", &ecs.ListServicesInput{
		Cluster: aws.String("cluster"),
	}).Return(nil, []*ecs.ListServicesOutput{
		{
			ServiceArns: []*string{
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app--web"),
			},
		},
		{
			ServiceArns: []*string{
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app--worker"),
			},
		},
	})
	services, err := b.Services("app")
	assert.NoError(t, err)
	assert.Equal(t, services, map[string]string{
		"web":    "app--web",
		"worker": "app--worker",
	})
}

func TestStackBuilder_Services_Dirty(t *testing.T) {
	c := new(mockECSClient)
	b := &StackBuilder{
		Cluster: "cluster",
		ecs:     c,
	}

	c.On("ListServicesPages", &ecs.ListServicesInput{
		Cluster: aws.String("cluster"),
	}).Return(nil, []*ecs.ListServicesOutput{
		{
			ServiceArns: []*string{
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app"),
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app--web"),
				nil,
			},
		},
	})
	services, err := b.Services("app")
	assert.NoError(t, err)
	assert.Equal(t, services, map[string]string{
		"web": "app--web",
	})
}

func TestStackBuilder_Remove(t *testing.T) {
	c := new(mockECSClient)
	b := &StackBuilder{
		Cluster: "cluster",
		ecs:     c,
	}

	c.On("ListServicesPages", &ecs.ListServicesInput{
		Cluster: aws.String("cluster"),
	}).Return(nil, []*ecs.ListServicesOutput{
		{
			ServiceArns: []*string{
				aws.String("arn:aws:ecs:us-east-1:012345678910:service/app--web"),
			},
		},
	})
	c.On("DeleteService", &ecs.DeleteServiceInput{
		Cluster: aws.String("cluster"),
		Service: aws.String("app--web"),
	}).Return(&ecs.DeleteServiceOutput{}, nil)
	err := b.Remove("app")
	assert.NoError(t, err)
}

// mockECSClient is an implementation of the ecsClient interface for testing.
type mockECSClient struct {
	mock.Mock
}

func (c *mockECSClient) ListServicesPages(input *ecs.ListServicesInput, fn func(*ecs.ListServicesOutput, bool) bool) error {
	args := c.Called(input)
	for _, resp := range args.Get(1).([]*ecs.ListServicesOutput) {
		if !fn(resp, false) {
			break
		}
	}
	return args.Error(0)
}

func (c *mockECSClient) DeleteService(input *ecs.DeleteServiceInput) (*ecs.DeleteServiceOutput, error) {
	args := c.Called(input)
	return args.Get(0).(*ecs.DeleteServiceOutput), args.Error(1)
}
