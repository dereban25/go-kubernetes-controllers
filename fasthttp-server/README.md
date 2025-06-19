# FastHTTP Server with Dual Logging

FastHTTP server with Cobra commands, request logging, and dual logging (console + timestamped files).

## Features

- 🚀 FastHTTP server with high performance
- 📊 Cobra CLI with server command and flags
- 🔍 Request ID generation and tracing
- 📝 Dual logging: console + timestamped log files
- ⏱️ Server start/stop times and uptime tracking
- 🌐 JSON responses with timestamps and request IDs

## Quick Start

```bash
# Build and run
go build -o fasthttp-server .
./fasthttp-server server -p 8080 -l debug

# Test endpoints
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/status

# Check logs
ls logs/                    # List log files
tail -f logs/server_*.log   # Follow current log
```

## Logging

- **Console**: All logs visible in terminal
- **Files**: Saved to `logs/server_YYYY-MM-DD_HH-MM-SS.log`
- **Content**: Full session from start to stop with uptime

## Endpoints

- `GET /` - Root endpoint with server info
- `GET /health` - Health check
- `GET /api/v1/status` - Server status with uptime

## Request Tracing

Every request gets a unique UUID that appears in:
- Response JSON (`request_id` field)
- Response header (`X-Request-ID`)
- All related log entries

## Commands

```bash
# Start server
./fasthttp-server server -p 8080 -l debug

# Custom port and log level
./fasthttp-server server -p 3000 -l info

# Help
./fasthttp-server --help
./fasthttp-server server --help
```

## Log File Example

```
2024/12/07 14:30:15.123456 [SYSTEM] Logging started - Console and File: logs/server_2024-12-07_14-30-15.log
2024/12/07 14:30:15.123789 [SERVER] Starting FastHTTP server at 2024-12-07 14:30:15
2024/12/07 14:30:20.456789 [INFO] ID=a1b2c3d4... | Incoming request: GET / from 127.0.0.1
2024/12/07 14:30:20.457123 [REQUEST] ID=a1b2c3d4... | GET / | Status=200 | Duration=1.2ms
2024/12/07 14:35:20.789123 [SERVER] Received shutdown signal at 2024-12-07 14:35:20
2024/12/07 14:35:20.789567 [SYSTEM] Server stopped at 2024-12-07 14:35:20
2024/12/07 14:35:20.789678 [SYSTEM] Total uptime: 5m5.665s
```
