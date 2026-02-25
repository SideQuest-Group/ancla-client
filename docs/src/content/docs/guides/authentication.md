---
title: Authentication
description: How to authenticate the Ancla CLI with your server.
---

The Ancla CLI authenticates via API keys stored in your config file. There are several ways to obtain and provide a key.

## Browser login (default)

```bash
ancla login
```

This starts a local callback server and opens your browser to the Ancla server's CLI auth page. After you log in, the API key is sent back to the CLI and saved automatically.

You'll see a confirmation code (e.g. `A1B2-C3D4`) in your terminal — verify it matches what the browser shows to prevent CSRF attacks.

If the browser doesn't open, copy the printed URL manually.

## Manual login

For headless environments or when browser login isn't available:

```bash
ancla login --manual
```

This prompts you to paste an API key directly. The key is validated against the server before saving.

To get an API key manually, log in to the Ancla web UI and navigate to your account settings.

## Environment variables

For CI/CD pipelines and automation, set the API key via environment variable:

```bash
export ANCLA_API_KEY=ancla_your_key_here
```

This takes precedence over config files but is overridden by the `--api-key` flag.

## CLI flag

For one-off commands:

```bash
ancla apps list --api-key ancla_your_key_here
```

:::caution
Avoid using `--api-key` in scripts — prefer environment variables to keep keys out of shell history and process lists.
:::

## Precedence

From highest to lowest:

1. `--api-key` flag
2. `ANCLA_API_KEY` environment variable
3. Local `.ancla/config.yaml` (nearest parent directory)
4. Global `~/.ancla/config.yaml`

## Verifying your session

```bash
ancla whoami
```

Shows the username, email, and admin status of the currently authenticated user.

## Logging out

Remove the stored API key:

```bash
ancla settings set api_key ""
```

Or delete the config file:

```bash
rm ~/.ancla/config.yaml
```
