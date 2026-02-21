{
  description = "OpforJellyfin — One Pace organizer for Jellyfin";

  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";
  };

  outputs = { self, nixpkgs }:
    let
      goVersion = 24;

      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
        };
      });
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      # ── Packages ────────────────────────────────────────────────────────────
      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = "opforjellyfin";
          version = "0.2.0";

          src = ./.;

          vendorHash = null; # set to the real hash after first `nix build`

          nativeBuildInputs = with pkgs; [ git ];

          meta = with pkgs.lib; {
            description = "One Pace downloader and organizer for Jellyfin";
            homepage = "https://github.com/tissla/opforjellyfin";
            license = licenses.gpl3;
            maintainers = [];
            mainProgram = "opfor";
          };
        };
      });

      # ── NixOS module ────────────────────────────────────────────────────────
      nixosModules.default = { config, lib, pkgs, ... }: {
        imports = [ ./nix/module.nix ];
        # Wire the flake package as the default so it doesn't need to be set manually.
        config = lib.mkIf config.services.opforjellyfin.enable {
          services.opforjellyfin.package = lib.mkDefault self.packages.${pkgs.system}.default;
        };
      };

      # ── Dev shells ──────────────────────────────────────────────────────────
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # go (version is specified by overlay)
            go

            # goimports, godoc, etc.
            gotools

            # golangci-lint
            golangci-lint

            # templ code generator
            templ
          ];
        };
      });
    };
}
