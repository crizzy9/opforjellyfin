# 🔗 Torrent Client Integration Guide

OpforJellyfin supports multiple torrent clients for downloading One Pace episodes. This guide covers setup for each supported client.

## Supported Clients

- **qBittorrent** - Recommended for most users
- **Deluge** - Alternative with good plugin support
- **Transmission** - Lightweight option
- **Internal Client** - Built-in fallback (no setup required)

---

## qBittorrent Setup

### Requirements
- qBittorrent v4.1.0 or higher
- WebUI enabled

### Configuration Steps

1. **Enable WebUI in qBittorrent:**
   - Open qBittorrent
   - Go to `Tools` → `Options` → `Web UI`
   - Check "Enable Web User Interface"
   - Set port (default: 8080)
   - Create username and password
   - Click `Apply`

2. **Configure in OpforJellyfin:**
   - Go to Settings in web UI
   - Set Client Type: `qBittorrent`
   - Client URL: `http://localhost:8080`
   - Username: Your qBittorrent username
   - Password: Your qBittorrent password
   - Click "Test Connection"
   - Save settings

### Docker Setup

If running qBittorrent in Docker:

```yaml
services:
  qbittorrent:
    image: linuxserver/qbittorrent
    ports:
      - "8080:8080"
    volumes:
      - ./qbittorrent/config:/config
      - ./downloads:/downloads
```

Use `http://qbittorrent:8080` as the URL if both containers are on the same network.

---

## Deluge Setup

### Requirements
- Deluge 1.3.0 or higher
- WebUI plugin enabled

### Configuration Steps

1. **Enable WebUI in Deluge:**
   - Open Deluge
   - Go to `Preferences` → `Plugins`
   - Enable "WebUI" plugin
   - Go to `Preferences` → `WebUI`
   - Note the port (default: 8112)
   - Set password (default: deluge)

2. **Configure in OpforJellyfin:**
   - Go to Settings in web UI
   - Set Client Type: `Deluge`
   - Client URL: `http://localhost:8112`
   - Password: Your Deluge password
   - Username: (not required for Deluge)
   - Click "Test Connection"
   - Save settings

### Docker Setup

```yaml
services:
  deluge:
    image: linuxserver/deluge
    ports:
      - "8112:8112"
    volumes:
      - ./deluge/config:/config
      - ./downloads:/downloads
```

---

## Transmission Setup

### Requirements
- Transmission 2.94 or higher
- RPC enabled

### Configuration Steps

1. **Enable RPC in Transmission:**
   - Edit `settings.json` (while Transmission is stopped)
   - Set `"rpc-enabled": true`
   - Set `"rpc-port": 9091`
   - Set `"rpc-authentication-required": true`
   - Set `"rpc-username": "your-username"`
   - Set `"rpc-password": "your-password"`
   - Start Transmission

2. **Configure in OpforJellyfin:**
   - Go to Settings in web UI
   - Set Client Type: `Transmission`
   - Client URL: `http://localhost:9091`
   - Username: Your Transmission username
   - Password: Your Transmission password
   - Click "Test Connection"
   - Save settings

### Docker Setup

```yaml
services:
  transmission:
    image: linuxserver/transmission
    ports:
      - "9091:9091"
    volumes:
      - ./transmission/config:/config
      - ./downloads:/downloads
```

---

## Internal Client

The internal client requires no configuration and is used automatically when no external client is configured.

**Pros:**
- No setup required
- Works out of the box
- No external dependencies

**Cons:**
- No seeding after download
- Limited to download-only
- Cannot manage from separate UI

---

## Complete Docker Compose Example

Here's a complete setup with OpforJellyfin and qBittorrent:

```yaml
version: '3.8'

services:
  opforjellyfin:
    build: .
    container_name: opforjellyfin
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
    volumes:
      - ./config:/config
      - ./downloads:/data
    ports:
      - "8090:8090"
    restart: unless-stopped
    command: serve --port 8090
    networks:
      - media
    depends_on:
      - qbittorrent

  qbittorrent:
    image: linuxserver/qbittorrent
    container_name: qbittorrent
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
      - WEBUI_PORT=8080
    volumes:
      - ./qbittorrent/config:/config
      - ./downloads:/downloads
    ports:
      - "8080:8080"
      - "6881:6881"
      - "6881:6881/udp"
    restart: unless-stopped
    networks:
      - media

networks:
  media:
    driver: bridge
```

**OpforJellyfin Configuration:**
- Client Type: `qBittorrent`
- Client URL: `http://qbittorrent:8080`
- Username: `admin` (default)
- Password: Check qBittorrent logs for initial password

---

## Troubleshooting

### Connection Test Fails

**qBittorrent:**
- Verify WebUI is enabled
- Check firewall settings
- Ensure username/password are correct
- Try accessing WebUI directly in browser

**Deluge:**
- Verify WebUI plugin is enabled
- Check if daemon is running
- Default password is `deluge`
- Try accessing WebUI directly in browser

**Transmission:**
- Verify RPC is enabled
- Check `rpc-whitelist` in settings
- Ensure authentication is enabled
- Try accessing web interface directly

### Downloads Not Starting

1. Check target directory is set in Settings
2. Verify client has write permissions to download directory
3. Check client logs for errors
4. Ensure client has sufficient disk space

### Status Not Updating

1. Verify client is still running
2. Check network connectivity between services
3. Ensure torrent wasn't manually removed from client
4. Check Activity page for error messages

---

## Best Practices

### Security

- Use strong passwords for all clients
- Don't expose torrent clients to internet
- Use VPN if downloading public torrents
- Keep clients updated

### Performance

- Set download limits in client if needed
- Configure upload limits (if seeding)
- Use SSD for temporary downloads if possible
- Monitor disk space regularly

### Organization

- Let OpforJellyfin manage file placement
- Don't manually move files during download
- Keep download directory separate from media library
- Regular cleanup of completed torrents

---

## Client Comparison

| Feature | qBittorrent | Deluge | Transmission | Internal |
|---------|-------------|--------|--------------|----------|
| Ease of Setup | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Performance | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| Features | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐ |
| Resource Usage | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| Seeding Support | ✅ | ✅ | ✅ | ❌ |
| Recommended | ✅ | ✅ | ✅ | Fallback |

---

## FAQ

**Q: Can I use multiple clients?**
A: No, only one client can be active at a time. Configure the one you prefer in Settings.

**Q: Does the internal client seed torrents?**
A: No, the internal client downloads only and doesn't seed.

**Q: Can I still use the CLI with external clients?**
A: Yes, but the CLI `download` command still uses the internal client. Use the web UI for external clients.

**Q: What happens if the client goes offline during download?**
A: Downloads will pause. They'll resume when the client comes back online and you restart them.

**Q: Can I manually add torrents to the client?**
A: Yes, but OpforJellyfin may not track them. It's best to use OpforJellyfin for all One Pace downloads.

**Q: Which client should I choose?**
A: qBittorrent is recommended for most users due to its balance of features, performance, and ease of use.

---

For more information, see:
- [README.md](README.md) - General documentation
- [WEB-UI.md](WEB-UI.md) - Web interface guide
- [DOCKER.md](DOCKER.md) - Docker setup guide
