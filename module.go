package kettlecycletest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"os"
	"sync"
	"time"

	"go.viam.com/rdk/app"
	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/sensor"
	toggleswitch "go.viam.com/rdk/components/switch"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	generic "go.viam.com/rdk/services/generic"
)

var Controller = resource.NewModel("viamdemo", "kettle-cycle-test", "controller")

func init() {
	resource.RegisterService(generic.API, Controller,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newKettleCycleTestController,
		},
	)
}

type Config struct {
	Arm              string `json:"arm"`
	RestingPosition  string `json:"resting_position"`
	PourPrepPosition string `json:"pour_prep_position"`
	ForceSensor      string `json:"force_sensor,omitempty"`

	// Camera capture settings (Camera, DatasetID, PartID required if Camera is set)
	// API credentials read from VIAM_API_KEY and VIAM_API_KEY_ID environment variables
	Camera    string `json:"camera,omitempty"`
	DatasetID string `json:"dataset_id,omitempty"`
	PartID    string `json:"part_id,omitempty"`
}

type trialState struct {
	trialID     string
	cycleCount  int
	startedAt   time.Time
	lastCycleAt time.Time
	stopCh      chan struct{}
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.Arm == "" {
		return nil, nil, fmt.Errorf("%s: arm is required", path)
	}
	if cfg.RestingPosition == "" {
		return nil, nil, fmt.Errorf("%s: resting_position is required", path)
	}
	if cfg.PourPrepPosition == "" {
		return nil, nil, fmt.Errorf("%s: pour_prep_position is required", path)
	}

	// If camera is configured, dataset_id and part_id are required
	// API credentials come from environment variables
	if cfg.Camera != "" {
		if cfg.DatasetID == "" || cfg.PartID == "" {
			return nil, nil, fmt.Errorf("%s: camera requires dataset_id and part_id", path)
		}
	}

	deps := []string{cfg.Arm, cfg.RestingPosition, cfg.PourPrepPosition}
	if cfg.ForceSensor != "" {
		deps = append(deps, cfg.ForceSensor)
	}
	if cfg.Camera != "" {
		deps = append(deps, cfg.Camera)
	}
	return deps, nil, nil
}

type kettleCycleTestController struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	cfg    *Config

	arm         arm.Arm
	resting     toggleswitch.Switch
	pourPrep    toggleswitch.Switch
	forceSensor sensor.Sensor // optional, may be nil

	// Camera capture (optional)
	camera     camera.Camera
	viamClient *app.ViamClient
	dataClient *app.DataClient
	datasetID  string
	partID     string

	cancelCtx  context.Context
	cancelFunc func()

	mu          sync.Mutex
	activeTrial *trialState
}

func newKettleCycleTestController(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (resource.Resource, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewController(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewController(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (resource.Resource, error) {
	a, err := arm.FromDependencies(deps, conf.Arm)
	if err != nil {
		return nil, fmt.Errorf("getting arm: %w", err)
	}

	resting, err := toggleswitch.FromDependencies(deps, conf.RestingPosition)
	if err != nil {
		return nil, fmt.Errorf("getting resting position switch: %w", err)
	}

	pourPrep, err := toggleswitch.FromDependencies(deps, conf.PourPrepPosition)
	if err != nil {
		return nil, fmt.Errorf("getting pour_prep position switch: %w", err)
	}

	var fs sensor.Sensor
	if conf.ForceSensor != "" {
		fs, err = sensor.FromDependencies(deps, conf.ForceSensor)
		if err != nil {
			return nil, fmt.Errorf("getting force sensor: %w", err)
		}
		logger.Infof("controller using force sensor: %s", conf.ForceSensor)
	}

	// Camera and DataClient initialization (optional)
	var cam camera.Camera
	var viamClient *app.ViamClient
	var dataClient *app.DataClient
	if conf.Camera != "" {
		cam, err = camera.FromDependencies(deps, conf.Camera)
		if err != nil {
			return nil, fmt.Errorf("getting camera: %w", err)
		}

		// Read API credentials from file (env vars don't work for hot-reloaded modules)
		apiKey, apiKeyID, err := readDataAPICredentials()
		if err != nil {
			return nil, fmt.Errorf("camera configured but failed to read API credentials: %w", err)
		}

		viamClient, err = app.CreateViamClientWithAPIKey(ctx, app.Options{}, apiKey, apiKeyID, logger)
		if err != nil {
			return nil, fmt.Errorf("creating viam client: %w", err)
		}
		dataClient = viamClient.DataClient()
		logger.Infof("controller using camera %s with dataset %s", conf.Camera, conf.DatasetID)
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &kettleCycleTestController{
		name:        name,
		logger:      logger,
		cfg:         conf,
		arm:         a,
		resting:     resting,
		pourPrep:    pourPrep,
		forceSensor: fs,
		camera:      cam,
		viamClient:  viamClient,
		dataClient:  dataClient,
		datasetID:   conf.DatasetID,
		partID:      conf.PartID,
		cancelCtx:   cancelCtx,
		cancelFunc:  cancelFunc,
	}
	return s, nil
}

func (s *kettleCycleTestController) Name() resource.Name {
	return s.name
}

func (s *kettleCycleTestController) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	command, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'command' field")
	}

	switch command {
	case "execute_cycle":
		return s.handleExecuteCycle(ctx)
	case "start":
		return s.handleStart()
	case "stop":
		return s.handleStop()
	case "status":
		return s.handleStatus()
	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

func (s *kettleCycleTestController) handleExecuteCycle(ctx context.Context) (map[string]interface{}, error) {
	// Increment cycle count at start so all captured data uses correct cycle number
	s.mu.Lock()
	if s.activeTrial != nil {
		s.activeTrial.cycleCount++
	}
	s.mu.Unlock()

	if err := s.pourPrep.SetPosition(ctx, 2, nil); err != nil {
		return nil, fmt.Errorf("moving to pour_prep position: %w", err)
	}

	// Wait for arm to reach pour-prep position
	if err := s.waitForArmStopped(ctx); err != nil {
		s.logger.Warnf("error waiting for arm to stop at pour-prep: %v", err)
	}

	// Capture and upload image if camera is configured
	if s.camera != nil && s.dataClient != nil {
		if err := s.captureAndUploadImage(ctx); err != nil {
			return nil, fmt.Errorf("capturing image: %w", err)
		}
	}

	// Start force capture if sensor is configured
	if s.forceSensor != nil {
		s.mu.Lock()
		captureCmd := map[string]interface{}{"command": "start_capture"}
		if s.activeTrial != nil {
			captureCmd["trial_id"] = s.activeTrial.trialID
			captureCmd["cycle_count"] = s.activeTrial.cycleCount
		}
		s.mu.Unlock()

		_, err := s.forceSensor.DoCommand(ctx, captureCmd)
		if err != nil {
			s.logger.Warnf("failed to start force capture: %v", err)
		}
	}

	if err := s.resting.SetPosition(ctx, 2, nil); err != nil {
		// Try to end capture on error
		if s.forceSensor != nil {
			s.forceSensor.DoCommand(ctx, map[string]interface{}{"command": "end_capture"})
		}
		return nil, fmt.Errorf("returning to resting position: %w", err)
	}

	// Wait for arm to stop moving
	if err := s.waitForArmStopped(ctx); err != nil {
		s.logger.Warnf("error waiting for arm to stop: %v", err)
	}

	// End force capture
	var captureResult map[string]interface{}
	if s.forceSensor != nil {
		var err error
		captureResult, err = s.forceSensor.DoCommand(ctx, map[string]interface{}{"command": "end_capture"})
		if err != nil {
			s.logger.Warnf("failed to end force capture: %v", err)
		} else {
			s.logger.Infof("force capture: %v", captureResult)
		}
	}

	s.mu.Lock()
	if s.activeTrial != nil {
		s.activeTrial.lastCycleAt = time.Now()
	}
	s.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1 * time.Second):
	}

	result := map[string]interface{}{"status": "completed"}
	if captureResult != nil {
		result["force_capture"] = captureResult
	}
	return result, nil
}

func (s *kettleCycleTestController) waitForArmStopped(ctx context.Context) error {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for arm to stop")
		case <-ticker.C:
			moving, err := s.arm.IsMoving(ctx)
			if err != nil {
				return fmt.Errorf("checking arm movement: %w", err)
			}
			if !moving {
				return nil
			}
		}
	}
}

func (s *kettleCycleTestController) captureAndUploadImage(ctx context.Context) error {
	// Get raw image bytes from camera
	s.logger.Info("capturing image from camera")
	imageBytes, _, err := s.camera.Image(ctx, "image/jpeg", nil)
	if err != nil {
		return fmt.Errorf("getting image from camera: %w", err)
	}
	if len(imageBytes) == 0 {
		return fmt.Errorf("camera returned empty image")
	}
	s.logger.Infof("got %d bytes from camera", len(imageBytes))

	// Decode using image.Decode which auto-detects format
	img, format, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("decoding image: %w", err)
	}
	if img == nil {
		return fmt.Errorf("decoded image is nil")
	}
	s.logger.Infof("decoded image: format=%s, bounds=%v", format, img.Bounds())

	// Build tags from current trial state
	s.mu.Lock()
	var trialID string
	var cycleCount int
	if s.activeTrial != nil {
		trialID = s.activeTrial.trialID
		cycleCount = s.activeTrial.cycleCount
	}
	s.mu.Unlock()

	tags := formatCaptureTags(trialID, cycleCount)
	s.logger.Infof("uploading to dataset %s with tags %v", s.datasetID, tags)

	_, err = s.dataClient.UploadImageToDatasets(
		ctx,
		s.partID,
		img,
		[]string{s.datasetID},
		tags,
		app.MimeTypeJPEG,
		&app.FileUploadOptions{},
	)
	if err != nil {
		return fmt.Errorf("uploading image: %w", err)
	}

	s.logger.Infof("uploaded image with tags: %v", tags)
	return nil
}

func (s *kettleCycleTestController) handleStart() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeTrial != nil {
		return nil, fmt.Errorf("trial already running: %s", s.activeTrial.trialID)
	}

	now := time.Now()
	trialID := fmt.Sprintf("trial-%s", now.Format("20060102-150405"))
	stopCh := make(chan struct{})

	s.activeTrial = &trialState{
		trialID:   trialID,
		startedAt: now,
		stopCh:    stopCh,
	}

	// Start background cycling loop
	go s.cycleLoop(stopCh)

	return map[string]interface{}{
		"trial_id": trialID,
	}, nil
}

func (s *kettleCycleTestController) cycleLoop(stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case <-s.cancelCtx.Done():
			return
		default:
			s.handleExecuteCycle(s.cancelCtx)
		}
	}
}

func (s *kettleCycleTestController) handleStop() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeTrial == nil {
		return nil, fmt.Errorf("no active trial to stop")
	}

	// Signal the loop to stop
	close(s.activeTrial.stopCh)

	result := map[string]interface{}{
		"trial_id":    s.activeTrial.trialID,
		"cycle_count": s.activeTrial.cycleCount,
	}
	s.activeTrial = nil

	return result, nil
}

func (s *kettleCycleTestController) handleStatus() (map[string]interface{}, error) {
	return s.GetState(), nil
}

func (s *kettleCycleTestController) GetState() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeTrial == nil {
		return map[string]interface{}{
			"state":         "idle",
			"trial_id":      "",
			"cycle_count":   0,
			"last_cycle_at": "",
			"should_sync":   false,
		}
	}

	lastCycleAt := ""
	if !s.activeTrial.lastCycleAt.IsZero() {
		lastCycleAt = s.activeTrial.lastCycleAt.Format(time.RFC3339)
	}

	return map[string]interface{}{
		"state":         "running",
		"trial_id":      s.activeTrial.trialID,
		"cycle_count":   s.activeTrial.cycleCount,
		"last_cycle_at": lastCycleAt,
		"should_sync":   true,
	}
}

// formatCaptureTags creates tags for image upload based on trial state.
func formatCaptureTags(trialID string, cycleCount int) []string {
	tid := trialID
	if tid == "" {
		tid = "standalone"
	}
	return []string{
		fmt.Sprintf("trial_id:%s", tid),
		fmt.Sprintf("cycle_count:%d", cycleCount),
	}
}

// readDataAPICredentials reads API credentials from /etc/viam-data-credentials.json
// HACK: This is a workaround for hot-reloaded (unregistered) modules which don't
// support env var config in the Viam app UI. Once the module is published to the
// registry, this should be replaced with proper env var configuration.
func readDataAPICredentials() (apiKey, apiKeyID string, err error) {
	data, err := os.ReadFile("/etc/viam-data-credentials.json")
	if err != nil {
		return "", "", fmt.Errorf("reading credentials file: %w", err)
	}
	var creds struct {
		APIKey   string `json:"api_key"`
		APIKeyID string `json:"api_key_id"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", fmt.Errorf("parsing credentials file: %w", err)
	}
	if creds.APIKey == "" || creds.APIKeyID == "" {
		return "", "", fmt.Errorf("credentials file missing api_key or api_key_id")
	}
	return creds.APIKey, creds.APIKeyID, nil
}

func (s *kettleCycleTestController) Close(ctx context.Context) error {
	s.cancelFunc()
	if s.viamClient != nil {
		s.viamClient.Close()
	}
	return nil
}
