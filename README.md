# Obsidian Tasks MCP Server

A standalone Go MCP server that parses Obsidian tasks from markdown files in your vault and evaluates core Tasks query filters without relying on the REST API.

This server implements a subset of the query filters from the [Obsidian Tasks plugin](https://github.com/obsidian-tasks-group/obsidian-tasks). See the [Tasks plugin documentation](https://obsidian-tasks-group.github.io/obsidian-tasks/queries/) for complete query syntax and examples.

## Installation

```bash
go install github.com/hairyhenderson/obsidian-tasks-mcp@latest
```

Or install from source:

```bash
git clone https://github.com/hairyhenderson/obsidian-tasks-mcp.git
cd obsidian-tasks-mcp
go install .
```

## Usage

The server runs over stdin/stdout and communicates using the MCP protocol. It takes one or more root directories as command-line arguments:

```bash
obsidian-tasks-mcp -root /path/to/vault -root /path/to/other/vault
```

## MCP Tool: `query_tasks`

The server provides a single MCP tool `query_tasks` that accepts:

- `query` (string, optional): Tasks query string with filters (one filter per line). See the [Tasks plugin query documentation](https://obsidian-tasks-group.github.io/obsidian-tasks/queries/) for supported filters and syntax.
- `rootDirs` (array of strings, required): Root directories to scan for markdown files

Returns an array of task objects with `id`, `description`, `status`, `filePath`, `lineNumber`, `tags`, and `dueDate` fields.

## Configuration for Cursor

Add this to your Cursor MCP settings (typically `~/.cursor/mcp.json` or similar):

```json
{
  "mcpServers": {
    "obsidian-tasks": {
      "command": "obsidian-tasks-mcp",
      "args": [
        "-root",
        "/path/to/your/obsidian/vault"
      ]
    }
  }
}
```

You can specify multiple root directories:

```json
{
  "mcpServers": {
    "obsidian-tasks": {
      "command": "obsidian-tasks-mcp",
      "args": [
        "-root",
        "/path/to/vault1",
        "-root",
        "/path/to/vault2"
      ]
    }
  }
}
```

## Development

Run tests:

```bash
make test
```

Lint:

```bash
make lint
```

## License

This project is licensed under the [MIT License](LICENSE). See the LICENSE file for the full license text.

Copyright (c) 2026 Dave Henderson
