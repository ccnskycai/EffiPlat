package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceInstanceStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		s        ServiceInstanceStatusType
		expected bool
	}{
		{name: "Valid status running", s: ServiceInstanceStatusRunning, expected: true},
		{name: "Valid status stopped", s: ServiceInstanceStatusStopped, expected: true},
		{name: "Valid status deploying", s: ServiceInstanceStatusDeploying, expected: true},
		{name: "Valid status error", s: ServiceInstanceStatusError, expected: true},
		{name: "Valid status unknown", s: ServiceInstanceStatusUnknown, expected: true},
		{name: "Invalid status empty", s: ServiceInstanceStatusType(""), expected: false},
		{name: "Invalid status custom", s: ServiceInstanceStatusType("custom_status"), expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.s.IsValid())
		})
	}
}

func TestServiceInstance_TableName(t *testing.T) {
	var si ServiceInstance
	assert.Equal(t, "service_instances", si.TableName())
}
