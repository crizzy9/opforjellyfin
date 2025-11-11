# 🌐 OpforJellyfin Web UI Guide

The OpforJellyfin web interface provides a Sonarr-like experience for managing One Pace downloads.

## Features

- 📺 **Episodes Browser** - Search and filter available One Pace episodes
- 📊 **Activity Monitor** - Real-time download progress tracking
- ⚙️ **Settings Manager** - Configure directories and torrent clients
- 🔗 **Torrent Client Integration** - qBittorrent, Deluge, and Transmission support
- 🔧 **System Tools** - Sync metadata and view system information
- 🎯 **Sonarr-like Interface** - Familiar layout for media server enthusiasts

## Getting Started

### Starting the Web Server

#### Local Installation

```bash
./opfor serve --port 8090
```

#### Docker

```bash
docker-compose up -d
```

Then access the UI at `http://localhost:8090`

## Pages Overview

### 📺 Episodes

Browse and download One Pace episodes with powerful filtering:

- **Search Bar**: Filter episodes by name
- **Title Filter**: Filter by arc/saga name (e.g., "Wano", "Dressrosa")
- **Range Filter**: Filter by chapter range
- **Episode Cards**: Display quality, seeders, and metadata availability
- **Download Button**: Add episode to download queue

Each episode card shows:
- Episode title and arc name
- Chapter range
- Download key number
- Video quality badge (1080p, 720p, etc.)
- Seeder count
- Metadata availability indicator

### 📊 Activity

Monitor active downloads in real-time:

- **Live Progress Bars**: Visual download progress for each episode
- **Status Updates**: See current download status
- **Auto-Refresh**: Updates every 2 seconds
- **Queue Management**: View all queued and active downloads

Download states:
- ⏳ **Downloading** - Actively downloading from peers
- ✅ **Complete** - Download finished, placing files
- 📁 **Placed** - Files organized in Jellyfin structure

### ⚙️ Settings

Configure OpforJellyfin:

#### General Settings
- **Target Directory**: Where One Pace episodes will be stored
  - Example: `/media/One Piece/One Pace`
  - This should be your Jellyfin media folder

#### Torrent Client Settings
- **Client Type**: Choose between qBittorrent, Deluge, Transmission, or Internal
- **Client URL**: URL of your torrent client API (e.g., `http://localhost:8080`)
- **Credentials**: Username and password for authentication
- **Test Connection**: Verify client connectivity before saving

**Supported Clients:**
- **qBittorrent**: Requires WebUI enabled with API v2
- **Deluge**: Requires WebUI plugin with JSON-RPC enabled
- **Transmission**: Requires RPC enabled
- **Internal**: Built-in Go torrent client (no configuration needed)

### 🔧 System

System management and metadata:

- **Sync Metadata**: Update episode information from GitHub repository
- **System Information**: View version and configuration
- **Logs**: Monitor application activity (coming soon)

## Architecture

### Technology Stack

- **Backend**: Go with standard `net/http`
- **Frontend**: HTMX for dynamic updates
- **Styling**: Custom CSS with dark theme
- **Data**: Embedded templates and static files

### How It Works

1. **Metadata Syncing**: Pulls episode data from `tissla/one-pace-jellyfin` GitHub repo
2. **Episode Discovery**: Scrapes torrent tracker for available releases
3. **Download Management**: Queues downloads and tracks progress
4. **File Organization**: Automatically organizes downloads into Jellyfin-compatible structure

## Tips and Tricks

### First-Time Setup

1. **Set Target Directory First**: This must match your Jellyfin library path
2. **Sync Metadata**: Run this before browsing episodes
3. **Check Available Episodes**: Browse episodes page to see what's available

### Download Management

- **Quality Selection**: Higher quality = larger file size (1080p recommended)
- **Seeder Count**: More seeders = faster downloads
- **Metadata Badge**: Green badge means proper Jellyfin metadata available

### Jellyfin Integration

1. Create a separate Jellyfin library for One Pace
2. **Disable all metadata fetchers** in Jellyfin library settings
3. Point library to your target directory
4. Keep show **unlocked** for changes
5. After downloads complete, rescan library

## Troubleshooting

### Episodes Not Showing

- Run "Sync Metadata" in System page
- Check that target directory is set in Settings
- Refresh the Episodes page

### Downloads Not Starting

- Verify target directory exists and is writable
- Check that you have sufficient disk space
- Ensure torrent tracker is accessible

### Files Not Appearing in Jellyfin

- Verify target directory matches Jellyfin library path
- Check that Jellyfin library has metadata fetchers disabled
- Try rescanning the Jellyfin library
- Ensure show is unlocked in Jellyfin

## API Endpoints

For advanced users and integrations:

- `GET /api/episodes/list` - List all available episodes
- `GET /api/episodes/search?q=query` - Search episodes
- `POST /api/episodes/download` - Queue episode download
- `POST /api/settings/update` - Update settings
- `POST /api/system/sync` - Sync metadata
- `GET /api/activity/status` - Get download status

## Recently Added

✅ **External torrent client integration** (qBittorrent, Deluge, Transmission)
✅ **Connection testing** for torrent clients
✅ **Real-time status updates** from external clients
✅ **Automatic client detection** and configuration

## Future Features

- 📅 Automatic sync scheduling
- 🔔 Download notifications
- 📈 Download statistics and history
- 🎨 Custom themes
- 👥 Multi-user support
- 🔐 Authentication and security
- 📱 Mobile-responsive improvements

## Security Considerations

### Recommended Setup

- Run behind a reverse proxy (nginx, Caddy, Traefik)
- Use authentication (BasicAuth or OAuth)
- Only expose on local network, not internet
- Use HTTPS in production

### Docker Security

- Change default PUID/PGID to match your user
- Bind to localhost only: `127.0.0.1:8090:8090`
- Use Docker secrets for sensitive configuration

## Support

For issues, feature requests, or contributions:
- GitHub: [tissla/opforjellyfin](https://github.com/tissla/opforjellyfin)
- Create an issue with detailed description and logs
