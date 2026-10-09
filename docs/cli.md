# Command-Line Interface (CLI)

The `conduit` binary provides command-line flags for running the HTTP server, generating starter configuration files, and exporting OpenAPI 3.0 specifications.

---

## Command Syntax

```bash
conduit [flags]
```

---

## Options & Flags

| Flag                         | Short | Argument    | Default       | Description                                                                                    |
| ---------------------------- | :---: | :---------- | :------------ | ---------------------------------------------------------------------------------------------- |
| `--config <path>`            | `-c`  | File path   | Auto-detected | Specifies the configuration file path (JSON, YAML, TOML, or XML).                              |
| `--generate-config <format>` | `-g`  | Format name | None          | Generates a default starter configuration file (`json`, `yaml`, `toml`, `xml`) and terminates. |
| `--export-openapi <path>`    | `-e`  | File path   | None          | Exports the OpenAPI 3.0 specification derived from the database schema to disk and terminates. |
| `-h`, `--help`               |       | None        |               | Displays usage instructions and available command-line flags.                                  |

---

## Configuration Auto-Discovery

When executed without the `--config` (`-c`) flag, Conduit searches the current working directory in the following precedence order:

1. `config.json`
2. `config.yaml`
3. `config.yml`
4. `config.toml`
5. `config.xml`

If no configuration file is discovered, the server prints an error and exits:

```
failed to load configuration: no config file found in fallback order (json, yaml, toml, xml)
```

To specify an explicit configuration path:

```bash
conduit --config /etc/conduit/production.yaml
# Or short flag:
conduit -c ./my-config.json
```

---

## Generating Default Configurations (`--generate-config`)

You can generate a starter configuration file containing recommended production defaults, comments, and structure using `--generate-config` (`-g`).

Conduit supports generating configurations in four formats:

```bash
# YAML configuration
conduit --generate-config yaml     # Writes config.yaml

# JSON configuration
conduit --generate-config json     # Writes config.json

# TOML configuration
conduit --generate-config toml     # Writes config.toml

# XML configuration
conduit --generate-config xml      # Writes config.xml
```

The generated file includes:

- Network binding settings (`host: "0.0.0.0"`, `port: 8080`, `base_path: "/v1/"`).
- SQLite database configuration targeting `./app.db`.
- Open policy settings for local testing (`publicReads: true`, `publicWrites: true`, `publicMutation: true`).
- Sample database authentication block with table and column mappings.

---

## Exporting OpenAPI Specifications (`--export-openapi`)

To generate an OpenAPI 3.0 specification file for automated tooling, CI/CD verification, or client SDK generation without keeping an HTTP server process running:

```bash
conduit --config ./config.yaml --export-openapi ./openapi.json
# Or using short flags:
conduit -c ./config.yaml -e ./dist/openapi.json
```

Conduit connects to the configured database, introspects the table schemas, compiles the complete OpenAPI 3.0 document formatted with two-space indentation, writes it to the designated target path, and cleanly closes database connections before exiting.

---

## Graceful Teardown

When running as an active HTTP service, Conduit handles OS process signals (`SIGINT`, `SIGTERM`, Ctrl+C) to perform graceful shutdowns:

1. Ceases accepting new incoming connections.
2. Waits up to 10 seconds for active HTTP requests to complete.
3. Closes open authentication provider resources and operational database connections cleanly.
4. Exits with status code `0`.
