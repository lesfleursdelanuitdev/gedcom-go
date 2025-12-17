# GEDCOM CLI Documentation

Complete reference guide for the GEDCOM command-line interface.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Global Flags](#global-flags)
- [Configuration](#configuration)
- [Commands](#commands)
  - [parse](#parse)
  - [validate](#validate)
  - [export](#export)
  - [interactive](#interactive)
  - [search](#search)
- [Examples](#examples)
- [Tips and Tricks](#tips-and-tricks)

---

## Installation

### From Source

```bash
git clone https://github.com/lesfleursdelanuitdev/gedcom-go.git
cd gedcom-go
go build -o gedcom ./cmd/gedcom
```

### Using Go Install

```bash
go install github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom@latest
```

### Verify Installation

```bash
gedcom --version
```

---

## Quick Start

```bash
# Parse a GEDCOM file
gedcom parse file family.ged

# Validate with advanced rules
gedcom validate advanced family.ged --severity warning

# Export to JSON
gedcom export json family.ged -o family.json

# Search for individuals
gedcom search family.ged --name "John" --sex M

# Interactive mode
gedcom interactive family.ged
```

---

## Global Flags

These flags are available for all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | | Path to config file (default: `~/.gedcom/config.json`) |
| `--quiet` | `-q` | Quiet mode (suppress progress bars) |
| `--verbose` | `-v` | Verbose output |
| `--no-color` | | Disable colored output |
| `--help` | `-h` | Show help |
| `--version` | | Show version |

### Examples

```bash
# Quiet mode
gedcom parse file family.ged --quiet

# Verbose output
gedcom validate advanced family.ged --verbose

# Disable colors
gedcom search family.ged --name "John" --no-color
```

---

## Configuration

The CLI uses a JSON configuration file to customize behavior. Configuration is loaded from:

1. Command-line `--config` flag (highest priority)
2. `~/.gedcom/config.json`
3. `~/.config/gedcom/config.json`
4. Default values (lowest priority)

### Configuration File Format

Create `~/.gedcom/config.json`:

```json
{
  "parser": {
    "type": "hierarchical",
    "parallel": true,
    "stream": false
  },
  "validation": {
    "severity_threshold": "warning",
    "strict_mode": false
  },
  "output": {
    "default_format": "table",
    "color": true,
    "progress": true
  },
  "graph": {
    "cache_size": 1000,
    "enable_indexes": true
  },
  "export": {
    "pretty_print": true,
    "indent": 2
  }
}
```

### Configuration Options

| Section | Option | Type | Default | Description |
|---------|--------|------|---------|-------------|
| `parser` | `type` | string | `"hierarchical"` | Parser type (hierarchical/parallel/stream) |
| `parser` | `parallel` | boolean | `true` | Enable parallel parsing |
| `parser` | `stream` | boolean | `false` | Use streaming parser for large files |
| `validation` | `severity_threshold` | string | `"warning"` | Minimum severity (severe/warning/info/hint) |
| `validation` | `strict_mode` | boolean | `false` | Fail on validation errors |
| `output` | `default_format` | string | `"table"` | Default output format |
| `output` | `color` | boolean | `true` | Enable colored output |
| `output` | `progress` | boolean | `true` | Show progress bars |
| `graph` | `cache_size` | integer | `1000` | Query cache size |
| `graph` | `enable_indexes` | boolean | `true` | Enable filtering indexes |
| `export` | `pretty_print` | boolean | `true` | Pretty-print exported files |
| `export` | `indent` | integer | `2` | Indentation level |

### Environment Variables

You can also use environment variables:

- `GEDCOM_CONFIG` - Override config file path
- `GEDCOM_NO_COLOR` - Disable colors
- `GEDCOM_QUIET` - Quiet mode
- `GEDCOM_VERBOSE` - Verbose mode

---

## Commands

### parse

Parse and optionally validate GEDCOM files.

#### Subcommands

##### `parse file`

Parse a GEDCOM file and optionally export to different formats.

**Usage:**
```bash
gedcom parse file <input.ged> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file (JSON/XML/YAML) |
| `--format` | `-f` | Output format (json/xml/yaml) |
| `--parallel` | | Use parallel parser |
| `--stream` | | Use streaming parser for large files |
| `--verbose` | `-v` | Show detailed parsing info |

**Examples:**

```bash
# Basic parse
gedcom parse file family.ged

# Parse and export to JSON
gedcom parse file family.ged -o family.json -f json

# Parse with verbose output
gedcom parse file family.ged --verbose
```

##### `parse validate`

Parse a GEDCOM file with full validation.

**Usage:**
```bash
gedcom parse validate <input.ged> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--strict` | Fail on errors |

**Examples:**

```bash
# Parse with validation
gedcom parse validate family.ged

# Strict mode (fail on errors)
gedcom parse validate family.ged --strict
```

##### `parse check`

Perform a quick syntax check on a GEDCOM file.

**Usage:**
```bash
gedcom parse check <input.ged>
```

**Examples:**

```bash
# Quick syntax check
gedcom parse check family.ged
```

---

### validate

Validate GEDCOM files with severity levels.

#### Subcommands

##### `validate basic`

Perform basic validation on a GEDCOM file.

**Usage:**
```bash
gedcom validate basic <input.ged> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--fix` | Attempt to fix common issues (not yet implemented) |
| `--fix-output` | Output file for fixed GEDCOM (not yet implemented) |

**Examples:**

```bash
# Basic validation
gedcom validate basic family.ged
```

##### `validate advanced`

Perform advanced validation with configurable severity levels.

**Usage:**
```bash
gedcom validate advanced <input.ged> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--severity` | | Minimum severity (severe/warning/info/hint) |
| `--output` | `-o` | Report output file |
| `--format` | | Report format (text/json/html) |

**Severity Levels:**

- `severe` - Critical errors that prevent proper parsing
- `warning` - Issues that may cause problems
- `info` - Informational messages
- `hint` - Suggestions for improvement

**Examples:**

```bash
# Advanced validation (default: warning)
gedcom validate advanced family.ged

# Only show severe errors
gedcom validate advanced family.ged --severity severe

# Generate JSON report
gedcom validate advanced family.ged --severity warning -o report.json --format json
```

---

### export

Export GEDCOM data to different formats.

#### Subcommands

##### `export json`

Export a GEDCOM file to JSON format.

**Usage:**
```bash
gedcom export json <input.ged> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output file (required) |
| `--pretty` | | Pretty-print output (default: true) |
| `--indent` | | Indentation level (default: 2) |

**Examples:**

```bash
# Export to JSON
gedcom export json family.ged -o family.json

# Export without pretty-printing
gedcom export json family.ged -o family.json --pretty=false
```

##### `export xml`

Export a GEDCOM file to XML format.

**Usage:**
```bash
gedcom export xml <input.ged> [flags]
```

**Flags:** Same as `export json`

**Examples:**

```bash
# Export to XML
gedcom export xml family.ged -o family.xml
```

##### `export yaml`

Export a GEDCOM file to YAML format.

**Usage:**
```bash
gedcom export yaml <input.ged> [flags]
```

**Flags:** Same as `export json`

**Examples:**

```bash
# Export to YAML
gedcom export yaml family.ged -o family.yaml
```

##### `export gedcom`

Re-export a GEDCOM file to GEDCOM format (useful for normalization).

**Usage:**
```bash
gedcom export gedcom <input.ged> [flags]
```

**Flags:** Same as `export json`

**Examples:**

```bash
# Re-export to GEDCOM
gedcom export gedcom family.ged -o family_normalized.ged
```

---

### interactive

Start interactive mode to query GEDCOM data. Parse a file once, then perform multiple queries.

**Usage:**
```bash
gedcom interactive <input.ged> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--no-graph` | Don't build graph (faster startup, limited queries) |

**Interactive Commands:**

| Command | Aliases | Description |
|---------|---------|-------------|
| `help` | `h` | Show help |
| `exit` | `quit`, `q` | Exit interactive mode |
| `stats` | | Show file statistics |
| `individual <xref>` | `indi`, `i` | Show individual details |
| `family <xref>` | `fam`, `f` | Show family details |
| `search <name>` | | Search individuals by name |
| `parents <xref>` | | Show parents |
| `children <xref>` | | Show children |
| `siblings <xref>` | | Show siblings |
| `spouses <xref>` | | Show spouses |
| `ancestors <xref> [n]` | | Show ancestors (optional max generations) |
| `descendants <xref> [n]` | | Show descendants (optional max generations) |
| `relationship <x1> <x2>` | `rel` | Calculate relationship between two individuals |
| `path <x1> <x2>` | | Find path between two individuals |

**Examples:**

```bash
# Start interactive mode
gedcom interactive family.ged

# Interactive session:
gedcom> stats
Statistics:
  Individuals: 1234
  Families: 567
  Notes: 89
  Sources: 12
  Graph nodes: 1234
  Graph edges: 2345

gedcom> search John
Search results for 'John':
  @I1@: John /Doe/
  @I5@: John /Smith/

gedcom> individual @I1@
Individual: @I1@
  Name: John /Doe/
  Sex: M
  Birth: 1900-01-15
  Death: 1970-03-20

gedcom> parents @I1@
Parents of @I1@:
  @I10@: James /Doe/
  @I11@: Mary /Doe/

gedcom> ancestors @I1@ 3
Ancestors of @I1@ (max 3 generations):
  @I10@: James /Doe/
  @I11@: Mary /Doe/
  @I20@: Robert /Doe/
  ...

gedcom> relationship @I1@ @I2@
Relationship from @I1@ to @I2@:
  Type: 1st Cousin
  Degree: 1
  Removal: 0

gedcom> exit
```

**Note:** Interactive mode uses `go-prompt` for enhanced terminal experience with tab completion. If no TTY is detected, it falls back to simple input mode.

---

### search

Advanced search for individuals with multiple filters, operators, and output options.

**Usage:**
```bash
gedcom search <input.ged> [flags]
```

#### Name Filters

| Flag | Description |
|------|-------------|
| `--name` | Search by name (contains) |
| `--name-exact` | Search by name (exact match) |
| `--name-starts` | Search by name (starts with) |
| `--name-ends` | Search by name (ends with) |

#### Date Filters

| Flag | Description |
|------|-------------|
| `--birth-date` | Birth date (year, range YYYY-YYYY, or before:YYYY/after:YYYY) |
| `--birth-year` | Birth year (shorthand for --birth-date) |
| `--birth-date-before` | Born before year |
| `--birth-date-after` | Born after year |

#### Place Filters

| Flag | Description |
|------|-------------|
| `--birth-place` | Birth place (contains) |

#### Demographics

| Flag | Description |
|------|-------------|
| `--sex` | Sex (M, F, U) |

#### Boolean Filters

| Flag | Description |
|------|-------------|
| `--living` | Living individuals (has no death date) |
| `--deceased` | Deceased individuals (has death date) |
| `--has-children` | Has children |
| `--has-spouse` | Has spouse |
| `--no-children` | Does not have children |
| `--no-spouse` | Does not have spouse |

#### Output Options

| Flag | Short | Description |
|------|-------|-------------|
| `--format` | `-f` | Output format (table, json, yaml, csv, list) |
| `--fields` | | Comma-separated fields to display |
| `--sort` | | Sort by field (name, birth_date, xref) |
| `--sort-desc` | | Sort in descending order |
| `--limit` | `-n` | Limit number of results (0 = no limit, default: 100) |
| `--count-only` | | Only show count of results |
| `--output` | `-o` | Output file |
| `--compact` | | Compact output (xref and name only) |

**Examples:**

```bash
# Search by name
gedcom search family.ged --name "John"

# Multiple filters
gedcom search family.ged \
  --name "John" \
  --sex M \
  --birth-year 1900 \
  --has-children

# Search with date range
gedcom search family.ged \
  --birth-date 1900-1950 \
  --birth-place "New York"

# Count only
gedcom search family.ged --name "John" --count-only

# Sorted results
gedcom search family.ged --name "John" --sort name --limit 10

# Export to JSON
gedcom search family.ged --name "John" --format json -o results.json

# Compact output
gedcom search family.ged --name "John" --compact
```

**Date Format Examples:**

```bash
# Single year
--birth-year 1900

# Year range
--birth-date 1900-1950

# Before year
--birth-date-before 1950

# After year
--birth-date-after 1900
```

---

## Examples

### Complete Workflow

```bash
# 1. Parse and validate
gedcom parse validate family.ged --strict

# 2. Advanced validation with report
gedcom validate advanced family.ged \
  --severity warning \
  -o validation_report.json \
  --format json

# 3. Export to JSON
gedcom export json family.ged -o family.json --pretty

# 4. Search for specific individuals
gedcom search family.ged \
  --name "Smith" \
  --birth-year 1900-1950 \
  --format json \
  -o smiths.json

# 5. Interactive exploration
gedcom interactive family.ged
```

### Batch Processing

```bash
# Process multiple files
for file in *.ged; do
  echo "Processing $file..."
  gedcom parse file "$file" --quiet
  gedcom validate advanced "$file" --severity warning --quiet
done
```

### Integration with Scripts

```bash
#!/bin/bash
# Count individuals in a file
count=$(gedcom search "$1" --count-only --quiet 2>/dev/null | grep -o '[0-9]*')
echo "Found $count individuals"

# Export to JSON for further processing
gedcom export json "$1" -o "${1%.ged}.json" --quiet
```

---

## Tips and Tricks

### Performance

1. **Use `--no-graph` in interactive mode** if you don't need relationship queries:
   ```bash
   gedcom interactive family.ged --no-graph
   ```

2. **Use `--quiet` for scripts** to suppress progress bars:
   ```bash
   gedcom parse file family.ged --quiet
   ```

3. **Limit search results** to avoid processing large result sets:
   ```bash
   gedcom search family.ged --name "John" --limit 10
   ```

### Output Formats

1. **Use JSON for programmatic processing**:
   ```bash
   gedcom search family.ged --name "John" --format json -o results.json
   ```

2. **Use compact mode for quick overviews**:
   ```bash
   gedcom search family.ged --name "John" --compact
   ```

3. **Use table format for human-readable output** (default):
   ```bash
   gedcom search family.ged --name "John" --format table
   ```

### Validation

1. **Start with basic validation** before advanced:
   ```bash
   gedcom validate basic family.ged
   ```

2. **Use severity filtering** to focus on important issues:
   ```bash
   # Only show severe errors
   gedcom validate advanced family.ged --severity severe
   ```

3. **Generate reports** for documentation:
   ```bash
   gedcom validate advanced family.ged \
     --severity warning \
     -o report.json \
     --format json
   ```

### Interactive Mode

1. **Use tab completion** for command names and XREF IDs (if supported)

2. **Use aliases** for faster typing:
   - `i` instead of `individual`
   - `f` instead of `family`
   - `q` instead of `exit`

3. **Check statistics first** to understand the data:
   ```bash
   gedcom> stats
   ```

### Search Tips

1. **Combine multiple filters** for precise searches:
   ```bash
   gedcom search family.ged \
     --name "John" \
     --sex M \
     --birth-year 1900-1950 \
     --has-children
   ```

2. **Use `--count-only`** to quickly check if results exist:
   ```bash
   gedcom search family.ged --name "John" --count-only
   ```

3. **Sort results** for better readability:
   ```bash
   gedcom search family.ged --name "John" --sort name
   ```

---

## Troubleshooting

### Common Issues

1. **"File not found" error**
   - Check file path is correct
   - Use absolute path if needed

2. **"Parse failed" error**
   - File may be corrupted or invalid GEDCOM format
   - Try `gedcom parse check` for syntax errors

3. **Interactive mode not working**
   - Ensure you have a TTY (terminal)
   - Use `--no-graph` if graph building fails

4. **Search returns no results**
   - Check filter syntax
   - Try broader filters first
   - Use `--count-only` to verify query

### Getting Help

```bash
# Command help
gedcom --help
gedcom parse --help
gedcom validate --help
gedcom export --help
gedcom search --help
gedcom interactive --help

# In interactive mode
gedcom> help
```

---

## See Also

- [README.md](../README.md) - Project overview
- [CODEBASE_ANALYSIS_2025.md](../CODEBASE_ANALYSIS_2025.md) - Architecture documentation
- [CLI_DESIGN.md](../CLI_DESIGN.md) - CLI design document
