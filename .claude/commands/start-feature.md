---
name: start-feature
description: Create and switch to a new feature branch
arguments:
  - name: feature_name
    description: Name of the feature (without 'feature/' prefix)
    required: true
---

Create and switch to a feature branch:

```bash
git checkout -b feature/{{feature_name}}
```

This creates a new branch named `feature/{{feature_name}}` and switches to it, following the project's branching conventions.
