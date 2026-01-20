package kettlecycletest

import (
	"context"
	"testing"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func TestNewController(t *testing.T) {
	logger := logging.NewTestLogger(t)
	name := resource.NewName(resource.APINamespaceRDK.WithServiceType("generic"), "test")

	ctrl, err := NewController(context.Background(), nil, name, &Config{}, logger)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}
	if ctrl == nil {
		t.Fatal("NewController returned nil")
	}
	if ctrl.Name() != name {
		t.Errorf("Name() = %v, want %v", ctrl.Name(), name)
	}
}

func TestDoCommand(t *testing.T) {
	logger := logging.NewTestLogger(t)
	name := resource.NewName(resource.APINamespaceRDK.WithServiceType("generic"), "test")

	ctrl, err := NewController(context.Background(), nil, name, &Config{}, logger)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	_, err = ctrl.(*kettleCycleTestController).DoCommand(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Error("DoCommand should return error for unimplemented stub")
	}
}

func TestClose(t *testing.T) {
	logger := logging.NewTestLogger(t)
	name := resource.NewName(resource.APINamespaceRDK.WithServiceType("generic"), "test")

	ctrl, err := NewController(context.Background(), nil, name, &Config{}, logger)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	err = ctrl.(*kettleCycleTestController).Close(context.Background())
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
