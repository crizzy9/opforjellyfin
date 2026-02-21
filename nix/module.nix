# NixOS Module for OpforJellyfin
{ config, lib, pkgs, ... }:

let
  cfg = config.services.opforjellyfin;
  defaultUser = "opforjellyfin";
in
{
  options.services.opforjellyfin = {
    enable = lib.mkEnableOption "OpforJellyfin — One Pace organizer for Jellyfin";

    package = lib.mkOption {
      type = lib.types.package;
      description = "The opforjellyfin package to use.";
    };

    port = lib.mkOption {
      type = lib.types.port;
      default = 8090;
      description = "Port the web interface listens on.";
    };

    dataDir = lib.mkOption {
      type = lib.types.str;
      default = "/var/lib/opforjellyfin";
      description = "Directory for media library and metadata.";
    };

    configDir = lib.mkOption {
      type = lib.types.str;
      default = "/var/lib/opforjellyfin/config";
      description = "Directory for configuration files.";
    };

    verbose = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Enable verbose logging to stdout (useful when tailing journald).";
    };

    openFirewall = lib.mkOption {
      type = lib.types.bool;
      default = false;
      description = "Open the configured port in the firewall.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = defaultUser;
      description = "User account under which the service runs.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = defaultUser;
      description = "Group under which the service runs.";
    };
  };

  config = lib.mkIf cfg.enable {
    # Create user and group
    users.users.${cfg.user} = lib.mkDefault {
      isSystemUser = true;
      group = cfg.group;
      home = cfg.dataDir;
      description = "OpforJellyfin service user";
    };

    users.groups.${cfg.group} = lib.mkDefault {};

    # Ensure directories exist
    systemd.tmpfiles.rules = [
      "d '${cfg.dataDir}'   0750 ${cfg.user} ${cfg.group} - -"
      "d '${cfg.configDir}' 0750 ${cfg.user} ${cfg.group} - -"
    ];

    # Open firewall if requested
    networking.firewall.allowedTCPPorts = lib.mkIf cfg.openFirewall [ cfg.port ];

    # Systemd service
    systemd.services.opforjellyfin = {
      description = "OpforJellyfin — One Pace organizer for Jellyfin";
      wantedBy = [ "multi-user.target" ];
      after = [ "network.target" ];

      environment = {
        HOME = cfg.configDir;
        VERBOSE = if cfg.verbose then "true" else "false";
      };

      serviceConfig = {
        ExecStart = "${cfg.package}/bin/opfor serve --port ${toString cfg.port}";
        User = cfg.user;
        Group = cfg.group;
        WorkingDirectory = cfg.configDir;
        Restart = "on-failure";
        RestartSec = "10s";

        # Hardening
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ReadWritePaths = [ cfg.dataDir cfg.configDir ];
        ProtectHome = true;
        CapabilityBoundingSet = "";
      };
    };
  };
}
