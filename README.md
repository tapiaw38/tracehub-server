# TraceHub

TraceHub is an AI-powered CLI tool that monitors microservice logs in real-time, automatically detects errors, and uses Claude AI to generate fixes and create pull requests.

## Features

- 🔍 **Real-time Log Monitoring** - Monitor multiple microservices simultaneously
- 🤖 **AI-Powered Error Analysis** - Leverage Claude AI to analyze errors and suggest fixes
- 📊 **Interactive Dashboard** - Beautiful TUI built with Bubbletea
- 🔧 **Auto-Fix Generation** - Get code fixes with explanations and tests
- 🌿 **Git Integration** - Automatically create branches, commits, and PRs
- 🎯 **Smart Error Detection** - Pattern-based error detection with severity classification
- 📝 **Multi-Format Support** - Parse JSON logs, plain text, and custom formats

## Installation

### Prerequisites

- Go 1.21 or higher
- Git
- GitHub CLI (optional, for PR creation)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/tapiaw38/tracehub.git
cd tracehub

# Build
go build -o tracehub ./cmd/tracehub

# Install
sudo mv tracehub /usr/local/bin/
```

## Quick Start

### 1. Generate Configuration

```bash
tracehub init
```

This creates a `tracehub.yaml` file in your current directory.

### 2. Configure Your Services

Edit `tracehub.yaml` to match your setup:

```yaml
services:
  - name: payment-service
    log_path: /var/log/payment/app.log
    format: json
    repo_path: /home/user/projects/payment

detection:
  patterns:
    - "panic:"
    - "nil pointer"
  severity_keywords:
    critical: ["panic", "fatal"]
    high: ["error", "failed"]
    medium: ["warning", "timeout"]

git:
  branch_prefix: "autofix/"
  commit_prefix: "[AutoFix]"

claude:
  model: "claude-sonnet-4-20250514"
  max_tokens: 4000
```

### 3. Set Up Environment Variables

```bash
# Required: Claude API key
export ANTHROPIC_API_KEY="sk-ant-your-key-here"

# Optional: GitHub token for PR creation
export GITHUB_TOKEN="ghp_your-token-here"
```

### 4. Start Monitoring

```bash
# Monitor all configured services
tracehub monitor

# Monitor specific services only
tracehub monitor --services payment,transfer
```

## Usage

### Interactive Dashboard

Once running, you'll see the TraceHub dashboard:

```
┌─ TraceHub Dashboard ─────────────────────────────┐
│                                                   │
│ ● payment-service      │ 45 req/s │ 0.2% errors │
│ ● account-service      │ 120 req/s│ 0.0% errors │
│ ⚠ transfer-service     │ 23 req/s │ 5.1% errors │
│                                                   │
│ Recent Errors:                                    │
│ [15:34:22] transfer-service: nil pointer at L145 │
│                                                   │
│ [F1: Services] [F2: Errors] [A: Analyze] [Q: Quit]│
└───────────────────────────────────────────────────┘
```

### Keyboard Shortcuts

**Dashboard View:**
- `↑/↓` or `k/j` - Navigate services
- `Enter` - View service details
- `E` - View errors list
- `A` - Analyze error with AI
- `Q` - Quit

**Error Analysis:**
- `A` - Analyze selected error
- `C` - Create PR with fix
- `R` - Reject proposed fix
- `ESC` - Go back

### Workflow

1. **Monitor** - TraceHub watches your log files in real-time
2. **Detect** - Errors are automatically detected and classified
3. **Analyze** - Press `A` to have Claude analyze an error
4. **Review** - Review the AI-generated fix proposal
5. **Apply** - Press `C` to create a PR with the fix

## Configuration

### Service Configuration

Each service requires:

- `name` - Service identifier
- `log_path` - Path to log file
- `format` - Log format (json, text, plain)
- `repo_path` - Path to Git repository

### Log Formats

**JSON:**
```json
{"timestamp":"2024-01-10T15:34:22Z","level":"error","message":"nil pointer"}
```

**Text:**
```
2024-01-10 15:34:22 [ERROR] nil pointer dereference at service.go:145
```

### Detection Patterns

TraceHub includes built-in patterns for common errors:

- Go panics
- Nil pointer dereferences
- HTTP 5xx errors
- Database connection errors
- Timeout errors

You can add custom patterns in the configuration.

### Claude Models

Supported models:
- `claude-sonnet-4-20250514` (recommended, fast and cost-effective)
- `claude-opus-4-20250514` (most capable, higher cost)

## Architecture

```
tracehub/
├── cmd/tracehub/           # CLI entry point
├── internal/
│   ├── tui/                # Terminal UI (Bubbletea)
│   ├── collector/          # Log reading and parsing
│   ├── detector/           # Error detection and classification
│   ├── ai/                 # Claude AI integration
│   ├── git/                # Git operations
│   └── config/             # Configuration management
├── pkg/models/             # Data models
└── configs/                # Example configurations
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o tracehub ./cmd/tracehub
```

### Adding Custom Error Patterns

Edit your `tracehub.yaml`:

```yaml
detection:
  patterns:
    - "your custom pattern"
    - "another pattern"
```

## Troubleshooting

### Logs Not Appearing

- Check file permissions on log files
- Verify log_path is correct
- Ensure the log format matches configuration

### AI Analysis Fails

- Verify ANTHROPIC_API_KEY is set correctly
- Check internet connectivity
- Ensure you have API credits

### PR Creation Fails

- Set GITHUB_TOKEN environment variable
- Verify GitHub CLI is installed
- Check repository permissions

## Roadmap

### MVP (Current)
- [x] Real-time log monitoring
- [x] Interactive TUI
- [x] Error detection
- [x] Claude AI integration
- [x] Git operations
- [x] Basic PR creation

### Future Features
- [ ] OpenTelemetry traces support
- [ ] Web dashboard (WebSocket + HTML)
- [ ] Slack/Discord notifications
- [ ] Performance analysis (latency, throughput)
- [ ] Multi-language support (Python, Node.js)
- [ ] Historical log replay
- [ ] Markdown/HTML reports

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Credits

Built with:
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration
- [Claude AI](https://anthropic.com) - Error analysis
- [go-git](https://github.com/go-git/go-git) - Git operations

## Support

For issues and questions:
- GitHub Issues: https://github.com/tapiaw38/tracehub/issues
- Documentation: https://github.com/tapiaw38/tracehub/wiki

---

Made with ❤️ by the TraceHub team
