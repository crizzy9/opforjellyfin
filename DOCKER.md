# 🐳 Docker Setup for OpforJellyfin

This guide explains how to run OpforJellyfin using Docker and Docker Compose.

## Prerequisites

- Docker installed on your system
- Docker Compose installed (usually comes with Docker Desktop)

## Quick Start with Web UI

### 1. Build and Start the Service

```bash
docker-compose up -d
```

### 2. Access the Web Interface

Open your browser and navigate to:

```
http://localhost:8090
```

### 3. Initial Setup via Web UI

1. Go to **Settings** and set your target directory (e.g., `/data/One Piece/One Pace`)
2. Go to **System** and click "Sync Metadata"
3. Go to **Episodes** to browse and download One Pace episodes

## CLI Usage (Alternative)

You can still use CLI commands alongside the web UI:

### Initial Setup - Set Target Directory

```bash
docker-compose run --rm opforjellyfin setDir "/data/One Piece/One Pace"
```

### Sync Metadata

```bash
docker-compose run --rm opforjellyfin sync
```

## Usage

### List Available Episodes

List all available One Pace episodes:

```bash
docker-compose run --rm opforjellyfin list
```

Filter by title:

```bash
docker-compose run --rm opforjellyfin list -t wano
```

Filter by chapter range:

```bash
docker-compose run --rm opforjellyfin list -r 15-20
```

### Download Episodes

Download one or more episodes using their download keys:

```bash
docker-compose run --rm opforjellyfin download 1 2 3
```

### Check Download Status

View information about your downloads:

```bash
docker-compose run --rm opforjellyfin info
```

### View Logs

Check the application logs:

```bash
docker-compose run --rm opforjellyfin logs
```

### Clear Downloads

Clear completed downloads:

```bash
docker-compose run --rm opforjellyfin clear
```

## Volume Mounts

The Docker Compose setup creates two volume mounts:

- `./config:/config` - Stores application configuration and metadata
- `./downloads:/data` - Stores downloaded One Pace episodes

You can modify these paths in the `docker-compose.yml` file to match your setup.

## Environment Variables

The following environment variables can be configured in `docker-compose.yml`:

- `PUID` - User ID for file permissions (default: 1000)
- `PGID` - Group ID for file permissions (default: 1000)
- `TZ` - Timezone (default: UTC)

## Running as a Service

To keep the container running in the background (useful for future web UI):

```bash
docker-compose up -d
```

To stop the service:

```bash
docker-compose down
```

## Troubleshooting

### Permission Issues

If you encounter permission issues with mounted volumes, adjust the `PUID` and `PGID` environment variables to match your host user:

```bash
id -u  # Get your user ID
id -g  # Get your group ID
```

Then update these values in `docker-compose.yml`.

### Git Not Found

The Docker image includes git, which is required for syncing metadata. If you see git-related errors, ensure the image was built correctly.

## Integration with Jellyfin

If running Jellyfin in Docker, you can share volumes between containers:

```yaml
services:
  opforjellyfin:
    volumes:
      - jellyfin-media:/data
      
  jellyfin:
    volumes:
      - jellyfin-media:/media

volumes:
  jellyfin-media:
```

This allows OpforJellyfin to download directly into your Jellyfin media directory.
