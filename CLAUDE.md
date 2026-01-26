# Kettle Cycle Testing Demo Project

## Current Milestone
Milestone 5 complete. Camera captures snapshot at pour-prep position and uploads to Viam dataset with trial correlation tags.

Milestones 6-7 (motion service) deferred - WIP on feature/motion-linear-constraint branch. Next: Milestone 8 (mock vision service).

*(Keep this updated whenever a project phase or milestone advances.)*

## Open Questions
- How will we detect a real break? (mechanical design)
- How will we fake a break for testing? (removable patch concept)
- **Discuss with Viam engineers:** Best practices for API keys in modules that need Data Client access. Current approach requires manually configuring DATA_API_KEY/DATA_API_KEY_ID env vars on modules (VIAM_API_KEY is reserved for machine credentials). Is there a better pattern?

## Technical Debt
- `cycleLoop()` in module.go ignores errors from `handleExecuteCycle()` - should log failures during continuous trials
- Rename `samplingLoop()` in force_sensor.go to have a verb (e.g., `runSamplingLoop()`)
- Investigate selectively disabling data capture polling when not in a trial (vs relying on `should_sync=false`)
- Force sensor requires `load_cell` config but uses mock when `use_mock_curve=true` - consider making mock a virtual sensor for cleaner config
- **Credentials file hack:** Camera upload reads API keys from `/etc/viam-data-credentials.json` because hot-reloaded (unregistered) modules can't use env var config in Viam app UI. Once module is published to registry, replace with proper env var configuration.
- Lenient error handling: force sensor and camera failures currently log warnings instead of blocking. Trials should not start without all configured components functioning. Add validation at trial start.
- Investigate whether modules can access Data Client without explicit API keys (using machine's inherent auth context) - current impl requires VIAM_API_KEY/VIAM_API_KEY_ID env vars

# Implementation Notes
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

## Documentation

- [Project Spec](product_spec.md) - full description of project, specs
- [Changelog](changelog.md)
- [README](README.md) - docs especially crafted for Viam learners with architecture and design documentation.

### README Maintenance

The README has a **target outline** (in product_spec.md) and a **backlog** (below). After each milestone or significant change:

1. Update README sections that can now be written based on implemented code
2. Move completed backlog items into the README
3. Add new backlog items for content that requires future work (e.g., video, screenshots of UI that doesn't exist yet)

**Backlog** (content waiting on future implementation):
- Demo video/gif showing module in action
- Hardware setup section (parts list, wiring diagrams, 3D-printed fixture details)
- Architecture diagram showing component relationships
- Screenshots of Viam app configuration
- Lesson 1 walkthrough content for Milestone 1
- Lesson 2 walkthrough content for Milestone 2 (position-saver switches, arm as explicit dependency)
- Lesson 3 walkthrough content for Milestone 3 (trial lifecycle, sensor wrapping service state, conditional data capture)
- Lesson 4 walkthrough content for Milestone 4 (DoCommand coordination, wrapper component pattern, forceReader abstraction, capture state machine, waitForArmStopped timing)
- Lesson 5 walkthrough content for Milestone 5 (camera capture, dataset upload, per-image tagging, image.Decode pattern, API credentials workaround)


## Project Commands

### Slash Commands

| Command | Description |
|---------|-------------|
| `/start-feature <name>` | Create and switch to a feature branch |
| `/viam-cli-data` | Display Viam project metadata (org_id, machine_id, part_id, etc.) |
| `/cycle` | Execute a single test cycle on the arm |
| `/trial-start` | Start continuous trial (background cycling) |
| `/trial-stop` | Stop active trial, return cycle count |
| `/trial-status` | Check trial status and cycle count |
| `/logs [keyword]` | View machine logs (optionally filtered) |
| `/status` | Get machine/component health status |
| `/reload` | Hot-reload module to machine |
| `/gen-module` | Generate new Viam module scaffold |
| `/dataset-create <name>` | Create a new Viam dataset for the project |
| `/dataset-delete [id]` | Delete a Viam dataset |

### Viam CLI

- `viam machine part run --part <part_id> --method <method> --data '{}'` — run commands against the machine
- `viam machine logs --machine <machine_id> --count N` — view machine logs (uses machine_id, not part_id)
- `viam organizations list` — list orgs and their namespaces

**Limitations:** No CLI command to fetch machine config (use Viam app UI). No `--service` flag for generic services (use full gRPC method). See `viam-cli-patterns` skill for details.

### Development Commands
- `go test ./...` — run all unit tests
- `make module.tar.gz` — build packaged module
- `make reload-module` — hot-reload module to robot (uses PART_ID from viam-cli-data.json)
- `make test-cycle` — trigger execute_cycle DoCommand via CLI
- `make trial-start` — start a trial (continuous cycling)
- `make trial-stop` — stop the active trial
- `make trial-status` — check trial status and cycle count

### Module Generation

Use `/gen-module <subtype> <model_name>` slash command. Tips:
- Use `generic-service` for logic/orchestration, `generic-component` for hardware
- `--public-namespace` must match your Viam org's namespace

### Hot Reload Deployment

Use `/reload` or `make reload-module`. Builds, packages, uploads via shell service, and restarts the module on the target machine.

### TODO: Machine Config Sync
Create a CLI tool/script to pull current machine config from Viam and store in repo, so machine construction is captured in version control.

## Troubleshooting

### Module Won't Register
- **Symptom:** Module uploads but doesn't appear in machine config
- **Fix:** Ensure your org's public namespace is set in Viam app (Settings → Organization)
- **Note:** Namespace in `viam module generate --public-namespace` must match org setting

### Hot Reload Fails
- **Symptom:** `viam module reload-local` hangs or errors
- **Fix:** Check `viam-cli-data.json` has correct part_id and machine is online

### Service vs Component Mismatch
- **Symptom:** "unknown resource type" error in logs
- **Fix:** Ensure the API in machine config matches the module registration:
  - `rdk:service:generic` → use `generic-service` subtype, `RegisterService()` in code
  - `rdk:component:generic` → use `generic-component` subtype, `RegisterComponent()` in code

### Debugging Workflow

1. **Check if module is healthy:** `/status`
2. **View recent errors:** `/logs error`
3. **Test the cycle:** `/cycle`
4. **View all logs:** `/logs`

## Reference

### Terms

- A complete cycle test for a given specimen is a "trial".
- A single cycle in that trial is a "cycle".

### Benefit notes:

- Using motion to program pouring decouples it from the saved poses and the particulars of the fixture that grips the kettle. 
- machine builder test cards remove need for writing hello world scripts and testing connectivity, even with CV components
- 

### Demo Flow
1. 2-minute Viam architecture overview
2. Build the machine in Viam app from scratch (introduce Viam concepts as we go)
3. Show physical setup, explain components
4. Start cycle routine via job component
5. Run several cycles, show data capture in action
6. Trigger simulated failure (remove patch from handle)
7. CV detects failure, alert fires, cycle stops
8. Show captured data: images, force profiles, event log
9. Show fragment configuration with variable substitution
