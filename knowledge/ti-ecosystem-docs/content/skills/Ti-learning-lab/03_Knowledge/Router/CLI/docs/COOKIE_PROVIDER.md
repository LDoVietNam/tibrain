# Cookie-based Provider Support

Ti CLI can import a browser cookie JSON export and store it in a local multi-profile state file. This is intended for internal/self-hosted services where you have permission to authenticate with cookies.

Do not commit cookies, do not paste them into shell commands, and do not store them in `ti.json`.

## Import cookies

```bash
ti-cli auth import-cookies sharedchat --file ./cookies.json
```

Default storage:

```text
~/.ti/auth/cookies.json
```

The file is written with `0600` permissions on Unix-like systems.

## Inspect without leaking secrets

```bash
ti-cli auth cookies
ti-cli auth status sharedchat
ti-cli auth path
```

The CLI only prints a redacted session preview.

## SharedChat-style provider

For the built-in SharedChat-style provider, configure:

```json
{
  "provider": "sharedchat",
  "cookie_file": "~/.ti/auth/cookies.json",
  "cookie_profile": "sharedchat",
  "cookie_domain": "chat.sharedchat.fun",
  "cookie_api_primary": "https://chat.sharedchat.cc",
  "cookie_api_fallback": "https://chat.sharedchat.fun"
}
```

Then:

```bash
ti-cli provider list
ti-cli ask
```

## Generic cookie HTTP provider

Use `compatible_providers` when a service exposes an OpenAI-compatible `/chat/completions` endpoint and expects cookies:

```json
{
  "compatible_providers": [
    {
      "name": "custom-cookie",
      "type": "cookie_http",
      "base_url": "https://example.internal/v1",
      "chat_path": "/chat/completions",
      "cookie_file": "~/.ti/auth/cookies.json",
      "cookie_profile": "sharedchat",
      "cookie_domain": "example.internal",
      "default_model": "gpt-4o-mini",
      "models": ["gpt-4o-mini"],
      "force_https": true
    }
  ]
}
```

Safety behavior:

- cookie values are never printed;
- cookies are sent only to the configured provider URL;
- `cookie_domain` can restrict accepted cookie domains;
- expired profiles report unhealthy in `provider list`;
- `force_https` rejects non-HTTPS URLs except localhost-style development URLs.
