# Claude Desktop Third-party Provider

Claude Desktop third-party provider support depends on the client version and plugin ecosystem. cloudbase provides the gateway, API key, and usage management.

## Requirements

- Claude Desktop installed.
- cloudbase API key created.
- API Guide opens normally.

## Configuration idea

```text
Base URL: value from cloudbase API Guide
API Key: your cloudbase API key
Model: model enabled by the current group
```

If you connect through MCP, a proxy, or an extension, confirm that layer is not overriding Base URL or API key.

## Security

Only install plugins you trust. Third-party plugins may read local configuration files.
