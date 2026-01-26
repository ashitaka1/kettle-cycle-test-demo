# Product Requirements

## Purpose
A demo suitable for a kitchen appliance R&D lab, showing the value of Viam's robotics platform capabilities in the domain of product cycle testing.

## User profile

1. A Viam employee demoing suitability of Viam for cycle testing to any appliance R&D lab
2. A developer learning to use Viam.

## Goals

**Goal:** An observer of the demo should understand how the features of the demo apply to the cycle testing domain
**Goal:** An observer of the demo should appreciate how Viam's features make it quick to set up and easy to develop intelligent, sophisticated testing apparatuses.
**Goal:** A developer who reads the project, or builds the demo themselves, learns how to use the included building blocks, how they work together, and learns best practices for developing on Viam.

**Non-Goal:** A fully working proof of concept with rigorous cycle testing best practices

Build time budget: ~8 human coding hours or less.

## Demo Scenario
Kettle handle stress testing — robotic arm grips kettle by handle in a custom fixture. Arm lifts kettle, performs pouring motion, places it back down. Repeat. Demonstrate cycle automation, failure detection, data capture, monitoring, and alerting.

## Features

### Required
- Operator can turn on and off demo cycle testing using CLI
- Operator can initiate a single cycle using CLI
- Operator can run cycles in training mode to collect labeled images for CV model
- Operator gets periodic updates about testing progress with images and stats
- System detects when kettle fails with CV and alerts operator via email
- System logs kettle put-down force profile, mid-pour images, and kettle state

## Milestones

1. ✅ New module created: README exists, demo component performs validation, machine config in repo, basic CLI dev tools function.
2. ✅ Arm moves back and forth between two saved positions (resting, pour-prep) on command.
3. ✅ Trial lifecycle management with start/stop/status commands, continuous background cycling, cycle-sensor component exposes state for data capture with conditional sync field.
4. ✅ Force sensor component captures force profiles during put-down phase via DoCommand coordination pattern, wrapper component pattern, forceReader abstraction (mock or hardware), capture state machine, and waitForArmStopped timing helper.
5. ✅ Camera stores snapshot of pour-prep pose with cycle record via UploadImageToDatasets with per-image tags (trial_id, cycle_count).
6. ⏸️ DEFERRED: Motion service moves arm between saved positions with `LinearConstraint` keeping kettle level. (WIP on feature/motion-linear-constraint - works but path planning needs iteration)
7. ⏸️ DEFERRED: Motion service performs simple tilt-and-return pouring motion while camera captures images. (blocked by M6)
8. Mock vision service returns configurable `handle_intact` or `handle_broken` result for development.
9. System logs kettle state in record, detects mock broken kettle, sends email alert to operator.
10. Training mode: CLI commands run cycles and tag images as `handle_intact` or `handle_broken`.
11. Machine config migrates to fragments with variables.
12. Periodic job sends trial status to operator.
13. CV model trained in Viam from collected images, replaces mock vision service.
14. Repo commits representing milestones get git tags, documented in README with follow-along lesson outline.

### Bonus round
- Graceful pouring arc (computed waypoint trajectory around spout tip)
- Alert notifications include images (requires cloud function)
- Stream deck UI
- Variable put-down force
- Pour aggressiveness

# Technical Design

### Tech stack

- Robotics programming, fleet management, data capture, and ML: Viam, obviously.
- Language: Go

### Hardware
- UFactory Lite6 robotic arm
- Raspberry Pi
- Webcam (monitoring + CV failure detection)
- Load cell under kettle resting position + MCP3008 SPI ADC + instrumentation amplifier (INA125 or AD620)
- Kettle with custom 3D-printed clamp for handle
- Pre-broken handle with removable patch for simulated failure

## Technical Architecture

### Viam Components
- **Arm**: UFactory Lite6 via UFactory module, standard arm API
- **Camera**: Webcam for monitoring and CV input. Builtin webcam support.
- **Sensor**: Load cell via `viam-labs:mcp300x-adc-sensor` module (raw ADC values)
- **Motion service**: Moves arm between positions with `LinearConstraint` to keep kettle level
- **Vision service**: TFLite binary classifier (`handle_intact` / `handle_broken`) trained in Viam, deployed via ML model service
- **Data capture**: Pour images, force profiles, cycle metadata — correlated via cycle ID tags (e.g., `cycle-2026-01-19-042`)
- **Cycle control**: DoCommands for start, stop, and training mode (`--training=intact` or `--training=broken` for CV data collection)
- **Periodic job**: Regular status reports with current state
- **CLI**: Viam CLI for interaction with robot and backend
- **Fragments**: Machine config with variables, demonstrating fleet scaling pattern
- **Alerting**: Built-in email trigger on CV detection

### Modules
- `viam-ufactory-xarm` — Lite6 arm support
- `viam-labs:mcp300x-adc-sensor` — load cell ADC reading
- `vmodutils` — position-saver switches that trigger arm movements to saved positions
- Custom controller module (this project) — orchestrates the pour cycle routine, training mode
  - Uses position-saver switches via toggleswitch.Switch interface
  - Arm is explicit dependency for clear dependency chain

### Additional Demo Elements

- Viam dashboard for live monitoring
- Viam Teleop for manual intervention

### Data Schema

**Data types:** Sensor (load cell, cycle metadata), Image (pour-position snapshots)

**Tags:**
- `trial:{id}` — groups cycles in a test session
- `cycle:{id}` — individual cycle correlation
- `intact kettle` / `broken kettle` — training labels
- `detected intact` / `detected broken` — inference results during testing

### Variables
- put-down force
- alert destination
- snapshots per pour
- pour aggressiveness

# Implementation Status

## Technical Debt
- `cycleLoop()` in module.go ignores errors from `handleExecuteCycle()` - should log failures during continuous trials
- Rename `samplingLoop()` in force_sensor.go to have a verb (e.g., `runSamplingLoop()`)
- Investigate selectively disabling data capture polling when not in a trial (vs relying on `should_sync=false`)
- Force sensor requires `load_cell` config but uses mock when `use_mock_curve=true` - consider making mock a virtual sensor for cleaner config
- **Credentials file hack:** Camera upload reads API keys from `/etc/viam-data-credentials.json` because hot-reloaded (unregistered) modules can't use env var config in Viam app UI. Once module is published to registry, replace with proper env var configuration.
- Lenient error handling: force sensor and camera failures currently log warnings instead of blocking. Trials should not start without all configured components functioning. Add validation at trial start.
- Investigate whether modules can access Data Client without explicit API keys (using machine's inherent auth context) - current impl requires VIAM_API_KEY/VIAM_API_KEY_ID env vars

## Implementation Notes
- Position-saver switches handle arm movements between saved positions (resting, pour-prep)
- Motion service with LinearConstraint deferred - works but needs path planning iteration (see feature/motion-linear-constraint branch)
- Simple tilt-and-return pour deferred until motion service issues resolved
- Mock vision service allows alerting development before CV model is trained
- Training mode tags images for dataset collection without acting on inference results
- Position-saver switches (vmodutils) trigger arm movements; arm is explicit dependency for clarity in service dependency chain
- `execute_cycle` moves: resting → pour_prep → pause (1s) → resting → pause (1s)
- Trial lifecycle: `start` begins continuous cycling in background goroutine, `stop` ends trial and returns count
- Trial-sensor component wraps controller state via `stateProvider` interface for Viam data capture
- `should_sync` field enables conditional data capture (only sync when trial is active)
- Service dependencies work like component dependencies; sensor declares controller as full resource name
- Force sensor wraps `forceReader` interface (mock or sensorForceReader) for hardware abstraction
- Controller calls force sensor's start_capture/end_capture DoCommands, passing trial metadata via parameters
- DoCommand coordination pattern avoids circular dependencies while enabling rich coordination
- Force sensor state machine: idle → waiting (for first non-zero) → active → idle
- `waitForArmStopped()` polls arm.IsMoving() to ensure clean capture timing
- Force sensor returns trial_id/cycle_count from start_capture params, setting should_sync accordingly
- Viam's builder UI sensor test card lets you verify force sensor readings without CLI commands
- Camera captures snapshot at pour-prep after arm stops moving via `waitForArmStopped()` pattern
- Images uploaded to Viam dataset via `datamanager.UploadImageToDatasets()` with per-image tags
- Tags format: `trial_id:<id>` and `cycle_count:<n>` for correlation with sensor data
- API credentials read from `/etc/viam-data-credentials.json` (workaround for unregistered modules)
- Fixed SDK bug workaround: pass empty struct to UploadImageToDatasets opts (nil panics)
- Cycle count incremented at cycle start to ensure camera and force sensor use same count

# Educational Content

## Code Comments
Follow any overall guidelines regarding comments, but as we expect Viam learners to use this project, leave placeholder comments for the education team to consider when new viam features are introduced to the codebase. This should go hand-in-hand with the README updates (see below). Placeholders simply name the feature and use with a "EDUCATION: " prefix.

## README as Learning Resource
README.md is maintained as a learning resource throughout development. Target audience: developers learning the Viam platform. Where notable, best practices and benefits of viam platform should be emphasised.

### README Target Outline

**Part 1: Reference**
1. Project Overview — what it does, demo scenario, 30-second video/gif
2. Architecture — diagram, components, data flow
3. Hardware Setup — parts list, wiring, physical assembly
4. Quick Start — clone, configure machine in Viam, run a cycle

**Part 2: Lessons (Milestone Walkthroughs)**
5. Lesson 1: Your First Viam Module (Milestone 1) — module structure, validation, machine config
6. Lesson 2: Controlling the Arm (Milestones 2, 6) — saved positions, motion service, LinearConstraint
7. Lesson 3: Data Capture & Tagging (Milestones 3-5) — trials, cycles, sensor + image capture, tag correlation
8. Lesson 4: The Pour Cycle (Milestone 7) — simple tilt motion, capturing images mid-pour
9. Lesson 5: Vision & Alerting (Milestones 8-9) — mock vision service, triggers, email alerts
10. Lesson 6: Training Your Model (Milestones 10, 13) — training mode, labeling in Viam, deploying real model
11. Lesson 7: Fragments & Fleet Patterns (Milestone 11) — variables, reusable config
12. Lesson 8: Periodic Status Reports (Milestone 12) — jobs, operator notifications

**Appendix**
- Troubleshooting
- Bonus round features
- Links to Viam docs 

# Milestone Architecture Decisions

## Milestone 5: Camera Snapshot with Tags

### Approach
Controller captures camera image at pour-prep position and uploads directly to Viam dataset with per-image tags using `datamanager.UploadImageToDatasets()`.

### Key Decisions

**Upload mechanism:** Use Data Management Service API (`datamanager.UploadImageToDatasets`), not Data Client API.
- Data Management Service is accessible from module without API keys
- Supports `tags []string` parameter for per-image tagging
- Requires pre-existing dataset (dataset_id in config)

**Dataset lifecycle:** Managed via Viam CLI, not programmatically.
- `viam dataset create/list/delete` commands handle lifecycle
- Add Makefile targets and/or slash commands as convenience wrappers
- Future UI/client app would handle this; CLI is sufficient for demo

**Tagging strategy:** Per-image tags include:
- `trial_id:<id>` — correlates with trial
- `cycle_count:<n>` — correlates with specific cycle
- Additional tags as needed (e.g., `phase:pour_prep`)

**Capture timing:** After arm reaches pour-prep position and stops moving.
- Use existing `waitForArmStopped()` pattern
- Capture before put-down phase begins

**Error handling - strict, not graceful:**
- If camera or data_manager configured but unavailable → error, block trial start
- Cycle testing without essential data capture is a system error, not a degraded mode
- This is a business rule: incomplete data capture defeats the purpose

**Configuration additions:**
```json
{
  "camera": "webcam-1",
  "data_manager": "data_manager-1",
  "dataset_id": "abc123-dataset-id"
}
```

### Admin Commands (Makefile/Slash)
Convenience wrappers around Viam CLI for dataset management:
- `create-dataset` / `/create-dataset <name>`
- `list-datasets` / `/list-datasets`

# Development Process
Collaborative TDD: Claude writes tests, user revises and approves, then implementation proceeds. No implementation without approved tests. All docs kept up to date.
