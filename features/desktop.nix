{
  pkgs,
  lib,
  ...
}:

let
  emoji = pkgs.callPackage ../packages/emoji.nix { };
  audio-change-output = pkgs.callPackage ../packages/audio-change-output.nix { };
  password = pkgs.callPackage ../packages/password.nix { };
  mirror-screen = pkgs.callPackage ../packages/mirror-screen.nix { };
  seedbox-mount = pkgs.callPackage ../packages/seedbox-mount.nix { };
  tailscaleSwitch = import ../packages/tailscale-switch.nix { inherit pkgs; };
in
{
  imports = [
    ./sub-features/sway.nix
    ./sub-features/firefox.nix
    ./sub-features/varasto.nix
    ./sub-features/phone-media-transfer-protocol.nix
  ];

  hardware = {
    graphics.enable = true;
  };

  home-manager = {
    users."joonas" = {
      programs.atuin = {
        # enable = true;
        # enableBashIntegration = true;
        flags = [
          "--disable-up-arrow" # don't take over up arrow as launching history browser
        ];
        settings = {
          # auto_sync = true;
          # sync_frequency = "5m";
          # sync_address = "https://api.atuin.sh";
          search_mode = "fuzzy";
        };
      };

      programs = {
        ghostty = {
          enable = true;

          # terminal starts faster when running in daemon mode
          systemd = {
            enable = true;
          };

          # https://ghostty.org/docs/config/reference
          # when change these `ctrl+shift+,` live-updates these
          settings = {
            # without this if local terminfo DB knows ghostty, it would use "ghostty". the trouble is
            # that SSH forwards the "TERM" ENV and if the remote host doesn't know about ghostty, we'd get
            # a broken terminal.
            term = "xterm-256color";

            background = "#131926";
            foreground = "#c5c8c6";
            background-opacity = 0.8;
            # background-blur = 20; # Sway doesn't support blur

            cursor-style = "block";
            cursor-opacity = 0.5;

            app-notifications = "no-clipboard-copy";
            copy-on-select = "clipboard"; # "copy text to the selection clipboard as well as the system clipboard."

            # doesn't look great. doesn't support ligatures.
            # font-family = "Source Code Pro";
            font-size = "9";

            keybind = [
              "ctrl+shift+r=toggle_mouse_reporting"
            ];
          };
        };
      };

      # You should not change this value, even if you update Home Manager. If you do
      # want to update the value, then make sure to first check the Home Manager
      # release notes.
      home.stateVersion = "25.11"; # Please read the comment before changing.
    };
  };

  programs.ssh.extraConfig = ''
    Host *
      # carry our terminal color support (probably `TRUECOLOR`) to remote servers
      SendEnv COLORTERM
  '';

  services = {
    # needed along with the system package.
    # without this installed Flatpak apps don't show up as desktop entries for example.
    flatpak.enable = true;
  };

  environment.systemPackages = with pkgs; [
    audio-change-output
    emoji
    flatpak
    libcec
    mirror-screen
    (mpv.override {
      # youtubeSupport = false;
    })
    password
    remmina
    seedbox-mount
    tailscaleSwitch
    ungoogled-chromium # for those rare times when Firefox doesn't work for a specific site
  ];

  fonts.packages = with pkgs; [
    # most popupar fonts are in nixpkgs
    # source-code-pro
  ];
}
