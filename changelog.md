# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Developer Experience Improvements

**Added**
- `/viam-cli-data` skill to expose project Viam metadata (org_id, machine_id, part_id, location_id, machine_address)
- Structured `viam-cli-data.json` configuration file with nested `dev_machine` block

**Changed**
- Renamed `machine.json` → `viam-cli-data.json` for clarity of purpose
- Makefile now uses `.dev_machine.part_id` path for PART_ID extraction
- Updated `.gitignore` to ignore `viam-cli-data.json` instead of `machine.json`

### Milestone 5: Camera Snapshot with Trial Correlation Tags

**Added**
- Camera component integration as optional dependency in controller config
- Image capture at pour-prep position using `camera.Image()` and `image.Decode()`
- Direct dataset upload via `UploadImageToDatasets()` from Data Management Service API
- Per-image tagging with `trial_id:<id>` and `cycle_count:<n>` for ML training correlation
- API credentials read from `/etc/viam-data-credentials.json` (workaround for unregistered modules)
- Dataset management slash commands: `/dataset-create <name>` and `/dataset-delete <id>`
- `.env.example` template for API credentials documentation
- Config validation requiring dataset_id and part_id when camera is configured

**Changed**
- Cycle count now incremented at start of `handleExecuteCycle()` instead of end, ensuring camera and force sensor use same count
- `waitForArmStopped()` timing pattern now used for both force capture and camera capture
- Controller Close method now closes Viam client when camera is configured

**Fixed**
- SDK bug workaround: pass empty struct to `UploadImageToDatasets` opts parameter (nil causes panic)

### Milestone 4: Load Cell Integration for Put-Down Force Capture

**Added**
- Force sensor component (`viamdemo:kettle-cycle-test:force-sensor`) wrapping physical load cells or mock data source
- `start_capture` and `end_capture` DoCommands for controlling capture windows during cycle execution
- Trial metadata (trial_id, cycle_count) passed via dependency injection pattern to force sensor
- Three-state capture lifecycle: idle → waiting (for first non-zero reading) → capturing → idle
- Configurable capture parameters: sample_rate_hz (default 50), buffer_size (default 100), zero_threshold (default 5.0), capture_timeout_ms (default 10000)
- Mock force reader simulating realistic force profile: near-zero while lifted, ramp from 50-200 during contact
- `sensorForceReader` wrapper for integrating physical Viam sensor components via configurable force_key
- `waitForArmStopped` helper ensures capture ends precisely when arm stops moving
- Force sensor readings include trial_id, cycle_count, samples array, sample_count, max_force, capture_state, and should_sync flag
- `GetSamplingPhase()` method to stateProvider interface exposing "put_down" phase for sensor coordination

**Changed**
- Controller integrates force sensor as optional dependency
- Execute_cycle triggers start_capture before resting movement, end_capture after arm stops
- Controller tracks samplingPhase field ("put_down" during capture window, empty otherwise)
- Cycle completion returns force_capture results when sensor is configured

### Milestone 3: Cycle Records and CLI Control

**Added**
- Trial lifecycle management via DoCommand with `start`, `stop`, and `status` commands
- Automatic background cycling when trial is running
- `GetState()` method exposing controller state with `should_sync` field for conditional data capture
- Cycle sensor component (`viamdemo:kettle-cycle-test:cycle-sensor`) for exposing controller state to Viam data capture
- Makefile targets for trial control: `trial-start`, `trial-stop`, `trial-status`
- Trial metadata tracking: trial ID, cycle count, timestamps
- 1-second pause at pour_prep position during each cycle
- 1-second pause at resting position during each cycle

**Changed**
- Controller now tracks active trial state and cycle counts
- `execute_cycle` updates cycle counter when trial is active
- Sensor component depends on controller service via explicit dependency chain

### Milestone 2: Arm Movement Between Saved Positions

**Added**
- `execute_cycle` DoCommand that cycles arm between resting and pour-prep positions
- Config validation for arm, resting_position, and pour_prep_position attributes
- Position-saver switch integration for arm movement control
- `make reload-module` target for hot-reload deployment to robot
- `make test-cycle` target for triggering cycle tests via CLI
- Cross-compilation support for Raspberry Pi (linux/arm64)

**Fixed**
- Makefile test-cycle target now uses correct gRPC method syntax for generic services

### Milestone 1: Module Foundation

**Added**
- Generic service module (`viamdemo:kettle-cycle-test:controller`) with DoCommand stub
- Unit tests for controller lifecycle (NewController, DoCommand, Close)
- Hot reload deployment workflow via `viam module reload-local`
- Makefile with build, test, and packaging targets
- Module metadata (meta.json) for Viam registry integration

**Changed**
- README updated with module structure, Milestone 1 summary, and development instructions

## [0.0.1] - 2026-01-19

### Added
- Project planning documents (product_spec.md, CLAUDE.md)
- Technical decisions for Viam components, data schema, motion constraints
- README target outline with lesson structure
- Claude Code agents for docs, changelog, test review, and retrospectives
