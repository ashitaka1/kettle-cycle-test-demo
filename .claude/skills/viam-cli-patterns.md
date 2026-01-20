---
name: viam-cli-patterns
description: Common Viam CLI command patterns and gRPC method names
---

## Calling DoCommand on a Generic Service

The `--service` flag does NOT exist. Use the full gRPC method name:

```bash
viam machine part run --part <part_id> \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "<service-name>", "command": {"command": "<cmd>", ...}}'
```

**Example for this project:**
```bash
viam machine part run --part $PART_ID \
  --method 'viam.service.generic.v1.GenericService.DoCommand' \
  --data '{"name": "cycle-tester", "command": {"command": "execute_cycle"}}'
```

## Calling DoCommand on a Component

Use `--component` flag (shorthand available):

```bash
viam machine part run --part <part_id> --component <name> --method DoCommand --data '{...}'
```

Or full method:
```bash
viam machine part run --part <part_id> \
  --method 'viam.component.generic.v1.GenericService.DoCommand' \
  --data '{"name": "<component-name>", "command": {...}}'
```

## Getting Machine Status

```bash
viam machine part run --part <part_id> \
  --method 'viam.robot.v1.RobotService.GetMachineStatus' \
  --data '{}'
```

## Viewing Logs

```bash
viam machine logs --machine <machine_id> --count <N> [--keyword <filter>]
```

Note: Uses `--machine` (machine_id), not `--part` (part_id).

## Common gRPC Method Names

| Purpose | Method |
|---------|--------|
| Generic Service DoCommand | `viam.service.generic.v1.GenericService.DoCommand` |
| Generic Component DoCommand | `viam.component.generic.v1.GenericService.DoCommand` |
| Machine Status | `viam.robot.v1.RobotService.GetMachineStatus` |
| Resource Names | `viam.robot.v1.RobotService.ResourceNames` |

## Known CLI Limitations

- **No command to fetch machine config** - must use Viam app UI
- **No `--service` flag** - use full gRPC method name instead
- **Logs use machine_id, not part_id** - different from most other commands
