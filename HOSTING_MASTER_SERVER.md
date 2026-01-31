# Hosting Your Own Master Server List

This document explains how to host your own master server list compatible with this SA-MP/Open.MP launcher.

## API Endpoint

Your master server should provide an HTTP/HTTPS endpoint that returns a JSON array of server objects.

**Example:** `https://your-domain.com/servers`

## Response Format

The endpoint must return a JSON array following the schema defined in `master-server-schema.json`.

### Required Fields

- `ip` (string): Server IP and port in `ip:port` format (e.g., `"127.0.0.1:7777"`)
- `hn` (string): Server hostname/name (max 255 characters)
- `pc` (integer): Current player count (0-1000)
- `pm` (integer): Maximum player slots (1-1000)

### Optional Fields

- `gm` (string): Gamemode name (max 39 characters)
- `la` (string): Server language (max 39 characters)
- `pa` (boolean): Password required (default: false)

## Example Response

```json
[
  {
    "ip": "127.0.0.1:7777",
    "hn": "My Awesome Roleplay Server",
    "pc": 42,
    "pm": 100,
    "gm": "Roleplay",
    "la": "English",
    "pa": false
  },
  {
    "ip": "192.168.1.50:7778",
    "hn": "Deathmatch Arena",
    "pc": 15,
    "pm": 50,
    "gm": "Deathmatch",
    "la": "Portuguese",
    "pa": true
  }
]
```

## Configuration

To use your custom master server with this launcher:

1. Edit `~/.config/omp-tui/config.json`
2. Set the `master_server` field to your endpoint URL:

```json
{
  "nickname": "Player",
  "gta_path": "/path/to/gta",
  "wine_prefix": "",
  "runtime": "auto",
  "master_server": "https://your-domain.com/servers"
}
```

## Implementation Tips

### Static File Hosting

The simplest approach is to host a static JSON file:

```bash
# Generate servers.json
echo '[{"ip":"127.0.0.1:7777","hn":"Test Server","pc":0,"pm":100}]' > servers.json

# Serve with any web server
python3 -m http.server 8000
# Access at: http://localhost:8000/servers.json
```

### Dynamic API

For a dynamic list that updates in real-time:

```python
# Flask example
from flask import Flask, jsonify

app = Flask(__name__)

@app.route('/servers')
def servers():
    # Query your database or SA-MP servers
    return jsonify([
        {
            "ip": "127.0.0.1:7777",
            "hn": "Dynamic Server",
            "pc": 10,
            "pm": 100,
            "gm": "Roleplay",
            "la": "English",
            "pa": False
        }
    ])

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000)
```

### CORS Headers

If hosting on a different domain, enable CORS:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET
```

## Validation

Validate your JSON against the schema:

```bash
# Using ajv-cli
npm install -g ajv-cli
ajv validate -s master-server-schema.json -d your-servers.json
```

## Performance Recommendations

- **Caching**: Set appropriate cache headers (e.g., `Cache-Control: max-age=300` for 5 minutes)
- **Compression**: Enable gzip/brotli compression
- **CDN**: Use a CDN for better global performance
- **Rate Limiting**: Implement rate limiting to prevent abuse

## Security

- Use HTTPS to prevent man-in-the-middle attacks
- Validate and sanitize server data before including in the list
- Implement rate limiting to prevent DoS attacks
- Monitor for malicious or fake server entries

## Official Open.MP API

The default master server is hosted at:
- **URL**: `https://api.open.mp/servers`
- **Format**: Same schema as documented above

## Support

For issues or questions about the launcher, please visit the project repository.
