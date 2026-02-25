---
title: Python SDK
description: Programmatic access to the Ancla platform from Python.
---

:::note
The Ancla Python SDK is under development. This section will be populated when the SDK is available.
:::

## Planned features

- Async client built on `httpx`
- Full type coverage with Pydantic models
- CLI-compatible authentication (shared config files)
- Context managers for deploy workflows

## Preview

```python
from ancla import AsyncClient

async with AsyncClient() as ancla:
    apps = await ancla.apps.list(org="my-org", project="my-project")
    for app in apps:
        print(f"{app.slug}: {app.latest_deployment.status}")
```

## Installation

The SDK will be published to PyPI once available:

```bash
pip install ancla-sdk
```
