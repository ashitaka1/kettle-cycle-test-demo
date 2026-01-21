---
name: dataset-delete
description: Delete a Viam dataset
---

Delete a dataset by ID:

```bash
viam dataset delete --dataset-id=$ARGUMENTS
```

If no dataset ID provided, first list available datasets:

```bash
viam dataset list --org-id=392f91f0-f4c0-47d9-8d95-143969aa2668
```

Then prompt the user to specify which dataset to delete.
