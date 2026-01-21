---
name: dataset-create
description: Create a new Viam dataset for the project
---

Create a dataset in the viamdemo organization:

```bash
viam dataset create --org-id=392f91f0-f4c0-47d9-8d95-143969aa2668 --name=$ARGUMENTS
```

If no name provided, prompt the user for one.

After creation, display the dataset ID for use in config.
