# IsometricIcon MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that wraps the [IsometricIcon](https://www.isometricon.com) API, enabling AI assistants to generate isometric icons from text prompts.

## Example Output

> Prompt: *"A rocket ship launching into space"*

<p align="center">
  <img src="assets/example.png" width="400" alt="Example isometric icon - rocket ship" />
</p>

## Tools

| Tool | Description |
|------|-------------|
| `login` | Authenticate with your IsometricIcon account |
| `generate_icon` | Generate an isometric icon from a text prompt |
| `check_credits` | Check your remaining credit balance |

## Installation

### From Release

Download the latest binary from the [Releases](https://github.com/igun997/isometricicon-mcp/releases) page.

### From Source

```bash
go install github.com/igun997/isometricicon-mcp@latest
```

Or build locally:

```bash
git clone https://github.com/igun997/isometricicon-mcp.git
cd isometricicon-mcp
go build -o isometricicon-mcp .
```

## Configuration

### Environment Variables

Set these to avoid passing credentials on every login:

```bash
export ISOMETRICON_EMAIL="your@email.com"
export ISOMETRICON_PASSWORD="your-password"
```

When set, the server auto-logs in on startup and the `login` tool uses them as defaults.

The JWT token is cached at `~/.config/isometricon/.isometricon-token.json` and reused across sessions until it expires (~60s).

### Claude Code

Add to your Claude Code MCP settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "isometricon": {
      "command": "/path/to/isometricicon-mcp",
      "env": {
        "ISOMETRICON_EMAIL": "your@email.com",
        "ISOMETRICON_PASSWORD": "your-password"
      }
    }
  }
}
```

### Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "isometricon": {
      "command": "/path/to/isometricicon-mcp",
      "env": {
        "ISOMETRICON_EMAIL": "your@email.com",
        "ISOMETRICON_PASSWORD": "your-password"
      }
    }
  }
}
```

## Usage

1. **Login** first with your IsometricIcon account credentials (automatic if env vars are set)
2. **Generate icons** by providing a text prompt
3. **Check credits** to see your remaining balance

### Tool: `login`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email` | string | No | Your email (falls back to `ISOMETRICON_EMAIL` env var) |
| `password` | string | No | Your password (falls back to `ISOMETRICON_PASSWORD` env var) |

### Tool: `generate_icon`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `prompt` | string | Yes | Text description of the icon to generate |
| `output_path` | string | No | File path to save the PNG (default: `./icon.png`) |

Returns the CDN URL and saves the image locally.

### Tool: `check_credits`

No parameters. Returns your current credit balance.

## Requirements

- An [IsometricIcon](https://www.isometricon.com) account
- Credits for icon generation

## License

MIT
