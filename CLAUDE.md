# Kettle Cycle Testing Demo Project

## Current Milestone
Milestone 5 complete. Camera captures snapshot at pour-prep position and uploads to Viam dataset with trial correlation tags.

Milestones 6-7 (motion service) deferred - WIP on feature/motion-linear-constraint branch. Next: Milestone 8 (mock vision service).

*(Keep this updated whenever a project phase or milestone advances.)*

## Documentation

- [Project Spec](product_spec.md) - Technical architecture, decisions, milestones, technical debt, implementation notes
- [Changelog](changelog.md) - Release notes and change history
- [README](README.md) - User-facing documentation for Viam learners with setup guides and walkthrough lessons

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
