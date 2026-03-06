# IsometricIcon MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server that wraps the [IsometricIcon](https://www.isometricon.com) API, enabling AI assistants to generate isometric icons from text prompts.

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

### Claude Code

Add to your Claude Code MCP settings (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "isometricon": {
      "command": "/path/to/isometricicon-mcp"
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
      "command": "/path/to/isometricicon-mcp"
    }
  }
}
```

## Usage

1. **Login** first with your IsometricIcon account credentials
2. **Generate icons** by providing a text prompt
3. **Check credits** to see your remaining balance

### Tool: `login`

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email` | string | Yes | Your IsometricIcon account email |
| `password` | string | Yes | Your IsometricIcon account password |

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
