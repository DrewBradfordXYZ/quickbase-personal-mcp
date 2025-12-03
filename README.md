# QuickBase Personal MCP Server

MCP server providing access to your personal QuickBase SDK repositories (quickbase-js, quickbase-go, quickbase-spec).

## Features

- **Search Code** - Search across all your SDK repos
- **Compare Implementations** - See JS vs Go implementations side-by-side
- **Get Auth Examples** - Quick access to authentication examples
- **List Features** - See what's implemented in your SDKs
- **Check Parity** - Compare feature support between JS and Go

## Installation

```bash
cd ~/MCP/quickbase-personal-mcp
go mod download
go build
```

## Configuration

The server looks for repos at:
- `~/Projects/Personal/quickbase-js`
- `~/Projects/Personal/quickbase-tree/quickbase-go`
- `~/Projects/Personal/quickbase-spec`

Edit `main.go` to customize these paths if your repos are elsewhere.

## Usage

Add to your Claude Code settings:

```json
{
  "mcpServers": {
    "quickbase-personal": {
      "type": "stdio",
      "command": "/Users/drew/MCP/quickbase-personal-mcp/quickbase-personal-mcp",
      "args": []
    }
  }
}
```

## Available Tools

### `search_code`
Search across your SDK repositories.

**Example:**
```json
{
  "query": "ticket auth",
  "repo": "all"
}
```

### `compare_implementations`
Compare JS vs Go implementations.

**Example:**
```json
{
  "feature": "ticket-auth"
}
```

Supported features:
- `ticket-auth`
- `temp-token`
- `user-token`
- `sso`
- `pagination`
- `retry`
- `throttle`

### `get_auth_example`
Get authentication examples.

**Example:**
```json
{
  "auth_type": "ticket",
  "language": "both"
}
```

### `list_features`
List implemented features by category.

### `check_parity`
Check feature parity between SDKs.

## Development

```bash
# Run in development
go run main.go

# Build
go build

# Test with MCP inspector
npx @modelcontextprotocol/inspector go run main.go
```

## License

MIT
