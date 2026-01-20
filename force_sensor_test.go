package kettlecycletest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)


// errorForceReader always returns an error
type errorForceReader struct{}

func (e *errorForceReader) ReadForce(ctx context.Context) (float64, error) {
	return 0, errors.New("simulated reader error")
}

func TestForceSensorConfig(t *testing.T) {
	t.Run("empty config is valid", func(t *testing.T) {
		cfg := &ForceSensorConfig{}
		deps, _, err := cfg.Validate("test")
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if len(deps) != 0 {
			t.Fatalf("expected 0 dependencies, got %d", len(deps))
		}
	})

	t.Run("config with load_cell returns dependency", func(t *testing.T) {
		cfg := &ForceSensorConfig{
			LoadCell: "my-adc-sensor",
		}
		deps, _, err := cfg.Validate("test")
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if len(deps) != 1 {
			t.Fatalf("expected 1 dependency, got %d: %v", len(deps), deps)
		}

		if deps[0] != "my-adc-sensor" {
			t.Errorf("expected load_cell dependency, got %q", deps[0])
		}
	})
}

func TestForceSensor_Readings(t *testing.T) {
	t.Run("returns empty trial metadata when no capture started", func(t *testing.T) {
		logger := logging.NewTestLogger(t)

		fs := &forceSensor{
			name:         resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:       logger,
			reader:       newMockForceReader(),
			sampleRateHz: 50,
			bufferSize:   100,
			samples:      make([]float64, 0),
		}

		readings, err := fs.Readings(context.Background(), nil)
		if err != nil {
			t.Fatalf("Readings failed: %v", err)
		}

		if readings["trial_id"] != "" {
			t.Errorf("expected empty trial_id, got %v", readings["trial_id"])
		}
		if readings["cycle_count"] != 0 {
			t.Errorf("expected cycle_count=0, got %v", readings["cycle_count"])
		}
		if readings["should_sync"] != false {
			t.Errorf("expected should_sync=false, got %v", readings["should_sync"])
		}
		samples := readings["samples"].([]interface{})
		if len(samples) != 0 {
			t.Errorf("expected empty samples, got %v", samples)
		}
	})

	t.Run("should_sync lifecycle: true during capture, false after end_capture", func(t *testing.T) {
		logger := logging.NewTestLogger(t)

		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   50,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0),
			state:          captureIdle,
		}

		// Before capture: should_sync is false
		readings, _ := fs.Readings(context.Background(), nil)
		if readings["should_sync"] != false {
			t.Errorf("expected should_sync=false before capture, got %v", readings["should_sync"])
		}

		// Start capture with trial metadata
		fs.DoCommand(context.Background(), map[string]interface{}{
			"command":     "start_capture",
			"trial_id":    "trial-123",
			"cycle_count": 5,
		})

		// During capture: should_sync is true
		readings, err := fs.Readings(context.Background(), nil)
		if err != nil {
			t.Fatalf("Readings failed: %v", err)
		}

		if readings["trial_id"] != "trial-123" {
			t.Errorf("expected trial_id=trial-123, got %v", readings["trial_id"])
		}
		if readings["cycle_count"] != 5 {
			t.Errorf("expected cycle_count=5, got %v", readings["cycle_count"])
		}
		if readings["should_sync"] != true {
			t.Errorf("expected should_sync=true during capture, got %v", readings["should_sync"])
		}

		// End capture
		result, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
		if err != nil {
			t.Fatalf("end_capture failed: %v", err)
		}

		// end_capture should return the trial metadata
		if result["trial_id"] != "trial-123" {
			t.Errorf("expected trial_id=trial-123 in end_capture result, got %v", result["trial_id"])
		}

		// After capture: should_sync is false
		readings, _ = fs.Readings(context.Background(), nil)
		if readings["should_sync"] != false {
			t.Errorf("expected should_sync=false after end_capture, got %v", readings["should_sync"])
		}
		if readings["trial_id"] != "" {
			t.Errorf("expected trial_id='' after end_capture, got %v", readings["trial_id"])
		}
	})

	t.Run("includes max_force when samples present", func(t *testing.T) {
		logger := logging.NewTestLogger(t)

		fs := &forceSensor{
			name:         resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:       logger,
			reader:       newMockForceReader(),
			sampleRateHz: 50,
			bufferSize:   100,
			samples:      []float64{10.0, 50.0, 30.0, 25.0},
		}

		readings, err := fs.Readings(context.Background(), nil)
		if err != nil {
			t.Fatalf("Readings failed: %v", err)
		}

		maxForce, ok := readings["max_force"].(float64)
		if !ok {
			t.Fatal("max_force not present or not float64")
		}
		if maxForce != 50.0 {
			t.Errorf("expected max_force=50.0, got %v", maxForce)
		}
	})
}

func TestForceSensor_Capture(t *testing.T) {
	t.Run("start_capture begins waiting state", func(t *testing.T) {
		logger := logging.NewTestLogger(t)

		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		result, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})
		if err != nil {
			t.Fatalf("start_capture failed: %v", err)
		}

		if result["status"] != "waiting" {
			t.Errorf("expected status=waiting, got %v", result["status"])
		}

		readings, _ := fs.Readings(context.Background(), nil)
		if readings["capture_state"] != "waiting" {
			t.Errorf("expected capture_state=waiting, got %v", readings["capture_state"])
		}

		// Cleanup
		fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
	})

	t.Run("captures samples after non-zero reading", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		// Start capture - mock reader simulates contact when start_capture is called
		_, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})
		if err != nil {
			t.Fatalf("start_capture failed: %v", err)
		}

		// Wait for samples to accumulate
		time.Sleep(150 * time.Millisecond)

		result, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
		if err != nil {
			t.Fatalf("end_capture failed: %v", err)
		}

		sampleCount := result["sample_count"].(int)
		if sampleCount == 0 {
			t.Error("expected samples to be captured")
		}

		maxForce := result["max_force"].(float64)
		if maxForce < 50.0 {
			t.Errorf("expected max_force >= 50, got %v", maxForce)
		}
	})

	t.Run("end_capture without start returns error", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		_, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
		if err == nil {
			t.Error("expected error when ending capture that wasn't started")
		}
	})

	t.Run("double start_capture returns error", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		_, err := fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})
		if err != nil {
			t.Fatalf("first start_capture failed: %v", err)
		}

		_, err = fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})
		if err == nil {
			t.Error("expected error on double start_capture")
		}

		// Cleanup
		fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
	})

	t.Run("buffer respects max size", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		bufferSize := 10
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   500, // Fast sampling to fill buffer quickly
			bufferSize:     bufferSize,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, bufferSize),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})

		// Wait long enough to exceed buffer size
		time.Sleep(100 * time.Millisecond)

		readings, _ := fs.Readings(context.Background(), nil)
		samples := readings["samples"].([]interface{})

		if len(samples) > bufferSize {
			t.Errorf("buffer exceeded max size: got %d, max %d", len(samples), bufferSize)
		}

		fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
	})

	t.Run("continues on reader error", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         &errorForceReader{},
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})

		// Should not panic, just log errors
		time.Sleep(50 * time.Millisecond)

		// Sensor should still be responsive
		readings, err := fs.Readings(context.Background(), nil)
		if err != nil {
			t.Fatalf("Readings failed after reader errors: %v", err)
		}

		// Samples should be empty since reader always errors
		samples := readings["samples"].([]interface{})
		if len(samples) != 0 {
			t.Errorf("expected no samples with erroring reader, got %d", len(samples))
		}

		fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
	})
}

func TestForceSensor_ThreadSafety(t *testing.T) {
	t.Run("concurrent readings during capture", func(t *testing.T) {
		logger := logging.NewTestLogger(t)
		
		fs := &forceSensor{
			name:           resource.NewName(resource.APINamespaceRDK.WithComponentType("sensor"), "test"),
			logger:         logger,
			reader:         newMockForceReader(),
			sampleRateHz:   100,
			bufferSize:     100,
			zeroThreshold:  5.0,
			captureTimeout: 10 * time.Second,
			samples:        make([]float64, 0, 100),
			state:          captureIdle,
		}

		go fs.samplingLoop()

		// Start capture
		fs.DoCommand(context.Background(), map[string]interface{}{"command": "start_capture"})

		// Concurrent readings
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					_, err := fs.Readings(context.Background(), nil)
					if err != nil {
						t.Errorf("concurrent Readings failed: %v", err)
					}
					time.Sleep(5 * time.Millisecond)
				}
			}()
		}

		wg.Wait()

		// Cleanup
		fs.DoCommand(context.Background(), map[string]interface{}{"command": "end_capture"})
	})
}

func TestMockForceReader(t *testing.T) {
	t.Run("returns near-zero when not in contact", func(t *testing.T) {
		reader := newMockForceReader()

		v, err := reader.ReadForce(context.Background())
		if err != nil {
			t.Fatalf("ReadForce failed: %v", err)
		}

		if v >= 5.0 {
			t.Errorf("expected near-zero when not in contact, got %v", v)
		}
	})

	t.Run("returns increasing values when in contact", func(t *testing.T) {
		reader := newMockForceReader()
		reader.SetContact(true)

		var values []float64
		for i := 0; i < 10; i++ {
			v, err := reader.ReadForce(context.Background())
			if err != nil {
				t.Fatalf("ReadForce failed: %v", err)
			}
			values = append(values, v)
		}

		// Values should be increasing (mock simulates ramp)
		for i := 1; i < len(values); i++ {
			if values[i] <= values[i-1] {
				t.Errorf("expected increasing values, got %v at index %d after %v", values[i], i, values[i-1])
			}
		}
	})

	t.Run("resets ramp when contact toggled", func(t *testing.T) {
		reader := newMockForceReader()
		reader.SetContact(true)

		// Read a few times to advance the ramp
		for i := 0; i < 5; i++ {
			reader.ReadForce(context.Background())
		}

		// Toggle contact off then on
		reader.SetContact(false)
		reader.SetContact(true)

		// First reading after toggle should be back at start of ramp
		v, _ := reader.ReadForce(context.Background())
		if v > 60.0 {
			t.Errorf("expected ramp to reset, got %v", v)
		}
	})
}
