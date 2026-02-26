#!/usr/bin/env python3
"""Generate Starlight-compatible API reference markdown from openapi.json.

Usage:
    python scripts/gen-api-docs.py [--spec openapi.json] [--out docs/src/content/docs/api]
"""

import json
import sys
from collections import defaultdict
from pathlib import Path

SPEC_PATH = Path("openapi.json")
OUT_DIR = Path("docs/src/content/docs/api")

# Tags to skip (admin, internal endpoints)
SKIP_TAGS = {"untagged"}

# Map tag names to URL-friendly slugs and display labels
TAG_META = {
    "Auth": ("authentication", "Authentication"),
    "Workspaces": ("workspaces", "Workspaces"),
    "Organizations": ("workspaces", "Workspaces"),  # legacy tag alias
    "Projects": ("projects", "Projects"),
    "Environments": ("environments", "Environments"),
    "Services": ("services", "Services"),
    "Applications": ("services", "Services"),  # legacy tag alias
    "Builds": ("builds", "Builds"),
    "Images": ("builds", "Builds"),  # legacy tag alias
    "Deploys": ("deploys", "Deploys"),
    "Releases": ("deploys", "Deploys"),  # legacy tag alias
    "Deployments": ("deploys", "Deploys"),  # legacy tag alias
    "ConfigVars": ("config-vars", "Config Vars"),
    "Configurations": ("config-vars", "Config Vars"),  # legacy tag alias
    "Pipeline": ("pipeline", "Pipeline"),
    "Promotions": ("promotions", "Promotions"),
    "Observability": ("observability", "Observability"),
    "Teams": ("teams", "Teams"),
    "Integrations": ("integrations", "Integrations"),
}


def parse_args():
    spec = SPEC_PATH
    out = OUT_DIR
    args = sys.argv[1:]
    i = 0
    while i < len(args):
        if args[i] == "--spec" and i + 1 < len(args):
            spec = Path(args[i + 1])
            i += 2
        elif args[i] == "--out" and i + 1 < len(args):
            out = Path(args[i + 1])
            i += 2
        else:
            i += 1
    return spec, out


def resolve_ref(spec, ref):
    """Resolve a $ref pointer like '#/components/schemas/Foo'."""
    parts = ref.lstrip("#/").split("/")
    obj = spec
    for p in parts:
        obj = obj.get(p, {})
    return obj


def schema_to_json_example(spec, schema, depth=0):
    """Generate a JSON example from a schema, resolving $ref."""
    if depth > 4:
        return "..."

    if "$ref" in schema:
        schema = resolve_ref(spec, schema["$ref"])

    if "example" in schema:
        return schema["example"]

    schema_type = schema.get("type", "object")

    if schema_type == "object":
        props = schema.get("properties", {})
        if not props:
            return {}
        result = {}
        for name, prop in props.items():
            result[name] = schema_to_json_example(spec, prop, depth + 1)
        return result
    elif schema_type == "array":
        items = schema.get("items", {})
        return [schema_to_json_example(spec, items, depth + 1)]
    elif schema_type == "string":
        if schema.get("format") == "uuid":
            return "uuid"
        if schema.get("format") == "date-time":
            return "2026-01-15T10:30:00Z"
        if schema.get("enum"):
            return schema["enum"][0]
        return "string"
    elif schema_type == "integer":
        return 0
    elif schema_type == "number":
        return 0.0
    elif schema_type == "boolean":
        return True
    else:
        return None


def format_path(path):
    """Strip /api/v1 prefix for display."""
    return path.removeprefix("/api/v1")


def generate_endpoint_section(spec, method, path, operation):
    """Generate markdown for a single endpoint."""
    lines = []
    op_id = operation.get("operationId", operation.get("summary", ""))
    summary = operation.get("summary", op_id)
    description = operation.get("description", "")

    # Heading from operationId, converted to readable
    heading = " ".join(
        w.capitalize() if w[0].islower() else w
        for w in _split_camel(summary)
    )
    lines.append(f"## {heading}")
    lines.append("")
    lines.append(f"```http")
    lines.append(f"{method.upper()} {format_path(path)}")
    lines.append(f"```")
    lines.append("")

    if description:
        lines.append(description)
        lines.append("")

    # Path parameters
    params = operation.get("parameters", [])
    path_params = [p for p in params if p.get("in") == "path"]
    query_params = [p for p in params if p.get("in") == "query"]

    if path_params:
        lines.append("**Path parameters:**")
        lines.append("")
        lines.append("| Name | Type | Description |")
        lines.append("|------|------|-------------|")
        for p in path_params:
            ptype = p.get("schema", {}).get("type", "string")
            desc = p.get("description", "")
            lines.append(f"| `{p['name']}` | {ptype} | {desc} |")
        lines.append("")

    if query_params:
        lines.append("**Query parameters:**")
        lines.append("")
        lines.append("| Name | Type | Required | Description |")
        lines.append("|------|------|----------|-------------|")
        for p in query_params:
            ptype = p.get("schema", {}).get("type", "string")
            req = "Yes" if p.get("required") else "No"
            desc = p.get("description", "")
            lines.append(f"| `{p['name']}` | {ptype} | {req} | {desc} |")
        lines.append("")

    # Request body
    req_body = operation.get("requestBody", {})
    if req_body:
        content = req_body.get("content", {})
        json_content = content.get("application/json", {})
        schema = json_content.get("schema", {})
        if schema:
            example = schema_to_json_example(spec, schema)
            if example:
                lines.append("**Request body:**")
                lines.append("")
                lines.append("```json")
                lines.append(json.dumps(example, indent=2))
                lines.append("```")
                lines.append("")

    # Response
    responses = operation.get("responses", {})
    success_codes = [c for c in responses if c.startswith("2")]
    for code in success_codes[:1]:
        resp = responses[code]
        resp_content = resp.get("content", {}).get("application/json", {})
        schema = resp_content.get("schema", {})
        if schema:
            example = schema_to_json_example(spec, schema)
            if example and example != "...":
                lines.append("**Response:**")
                lines.append("")
                lines.append("```json")
                lines.append(json.dumps(example, indent=2))
                lines.append("```")
                lines.append("")

    return "\n".join(lines)


def _split_camel(s):
    """Split CamelCase into words."""
    words = []
    current = []
    for c in s:
        if c.isupper() and current:
            words.append("".join(current))
            current = [c]
        else:
            current.append(c)
    if current:
        words.append("".join(current))
    return words


def main():
    spec_path, out_dir = parse_args()

    with open(spec_path) as f:
        spec = json.load(f)

    # Group endpoints by tag
    by_tag = defaultdict(list)
    for path, methods in spec.get("paths", {}).items():
        if not path.startswith("/api/v1"):
            continue
        for method, operation in methods.items():
            if method in ("parameters", "servers"):
                continue
            tags = operation.get("tags", ["untagged"])
            for tag in tags:
                if tag not in SKIP_TAGS:
                    by_tag[tag].append((method, path, operation))

    out_dir.mkdir(parents=True, exist_ok=True)

    # Determine which files will be generated so we can remove stale ones.
    generated_slugs = set()
    for tag in by_tag:
        slug, _ = TAG_META.get(tag, (tag.lower(), tag))
        generated_slugs.add(f"{slug}.md")

    # Remove stale generated files, but keep index.md and hand-written pages.
    for f in out_dir.glob("*.md"):
        if f.name == "index.md":
            continue
        # Only remove files that are auto-generated (contain the marker comment).
        content = f.read_text()
        if "Auto-generated from openapi.json" in content:
            f.unlink()

    count = 0
    for tag, endpoints in sorted(by_tag.items()):
        slug, label = TAG_META.get(tag, (tag.lower(), tag))
        filepath = out_dir / f"{slug}.md"

        lines = [
            "---",
            f'title: "{label}"',
            f"description: API reference for {label.lower()} endpoints.",
            "---",
            "",
            f"<!-- Auto-generated from openapi.json — do not edit manually -->",
            "",
        ]

        # Sort endpoints: list/get before create/update/delete
        method_order = {"get": 0, "post": 1, "put": 2, "patch": 3, "delete": 4}
        endpoints.sort(key=lambda e: (method_order.get(e[0], 5), e[1]))

        for method, path, operation in endpoints:
            lines.append(generate_endpoint_section(spec, method, path, operation))
            lines.append("")

        filepath.write_text("\n".join(lines))
        count += 1

    print(f"Generated {count} API reference pages in {out_dir}")


if __name__ == "__main__":
    main()
