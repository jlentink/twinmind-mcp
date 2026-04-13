# TwinMind MCP & CLI

Access your [TwinMind](https://twinmind.com) meeting recordings from the command line or through Claude Desktop via MCP.

## Features

- **List recordings** - Browse all your meeting recordings with pagination
- **Get recording details** - View full summary, action items, transcript, and notes
- **Search recordings** - Find recordings by keyword in titles

## Installation

### From GitHub Releases

Download the latest binaries from the [Releases](https://github.com/jlentink/twinmind-mcp/releases) page.

### From Source

Requires Go 1.23+.

```bash
go install github.com/jlentink/twinmind-mcp/cmd/twinmind-cli@latest
go install github.com/jlentink/twinmind-mcp/cmd/twinmind-mcp@latest
```

## Authentication

Before using either tool, authenticate with your TwinMind account:

```bash
twinmind-cli auth login
```

This opens your browser for Google OAuth. After signing in, credentials are stored locally at `~/.config/twinmind/config.yaml`. Tokens auto-refresh on subsequent requests.

Check your auth status:

```bash
twinmind-cli auth status
```

## CLI Usage

### List Recordings

```bash
# List latest 20 recordings
twinmind-cli recordings list

# List with pagination
twinmind-cli recordings list --limit 50 --offset 0

# Output as JSON
twinmind-cli recordings list --json
```

### Get Recording Details

```bash
# Get full details by meeting ID
twinmind-cli recordings get <meeting_id>

# Output as JSON
twinmind-cli recordings get <meeting_id> --json
```

### Search Recordings

```bash
# Search by keyword in titles
twinmind-cli recordings search "standup"

# Output as JSON
twinmind-cli recordings search "standup" --json
```

## MCP Server (Claude Desktop)

Add the following to your Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "twinmind": {
      "command": "/path/to/twinmind-mcp"
    }
  }
}
```

Replace `/path/to/twinmind-mcp` with the actual path to the binary.

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `list_recordings` | List all meeting recordings with optional `limit` and `offset` |
| `get_recording` | Get full details of a recording by `meeting_id` |
| `search_recordings` | Search recordings by keyword `query` |

## Configuration

Configuration is stored at `~/.config/twinmind/config.yaml` (or `$XDG_CONFIG_HOME/twinmind/config.yaml`).

Environment variables (prefix `TWINMIND_`):

| Variable | Description | Default |
|----------|-------------|---------|
| `TWINMIND_API_BASE_URL` | API base URL | `https://api.thirdear.live` |
| `TWINMIND_CONFIG` | Config file path | `~/.config/twinmind/config.yaml` |

## Development

Requires [just](https://github.com/casey/just) as a build tool.

```bash
# Build both binaries
just build

# Run tests
just test

# Run linter
just lint

# Install to GOPATH/bin
just install

# Run CLI directly
just run-cli recordings list

# Clean build artifacts
just clean
```

### Release

Releases are automated via GitHub Actions. Tag a commit and push:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This triggers [GoReleaser](https://goreleaser.com/) to build binaries for Linux, macOS, and Windows (amd64/arm64) and create a GitHub release.

## License

MIT
