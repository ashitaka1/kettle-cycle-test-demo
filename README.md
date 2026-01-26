# Kettle Cycle Testing Demo

A Viam robotics platform demo for appliance R&D labs, demonstrating cycle testing, failure detection, data capture, and alerting.

> **Status:** Milestone 5 complete — camera captures snapshot at pour-prep position and uploads to Viam dataset with trial correlation tags. See [product_spec.md](product_spec.md) for full roadmap.

## What This Demo Does

This demo shows how to use Viam's robotics platform for automated product testing. A robotic arm grips a kettle by its handle, lifts it, performs a pouring motion, and sets it back down—repeatedly—to stress-test the handle. Computer vision detects when the handle fails, triggering alerts and stopping the test. All sensor data, images, and events are captured and synced to the cloud for analysis.

Key Viam features demonstrated:
- Modular services for orchestrating complex routines
- Motion planning with constraints (keeping the kettle level during movement)
- Tag-based data correlation across sensors and cameras
- Vision service integration with custom ML models
- Built-in alerting and monitoring
- Hot-reload deployment for rapid iteration

## Module Structure

This project is a Viam module providing two resources:

**Controller Service:**
- **API:** `rdk:service:generic`
- **Model:** `viamdemo:kettle-cycle-test:controller`
- **Implementation:** `module.go`
- **Tests:** `module_test.go`

The controller orchestrates arm movements, trial lifecycle, and cycle execution. Using a generic service allows it to coordinate multiple hardware resources without implementing hardware-specific interfaces.

**Cycle Sensor Component:**
- **API:** `rdk:component:sensor`
- **Model:** `viamdemo:kettle-cycle-test:cycle-sensor`
- **Implementation:** `sensor.go`
- **Tests:** `sensor_test.go`

The sensor exposes controller state (trial ID, cycle count, running status) for Viam data capture. Its `should_sync` field enables conditional data capture—only syncing data when a trial is active.

**Force Sensor Component:**
- **API:** `rdk:component:sensor`
- **Model:** `viamdemo:kettle-cycle-test:force-sensor`
- **Implementation:** `force_sensor.go`
- **Tests:** `force_sensor_test.go`

The force sensor captures force profiles during the put-down phase of each cycle. It demonstrates the "wrapper component" pattern—a virtual component that enriches raw sensor data by observing controller state. During the put-down phase, it captures a rolling array of force samples and reports the maximum force detected.

**Entry Point:**
- `cmd/module/main.go` - Registers both resources with the Viam module system

## Setup

### Machine Configuration

Create a `viam-cli-data.json` file in the project root with your Viam machine details:

```json
{
  "dev_machine": {
    "org_id": "your-org-id",
    "location_id": "your-location-id",
    "machine_id": "your-machine-id",
    "part_id": "your-part-id",
    "machine_address": "your-machine.viam.cloud"
  }
}
```

You can find these values in the Viam app under your machine's settings.

### Adding the Controller Service

In the Viam app, add a generic service to your machine:
- **Name:** `cycle-tester`
- **API:** `rdk:service:generic`
- **Model:** `viamdemo:kettle-cycle-test:controller`

**Configuration attributes:**
```json
{
  "arm": "your-arm-name",
  "resting_position": "resting-switch-name",
  "pour_prep_position": "pour-prep-switch-name",
  "force_sensor": "force-sensor",
  "camera": "webcam-1",
  "dataset_id": "your-dataset-id",
  "part_id": "your-part-id"
}
```

Required fields:
- `arm` - Name of the arm component (explicit dependency)
- `resting_position` - Position-saver switch for the resting pose
- `pour_prep_position` - Position-saver switch for the pour-prep pose

Optional fields:
- `force_sensor` - Name of force sensor component to coordinate with during cycles
- `camera` - Name of camera component for capturing cycle images (requires dataset_id and part_id)
- `dataset_id` - Viam dataset ID for image uploads (required if camera is set)
- `part_id` - Machine part ID for image uploads (required if camera is set)

### Adding the Cycle Sensor

Add a sensor component to expose controller state for data capture:
- **Name:** `cycle-sensor`
- **API:** `rdk:component:sensor`
- **Model:** `viamdemo:kettle-cycle-test:cycle-sensor`

**Configuration attributes:**
```json
{
  "controller": "cycle-tester"
}
```

The sensor depends on the controller service and exposes its state through the standard sensor `Readings()` interface.

### Adding the Force Sensor

Add a sensor component to capture force profiles during put-down:
- **Name:** `force-sensor`
- **API:** `rdk:component:sensor`
- **Model:** `viamdemo:kettle-cycle-test:force-sensor`

**Configuration attributes:**
```json
{
  "load_cell": "adc-sensor",
  "force_key": "value",
  "sample_rate_hz": 50,
  "buffer_size": 100,
  "zero_threshold": 5.0,
  "capture_timeout_ms": 10000
}
```

Configuration fields:
- `load_cell` (optional) - Name of ADC sensor component to read force values from. If omitted, uses internal mock reader.
- `force_key` (optional) - Key in sensor readings map containing force value, defaults to "value"
- `sample_rate_hz` (optional) - Force sampling rate, defaults to 50 Hz
- `buffer_size` (optional) - Maximum samples to retain, defaults to 100
- `zero_threshold` (optional) - Readings below this are considered "zero" (kettle not in contact), defaults to 5.0
- `capture_timeout_ms` (optional) - Timeout for capture window if end_capture not called, defaults to 10000 ms

The force sensor uses a mock reader when no `load_cell` is configured. Hardware integration with MCP3008 ADC is supported via the `load_cell` dependency.

## Milestone 1: Foundation

The module foundation is now in place:

**What's Working:**
- Generic service scaffolding generated with `viam module generate`
- Module builds and packages correctly (see `Makefile` for build targets)
- Hot-reload deployment to remote machines via `viam module reload-local`
- Service registers with Viam RDK and responds to lifecycle events
- DoCommand stub returns "not implemented" (ready for command routing)
- Three unit tests validate resource creation, command handling, and cleanup

**Key Files:**
- `module.go` - Controller implementation
- `module_test.go` - Unit tests
- `cmd/module/main.go` - Module entry point
- `meta.json` - Module metadata for Viam registry
- `Makefile` - Build and test targets

**Viam Concepts Introduced:**
- **Generic services** - orchestration layer that coordinates multiple components
- **Module structure** - how to package and deploy custom Viam functionality
- **Resource lifecycle** - constructor, DoCommand interface, Close method
- **Hot reload** - rapid iteration with `viam module reload-local`

## Milestone 2: Arm Movement

The controller now moves the arm between saved positions using position-saver switches.

**What's Working:**
- `execute_cycle` DoCommand moves arm through a test cycle
- Sequence: resting → pour_prep → pause (1 second) → resting
- Config validation ensures required dependencies (arm, switches) are present
- Position-saver switches (from vmodutils module) trigger arm movements
- Arm is an explicit dependency for clarity in the service dependency chain
- Comprehensive unit tests for config validation and execute_cycle behavior
- Makefile targets for hot-reload deployment and cycle testing

**Key Implementation Details:**
- `/Users/apr/Developer/kettle-cycle-test-demo/module.go` - Config struct with required fields, Validate method, handleExecuteCycle logic
- `/Users/apr/Developer/kettle-cycle-test-demo/module_test.go` - Tests for config validation errors, successful cycle execution, error handling
- `/Users/apr/Developer/kettle-cycle-test-demo/Makefile` - `reload-module` and `test-cycle` targets for development workflow

**Viam Concepts Introduced:**
- **Position-saver switches** - vmodutils module provides switches that trigger arm movements to saved positions
- **Explicit dependencies** - arm declared in config.Validate() for clear dependency ordering
- **Config validation** - Validate method returns required dependencies and validates user input
- **DoCommand routing** - switch statement handles different commands (currently just "execute_cycle")

**Testing the cycle:**
```bash
make reload-module  # Deploy to robot
make test-cycle     # Trigger execute_cycle via CLI
```

Or manually via Viam CLI:
```bash
viam machine part run --part <part_id> \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "cycle-tester", "command": {"command": "execute_cycle"}}'
```

## Milestone 3: Trial Lifecycle and Data Capture Readiness

The controller now manages trial lifecycle with continuous cycling and exposes state for data capture.

**What's Working:**
- `start` DoCommand begins a trial and starts continuous background cycling
- `stop` DoCommand ends the trial and returns cycle count
- `status` DoCommand returns current trial state
- Automatic cycle counting during active trials
- `GetState()` method exposes trial metadata for the sensor component
- `should_sync` field in state enables conditional data capture (true during trials, false when idle)
- New cycle-sensor component provides Viam data capture integration
- Makefile targets for trial management: `trial-start`, `trial-stop`, `trial-status`
- Background cycling loop runs until stopped or module closes

**Key Implementation Details:**
- `/Users/apr/Developer/kettle-cycle-test-demo/module.go` - `trialState` struct, `cycleLoop` goroutine, `GetState()` method, trial management commands
- `/Users/apr/Developer/kettle-cycle-test-demo/sensor.go` - Sensor component that wraps controller state, `stateProvider` interface for dependency injection
- `/Users/apr/Developer/kettle-cycle-test-demo/cmd/module/main.go` - Dual resource registration (controller service + cycle sensor component)
- `/Users/apr/Developer/kettle-cycle-test-demo/Makefile` - `trial-start`, `trial-stop`, `trial-status` targets

**Viam Concepts Introduced:**
- **Trial lifecycle** - DoCommand patterns for stateful operations (start/stop/status)
- **Background routines** - Goroutines with cancellation for continuous operation
- **State exposure** - Sensor component wraps service state for data capture
- **Conditional sync** - `should_sync` field controls when data is captured to cloud
- **Service dependencies** - Sensor depends on generic service, not just components
- **Interface-based design** - `stateProvider` interface decouples sensor from controller implementation

**Cycle Sequence:**
Each cycle when a trial is running:
1. Move to pour_prep position
2. Pause 1 second
3. Return to resting position
4. Pause 1 second
5. Increment cycle count
6. Repeat until stopped

**Managing Trials:**
```bash
# Start a trial (begins continuous cycling)
make trial-start

# Check trial status
make trial-status

# Stop the trial
make trial-stop
```

Or manually via Viam CLI:
```bash
# Start
viam machine part run --part <part_id> \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "cycle-tester", "command": {"command": "start"}}'

# Status
viam machine part run --part <part_id> \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "cycle-tester", "command": {"command": "status"}}'

# Stop
viam machine part run --part <part_id> \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "cycle-tester", "command": {"command": "stop"}}'
```

**Sensor Readings:**
Query the cycle-sensor to see trial state:
```bash
viam machine part run --part <part_id> \
  --method 'viam.component.sensor.v1.SensorService.GetReadings' \
  --data '{"name": "cycle-sensor"}'
```

Returns:
```json
{
  "state": "running",
  "trial_id": "trial-20260120-143052",
  "cycle_count": 42,
  "last_cycle_at": "2026-01-20T14:35:12Z",
  "should_sync": true
}
```

When idle:
```json
{
  "state": "idle",
  "trial_id": "",
  "cycle_count": 0,
  "last_cycle_at": "",
  "should_sync": false
}
```

## Milestone 4: Force Profile Capture

The force sensor component now captures force data during the put-down phase of each cycle.

**What's Working:**
- Controller calls `start_capture` DoCommand before arm movement, `end_capture` after arm stops
- Force sensor captures samples at configurable rate (default 50 Hz) between start/end commands
- Waits for first non-zero reading (above threshold) before capturing to skip "air time"
- Captures rolling buffer of force samples with configurable size (default 100 samples)
- Reports trial metadata (trial_id, cycle_count), sample array, sample count, and max force
- Mock force reader simulates realistic force profile: zeros while lifted, ramp on contact
- `should_sync` field true only during active trials (when trial_id present)
- Configurable parameters: sample_rate_hz, buffer_size, zero_threshold, capture_timeout_ms
- `waitForArmStopped()` helper polls arm.IsMoving() to ensure clean capture window

**Key Implementation Details:**
- `/Users/apr/Developer/kettle-cycle-test-demo/force_sensor.go` - Capture state machine, `forceReader` abstraction, `samplingLoop` goroutine, DoCommand handling
- `/Users/apr/Developer/kettle-cycle-test-demo/module.go` - Lines 165-212 (capture coordination), lines 236-257 (`waitForArmStopped()` helper)

**Viam Concepts Introduced:**
- **Wrapper component pattern** - Virtual component that transforms/enriches raw sensor data (like cropped-camera)
- **DoCommand coordination** - Controller and sensor coordinate via DoCommand interface without circular dependencies
- **Dependency injection** - Trial metadata passed via command parameters, not constructor
- **Interface abstraction** - `forceReader` interface enables mock vs hardware swapping
- **State machine** - Capture states: idle → waiting (for first non-zero) → active → idle
- **Sensor test cards** - Viam builder UI provides test cards for verifying sensor readings without CLI

**Force Sensor Readings:**

Query the force sensor to see captured force profile:
```bash
viam machine part run --part <part_id> \
  --method 'viam.component.sensor.v1.SensorService.GetReadings' \
  --data '{"name": "force-sensor"}'
```

During active trial with put-down data:
```json
{
  "trial_id": "trial-20260120-143052",
  "cycle_count": 42,
  "should_sync": true,
  "samples": [50.0, 51.5, 53.0, 54.5, 56.0, ...],
  "sample_count": 87,
  "max_force": 198.5
}
```

When idle or between put-downs:
```json
{
  "trial_id": "",
  "cycle_count": 0,
  "should_sync": false,
  "samples": [],
  "sample_count": 0
}
```

**Testing with Viam App:**

The force sensor can be tested directly in the Viam app builder UI:
1. Navigate to your machine in the Viam app
2. Find the force-sensor component
3. Click the test card to view live readings
4. Start a trial with `make trial-start`
5. Watch force profiles appear as cycles run

This is much faster than using CLI commands during development.

**Architecture Insight:**

The force sensor demonstrates two key Viam patterns: **wrapper components** and **DoCommand coordination**.

Like Viam's built-in cropped-camera (which takes a camera dependency and returns a cropped region), the force sensor wraps a raw force reader and enriches it with cycle awareness. The `forceReader` interface abstracts away hardware details, allowing a mock implementation during development and real MCP3008 ADC integration later.

**DoCommand coordination** avoids circular dependencies while enabling rich interactions:
- Controller calls force sensor's `start_capture` DoCommand, passing trial_id and cycle_count
- Force sensor begins sampling, waiting for first non-zero reading to skip "air time"
- Controller calls `waitForArmStopped()` to poll arm.IsMoving() until movement completes
- Controller calls `end_capture` DoCommand to finalize the capture
- Force sensor returns sample_count and max_force in response

This pattern provides several benefits:
- **No circular dependencies** - Controller depends on sensor, but sensor doesn't depend on controller
- **Dependency injection** - Trial metadata comes from command parameters, not constructor
- **Separation of concerns** - Physical sensor reading vs cycle-aware data capture
- **Testability** - Mock reader for development, real hardware later
- **Data quality** - Only captures relevant data (during put-down) with proper correlation tags

The controller optionally declares force_sensor in its config. If configured, it coordinates capture timing; if not, cycles run without force data. This loose coupling makes both components easier to test and maintain.

## Milestone 5: Camera Snapshot with Tags

The controller now captures camera images at pour-prep position and uploads them to Viam datasets with correlation tags.

**What's Working:**
- Camera capture at pour-prep position after arm stops moving
- Image upload to Viam dataset via Data Management Service API
- Per-image tags for correlation: `trial_id:<id>` and `cycle_count:<n>`
- `waitForArmStopped()` ensures clean capture timing (no motion blur)
- Image decoding with `image.Decode()` auto-detecting JPEG/PNG format
- API credentials read from `/etc/viam-data-credentials.json`
- Dataset management via slash commands: `/dataset-create`, `/dataset-delete`

**Key Implementation Details:**
- `/Users/apr/Developer/kettle-cycle-test-demo/module.go` - Lines 323-372 (`captureAndUploadImage` method), lines 465-476 (`formatCaptureTags` helper), lines 479-514 (`readDataAPICredentials` workaround)
- Camera configured as optional dependency (like force sensor)
- Cycle count incremented at start of cycle to ensure camera and force sensor use same count

**Viam Concepts Introduced:**
- **Data Management Service API** - `UploadImageToDatasets()` for programmatic dataset upload from modules
- **Per-image tagging** - Tags parameter enables ML training correlation across datasets
- **Image decoding** - `image.Decode()` auto-detects format from raw bytes (JPEG, PNG, etc.)
- **Dataset lifecycle** - Datasets created/managed via Viam CLI, not programmatically
- **API authentication** - Data Client requires API key + key ID for dataset operations

**Dataset Management:**

Create a dataset for the project:
```bash
viam dataset create --org-id <org_id> --name kettle-trial-images
```

Or use the slash command:
```bash
/dataset-create kettle-trial-images
```

List datasets to get the dataset ID:
```bash
viam dataset list --org-id <org_id>
```

**Camera Configuration:**

Add camera to controller config (all three fields required):
```json
{
  "camera": "webcam-1",
  "dataset_id": "abc123-your-dataset-id",
  "part_id": "def456-your-part-id"
}
```

**API Credentials Setup:**

The controller needs API credentials to upload images. Create a credentials file on the robot at `/etc/viam-data-credentials.json`:
```json
{
  "api_key": "your-api-key-here",
  "api_key_id": "your-api-key-id-here"
}
```

Get API credentials from Viam app: Settings → API Keys → Create Key

**Note:** This credentials file approach is a temporary workaround for hot-reloaded (unregistered) modules. Once the module is published to the Viam registry, it will use standard environment variable configuration.

**Image Tags:**

Each uploaded image includes tags for correlation with sensor data:
- `trial_id:trial-20260120-143052` - Links image to specific trial
- `cycle_count:42` - Links image to specific cycle within trial

These tags enable ML model training where you can filter images by trial or cycle, and correlate visual data with force sensor readings from the same cycle.

**Architecture Insight:**

The camera integration demonstrates the **Data Management Service API** pattern for programmatic dataset upload from modules.

Unlike the Data Client SDK (used for querying captured data), the Data Management Service provides direct upload capabilities with per-image tagging. This is essential for ML workflows where each image needs metadata for training correlation.

**Key timing decision:** Images are captured after the arm stops moving at pour-prep position. This ensures:
- No motion blur in captured images
- Consistent camera position across all cycles
- Clean correlation with force sensor data (both use same cycle_count)

The `waitForArmStopped()` pattern (introduced in Milestone 4 for force capture) is reused here, demonstrating how helper functions create consistency across multiple capture workflows.

**Tag format choice:** Tags use colon separator (`trial_id:value`) following Viam dataset conventions. This format is queryable in the Viam app UI and ML model training interface, enabling filtering like "show me all images from trial X" or "compare cycles 1-10 vs 90-100".

**Cycle count timing:** The cycle count is incremented at the start of `handleExecuteCycle()` rather than at the end. This ensures the camera snapshot and force sensor readings both use the same cycle number, even though they're captured at different points in the cycle. Without this, you'd have off-by-one errors where image N is correlated with force data from cycle N+1.

## Development

### Build and Deploy

```bash
make reload-module
```

Builds for the target architecture, uploads, and restarts the module on the configured machine. Uses PART_ID from `viam-cli-data.json`.

Alternatively, use the Viam CLI directly:
```bash
viam module reload-local --part-id <part_id from viam-cli-data.json>
```

### Run Tests

```bash
go test ./...
```

### Local Build

```bash
make module.tar.gz
```

Creates a packaged module ready for upload to the Viam registry.

---

*Full documentation will be added as development progresses. See the [README Target Outline](product_spec.md#readme-target-outline) for planned content.*
