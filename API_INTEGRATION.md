# Open.MP API Integration

## Overview

The omp-launcher-tui is now fully integrated with the **Official Open.MP Server List API** (`https://api.open.mp/servers`).

## What Changed

### Server List Fetching
- **Before**: Used SA-MP UDP protocol to query master server at `master.open.mp:7777`
- **After**: Uses HTTP API at `https://api.open.mp/servers` for real-time server list

### Benefits
✅ No UDP firewall issues  
✅ Faster, more reliable  
✅ JSON response (easier parsing)  
✅ Better error handling  
✅ Automatic fallback to local `servers.json`  
✅ Live server data from official source  

## Server Data Retrieved

From the Open.MP API, the launcher fetches:

```json
{
  "hostname": "Server Name",
  "host": "203.0.113.10",
  "port": 7777,
  "players": 42,
  "maxplayers": 100,
  "gamemode": "Freeroam",
  "language": "English",
  "password": 0
}
```

Displayed in the TUI:
| Field | Source |
|-------|--------|
| Server Name | `hostname` |
| Host:Port | `host:port` |
| Players | `players`/`maxplayers` |
| Passworded | `password > 0` |
| Ping | Live UDP query |

## Configuration

### Default Config
```json
{
  "nickname": "Player",
  "gta_path": "/path/to/GTA",
  "wine_prefix": "",
  "runtime": "auto",
  "master_server": "https://api.open.mp/servers"
}
```

### Custom API Endpoint
To use a different server list API (if hosting your own):

```json
{
  "master_server": "https://your-api.example.com/servers"
}
```

The API endpoint should return a JSON array of server objects.

## Caching System

The launcher implements intelligent server caching:

1. **Cache file**: `servers_cache.json` stores server data including:
   - Server information (name, host, port)
   - Ping and player counts
   - Server rules
   - Last updated timestamp

2. **24-hour validity**: On startup, servers cached less than 24 hours ago are used
3. **Manual refresh**: Pressing R always fetches fresh data
4. **Data merging**: Fresh server lists preserve cached ping data during updates

## Fallback Mechanism

If the API is unreachable (network issues, API down, etc.):

1. Error is logged
2. Cache is loaded from `servers_cache.json`
3. If no cache, `servers.json` (local file) is loaded
4. User can still browse cached servers with ping data
5. Manual refresh can be attempted when connectivity returns

This ensures the app works even without internet connectivity once the cache is populated.

## API Compatibility

The launcher expects the following JSON structure:

```typescript
interface Server {
  hostname: string;      // Server name
  host: string;          // IP or hostname
  port: number;          // Port number
  players: number;       // Current players
  maxplayers: number;    // Max capacity
  gamemode?: string;     // Server gamemode
  language?: string;     // Server language
  password: number;      // 0 = no password, > 0 = passworded
}
```

Optional fields are ignored but won't break parsing.

## Performance

- **Timeout**: 5 seconds (configurable via context deadline)
- **Max response size**: 50 MB
- **Concurrent queries**: 64 workers for ping queries
- **Caching**: In-memory until refresh (R key)

## Network Flow

```
┌──────────────────┐
│   omp-launcher   │
└────────┬─────────┘
         │
         │ HTTP GET
         v
┌──────────────────────────────────┐
│ https://api.open.mp/servers      │
│ (Real-time server list)          │
└────────┬─────────────────────────┘
         │
         │ JSON array of servers
         v
┌──────────────────────────────┐
│ Parse & display in TUI       │
│ - Show server names          │
│ - Show player counts         │
│ - Detect passwords           │
└────────┬────────────────────┘
         │
         │ For selected server:
         │ UDP query for live ping
         v
┌──────────────────────────────┐
│ Selected Server              │
│ (Get actual ping/players)    │
└──────────────────────────────┘
```

## Code Changes

### `internal/server/master.go`
- Replaced UDP socket code with HTTP client
- Parses JSON response into Server objects
- Handles API errors with fallback

### `internal/config/config.go`
- Default master server URL: `https://api.open.mp/servers`
- Config field: `master_server`

### `config.json`
- Updated default to API endpoint

### `README.md`
- Updated documentation to reflect API usage
- Added fallback explanation

## Testing

```bash
# Build
make build

# Run
make run

# The launcher will:
# 1. Fetch from https://api.open.mp/servers
# 2. Display all servers in the list
# 3. Query each for live ping/players
# 4. Allow filtering and selection
```

## Troubleshooting

### "API returned status 404" or "connection refused"
- Check your internet connection
- The API endpoint might be temporarily down
- Fallback to `servers.json` if configured

### "failed to parse API response"
- Ensure the API returns valid JSON array
- Check API endpoint URL in config

### Too many servers displayed
- This is expected - the API returns all registered servers
- Use search (/) to filter by name or IP

## Future Enhancements

Possible improvements:
- Cache API response to disk with TTL
- Support for custom filtering (gamemode, language)
- Support for multiple API endpoints
- Batch server queries with async improvements
- Real-time server updates via WebSocket

## API Documentation

For more details about the Open.MP API:
- https://docs.open.mp/
- https://api.open.mp/

## Compatibility

- ✅ Open.MP servers (primary use)
- ✅ SA-MP servers (if running on Open.MP network)
- ⚠️ Custom API endpoints (must match JSON format)

---

**Status**: ✅ Integrated and tested with Open.MP API
