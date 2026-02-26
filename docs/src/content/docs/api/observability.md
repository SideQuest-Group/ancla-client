---
title: Observability
description: API reference for resource metrics and monitoring.
---

The observability API exposes CPU, memory, and pod count history for services.

## Get metrics

```http
GET /orgs/{org}/projects/{project}/observability/?service={svc}&env={env}&range=24h
```

### Query parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `service` | yes | Service slug |
| `env` | yes | Environment slug |
| `range` | no | Time range. One of: `1h`, `6h`, `24h`, `7d`, `30d`. Defaults to `24h`. |

### Response

```json
[
  {
    "timestamp": "2025-01-15T12:00:00Z",
    "cpu_usage_m": 150,
    "memory_usage_bytes": 268435456,
    "pod_count": 2,
    "restart_count": 0
  }
]
```

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 timestamp |
| `cpu_usage_m` | int | CPU usage in millicores |
| `memory_usage_bytes` | int | Memory usage in bytes |
| `pod_count` | int | Number of running pods |
| `restart_count` | int | Container restart count |
