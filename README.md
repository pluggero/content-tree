# content-tree

**content-tree** is a simple CLI tool that recursively scans a directory, reads all files, and outputs their contents in a plain, annotated format. It's useful for code reviews, archiving, or generating context snapshots for AI.

## Features

- Recursively traverses a directory
- Optionally excludes files or directories using glob patterns
- Outputs files with clear delimiters

## Installation

```
go install github.com/pluggero/content-tree@latest
```

## Usage

```bash
content-tree -path /your/project -exclude "*.log,venv/**,**/obj/**"
```

### Flags

- `-path` (default: `.`): Root directory to scan
- `-exclude`: Comma-separated glob patterns to exclude (e.g. `*.log,venv/**,**/obj/**`)

## Output Format

```
>>> START FILE "relative/path/to/file.ext"
<file contents>
<<< END FILE
```

## License

MIT
