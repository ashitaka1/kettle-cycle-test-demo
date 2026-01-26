---
name: viam-cli-data
description: Provides Viam CLI data for the project. Use when you need to supply arguments to the viam CLI for the various cloud resources used in the project, like org ids, machine ids, etc.
---

# Viam CLI Data

Exposes Viam identifiers and configuration stored in `viam-cli-data.json`:

```bash
cat viam-cli-data.json
```

## dev_machine

| Field | Description |
|-------|-------------|
| `org_id` | Viam organization ID for the dev machine |
| `machine_id` | Viam machine ID for the dev machine |
| `part_id` | Machine part ID for the dev machine's main part |
| `location_id` | Viam location ID for the dev machine |
| `machine_address` | Cloud hostname for the dev machine |
