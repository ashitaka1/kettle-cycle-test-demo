# Kettle Cycle Testing Demo

A Viam robotics platform demo for appliance R&D labs, demonstrating cycle testing, failure detection, data capture, and alerting.

> **Status:** Milestone 1 complete — module deployed and responding. See [product_spec.md](product_spec.md) for full roadmap.

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

This project is a Viam module providing a `generic` service for orchestrating cycle tests.

- **API:** `rdk:service:generic`
- **Model:** `viamdemo:kettle-cycle-test:controller`
- **Entry point:** `cmd/module/main.go`
- **Implementation:** `module.go` (controller logic)
- **Tests:** `module_test.go` (unit tests for validation, DoCommand, resource lifecycle)

The controller coordinates arm movements, pour cycles, sensor readings, image capture, and failure detection. Using a generic service rather than a component allows the module to orchestrate multiple hardware resources without implementing hardware-specific interfaces.

## Setup

### Machine Configuration

Create a `machine.json` file in the project root with your Viam machine details:

```json
{
  "org_id": "your-org-id",
  "location_id": "your-location-id",
  "machine_id": "your-machine-id",
  "part_id": "your-part-id",
  "machine_address": "your-machine.viam.cloud"
}
```

You can find these values in the Viam app under your machine's settings.

### Adding the Service

In the Viam app, add a generic service to your machine:
- **Name:** `cycle-tester`
- **API:** `rdk:service:generic`
- **Model:** `viamdemo:kettle-cycle-test:controller`

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

**Next:** Milestone 2 will add arm control, moving between saved positions on command.

## Development

### Build and Deploy

```bash
viam module reload-local --part-id <part_id from machine.json>
```

Builds for the target architecture, uploads, and restarts the module on the configured machine.

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
