# 🏴‍☠️ OpforJellyfin

![OpforJellyfin-logo](img/opforjellyfin.png)

**Automate download and organization of [One Pace](https://onepace.net) episodes for Jellyfin!**

> ✨ **Web UI** with Sonarr-like interface  
> ✨ **External torrent clients** (qBittorrent, Deluge, Transmission)  
> ✨ **CLI tools** for power users  
> ✨ **Docker support** for easy self-hosting  
> ✨ **Automatic file organization** with Jellyfin standards  
> ✨ **Complete metadata** for perfect library integration

---

## 🌐 Web Interface

OpforJellyfin includes a beautiful web UI for managing your One Pace library!

**Quick Start:**
```bash
./opfor serve --port 8090
# Visit http://localhost:8090
```

**Features:**
- 📺 Browse and search episodes with filters
- ⬇️ One-click downloads to external clients
- 📊 Real-time progress tracking
- ⚙️ Torrent client integration (qBittorrent, Deluge, Transmission)
- 🔄 Metadata sync and system management

See [WEB-UI.md](WEB-UI.md) for complete documentation.

## 🔗 Torrent Client Integration

OpforJellyfin supports multiple torrent clients:

- **qBittorrent** - Recommended (WebUI API v2)
- **Deluge** - Full support (JSON-RPC)
- **Transmission** - Full support (RPC API)
- **Internal Client** - Built-in Go torrent client (fallback)

Configure your preferred client in the Web UI Settings page. See [TORRENT-CLIENTS.md](TORRENT-CLIENTS.md) for detailed setup instructions.

### Quick Setup

1. Start OpforJellyfin web UI: `./opfor serve`
2. Go to **Settings** → Configure your torrent client
3. Click **Test Connection** to verify
4. Start downloading from **Episodes** page!

---

# 📢 NEWS 

## Current known issues (v1.0.1):

### Videos getting sorted into strayfolder?
There are some issues with concurrency due to workers simultaneously trying to create directories. 
This is fixed for next release.

### Have-tag not working?
The Have-tag does not work for bundles in this release. This is due to a misnamed variable. 
This is fixed for next release.

## Future plans:

1. 'sort' command - To target a directory containing files, and renaming/sorting them into your metadata directory.

2. Cache for downloadKeys. This will prevent mismatches between 'list' and 'download'.

3. Seeder-mode.

## 📸 Examples

1. Example command:

   ```bash
   ./opfor list -t wano
   ```

   ![List view example](img/example1.png)

2. Downloading episodes:

   ```bash
   ./opfor download 1 3
   ```

   ![Download view example](img/example2.png)

3. Finished download shows file placement:

   ![Finished download](img/example3.png)

4. Keep track of what you have:

   ```bash
   ./opfor info
   ```

   ![Info](img/example4.png)

5. And much more!

## 🔧 Installation

Choose your preferred method:

### 🐳 Docker (Recommended)

Easiest way to get started with the web UI:

```bash
git clone https://github.com/tissla/opforjellyfin.git
cd opforjellyfin
docker-compose up -d
# Visit http://localhost:8090
```

See [DOCKER.md](DOCKER.md) for complete Docker documentation.

---

## 🚀 Deployment

### 🔐 Verbose Logging (Docker)

To see all internal log messages without rebuilding:

```yaml
# docker-compose.yml or docker run
environment:
  - VERBOSE=true
```

Or with the CLI flag: `./opfor serve --verbose`

---

### 🐳 Gluetun VPN + OpforJellyfin

Routes all torrent traffic through a VPN container using Gluetun, while keeping the web UI accessible.

```yaml
networks:
  servarr:
    ipam:
      config:
        - subnet: 172.155.0.0/24

services:
  gluetun:
    image: qmcgaw/gluetun
    container_name: gluetun
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun:/dev/net/tun
    networks:
      servarr:
        ipv4_address: 172.155.0.2
    ports:
      - 8090:8090       # opforjellyfin web UI
      - 53334:53334     # torrent port (set FIREWALL_VPN_INPUT_PORTS to match)
    volumes:
      - /docker/gluetun:/gluetun
    environment:
      - VPN_SERVICE_PROVIDER=airvpn        # or mullvad, protonvpn, etc.
      - VPN_TYPE=wireguard
      - WIREGUARD_PRIVATE_KEY=<your-key>
      - WIREGUARD_PRESHARED_KEY=<your-key>
      - WIREGUARD_ADDRESSES=<vpn-address>
      - SERVER_COUNTRIES=Canada
      - FIREWALL_VPN_INPUT_PORTS=53334
      - FIREWALL_OUTBOUND_SUBNETS=192.168.1.0/24,172.155.0.0/24
      - HEALTH_VPN_DURATION_INITIAL=120s
    healthcheck:
      test: ping -c 1 www.google.com || exit 1
      interval: 60s
      timeout: 20s
      retries: 5
    restart: unless-stopped

  opforjellyfin:
    image: ghcr.io/tissla/opforjellyfin:latest
    container_name: opforjellyfin
    volumes:
      - /docker/opforjellyfin:/config
      - /data:/data
    environment:
      - TZ=UTC
      - VERBOSE=false   # set to true for debug logs
    command: serve --port 8090
    network_mode: service:gluetun   # shares gluetun's network stack
    depends_on:
      - gluetun
    restart: unless-stopped
```

> **Proxmox note**: If running inside an LXC/VM, make sure `/dev/net/tun` is available. Enable it in the container options or with `mknod /dev/net/tun c 10 200`.

---

### 🔀 Traefik Reverse Proxy

Add these labels to gluetun (since it owns the network interface) to expose opforjellyfin through Traefik:

```yaml
services:
  gluetun:
    # ... (existing config above)
    networks:
      - servarr
      - traefik      # must be on the same network Traefik uses
    labels:
      - "traefik.enable=true"
      # Router
      - "traefik.http.routers.opfor.rule=Host(`opfor.yourdomain.com`)"
      - "traefik.http.routers.opfor.entrypoints=websecure"
      - "traefik.http.routers.opfor.tls=true"
      - "traefik.http.routers.opfor.tls.certresolver=letsencrypt"
      # Service (points to port 8090 inside gluetun's namespace)
      - "traefik.http.services.opfor.loadbalancer.server.port=8090"
      # Optional: add auth middleware (Authelia, Authentik, basicauth, etc.)
      # - "traefik.http.routers.opfor.middlewares=authelia@docker"

networks:
  servarr:
    external: true
  traefik:
    external: true    # must already exist and be connected to Traefik
```

> **Why label gluetun and not opforjellyfin?** Because `opforjellyfin` uses `network_mode: service:gluetun`, it doesn't have its own network namespace — Traefik must route through gluetun.

---

### ❄️ NixOS Homelab

Add to your NixOS flake to deploy opforjellyfin as a proper systemd service:

**`flake.nix`**
```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    opforjellyfin.url = "github:tissla/opforjellyfin";
  };

  outputs = { nixpkgs, opforjellyfin, ... }: {
    nixosConfigurations.yourhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        opforjellyfin.nixosModules.default
        ./configuration.nix
      ];
    };
  };
}
```

**`configuration.nix`**
```nix
{
  services.opforjellyfin = {
    enable = true;
    port = 8090;
    dataDir = "/var/lib/opforjellyfin";
    verbose = false;        # set to true to see all logs in journald
    openFirewall = false;   # set to true to expose locally, use nginx/caddy for external
  };
}
```

Useful commands after enabling:
```bash
# Check service status
systemctl status opforjellyfin

# Follow logs (verbose mode recommended)
journalctl -fu opforjellyfin

# Rebuild and switch
nixos-rebuild switch --flake .#yourhost
```

---

### 📦 Binary Releases

For this program to work, you need to have 'git' installed.

#### [Releases](https://github.com/tissla/opforjellyfin/releases/tag/v1.0.0)

MacOS / Linux:

1. Download the file for your system.

2. Run `chmod +x opfor` to make the file runnable.

3. Run with `./opfor --help` to get started.

Windows:

1. Download the .exe file.

2. Run the file in Powershell or Windows Terminal with `.\opfor.exe --help` to get started.

A terminal that supports unicode-characters is heavily recommended for best experience.

### Build from source

1. **Install Go** (version ≥ 1.23)

2. Clone repo:

   ```bash
   git clone https://github.com/tissla/opforjellyfin.git
   cd opforjellyfin
   ```

3. Build binary:

   ```bash
   go build -o opfor
   ```

## 🚀 Usage (Start Here!)

1. Set your download directory before doing anything else. All your metadata will be stored here, and downloads will be matched to their proper folders.

   ```bash
   ./opfor setDir "/media/One Piece/One Pace"
   ```

1. Find all available episodes with 'list', or use the -t flag to specify a title, or -r flag to specify a chapter-range.

   ```bash
   ./opfor list
   ./opfor list -t Wano
   ./opfor list -r 15-20
   ```

1. Download a torrent by using the downloadkey, displayed in front of the title. You can download one or multiple at the same time.

   ```bash
   ./opfor download 15 16 17
   ```

## 📦 Metadata

I hope to continually update [metadata here!](https://github.com/tissla/one-pace-jellyfin)

The 'sync' command allows the user to stay up to date with new additions to the metadata-repo.

### Steps to make sure Jellyfin doesn't mess with the metadata

1. Create a library with no metadata-fetchers active just for One Pace. Disable all of them!

1. Make sure the show is **unlocked** for changes.

1. Run `./opfor sync` again if Jellyfin messed up your .nfo files before this.

1. Rescan library with **unlocked** metadata and _no fetchers active_.

## 🤝 Contributions

All pull requests are welcome. All criticisms are welcome. I'm here to build and to learn and to get better.

## ❤️ Acknowledgements

- SpykerNZ for his metadata
- Anacrolix awesome torrent lib
- Charm team for cool stuff that I should use more
- One Pace team for their amazing work!

## ⚠️ Disclaimer

This tool is provided **as-is** with no guarantees or warranties.  
Use it at your own risk.

While care has been taken to avoid destructive behavior, this tool manipulates files and downloads torrents – always review the source code and test cautiously.  
The author is not responsible for any damage to your system, loss of data, or violation of terms of service related to the use of this software.

Also note:  
This project is not affiliated with One Pace, Jellyfin, or any content providers. Please respect local laws and copyright regulations.
