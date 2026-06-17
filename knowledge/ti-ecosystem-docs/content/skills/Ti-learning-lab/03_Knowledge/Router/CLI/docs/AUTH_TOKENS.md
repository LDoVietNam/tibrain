# Ti Auth Token Manager

Ti can store provider auth tokens as redacted local profiles. This is intended for legitimate tokens you own, such as tokens copied from an official provider page while logged in.

Ti intentionally does **not** scrape protected application credentials or bypass login flows.

## Windsurf token helper

```bash
ti-cli auth windsurf
```

This prints the official token page flow:

1. Open `https://windsurf.com/editor/show-auth-token` while logged in.
2. Copy your token into a local file, for example `~/.ti/auth/windsurf.token`.
3. Import it:

```bash
ti-cli auth token import windsurf --provider windsurf --file ~/.ti/auth/windsurf.token
```

## Token commands

```bash
ti-cli auth token import windsurf --provider windsurf --file ~/.ti/auth/windsurf.token
echo "$TOKEN" | ti-cli auth token import my-provider --provider generic --stdin

ti-cli auth token list
ti-cli auth token show windsurf
ti-cli auth token remove windsurf
ti-cli auth token path
```

Token profiles are stored under `~/.ti/auth/tokens` with 0600 file permissions where supported. Ti only prints token previews such as `abcd********wxyz`.
