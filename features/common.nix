{ pkgs, config, ... }:
{
  # add allow-items here if updating nixpkgs channel tells you that a package is insecure and if you're unable to update to newer NixOS version
  nixpkgs.config.permittedInsecurePackages = [
    # "docker-28.5.2"
  ];

  hardware = {
    enableRedistributableFirmware = true;
  };

  networking = {
    networkmanager = {
      enable = true;
    };
  };
  time.timeZone = "Europe/Helsinki";
  i18n = {
    defaultLocale = "en_US.UTF-8";

    extraLocaleSettings = {
      LC_TIME = "fi_FI.UTF-8"; # only use Finnish for date/time formatting
      LC_MESSAGES = "en_US.UTF-8";
    };

    supportedLocales = [
      "en_US.UTF-8/UTF-8"
      "fi_FI.UTF-8/UTF-8"
    ];
  };
  console.keyMap = "fi";

  services = {
    openssh = {
      enable = true;
      openFirewall = true;
      settings = {
        # PasswordAuthentication = true;
        PermitRootLogin = "no";
        # for some reason ENV vars passed by client need to be whitelisted
        AcceptEnv = [ "COLORTERM" ];
      };
    };

    tailscale = {
      enable = true;
      openFirewall = true;
    };

    # FIXME: commented out because it seems to crash for some reason
    # avahi = {
    #   enable = true;
    #   nssmdns4 = true;
    #   openFirewall = true;
    # };
  };

  users.users.joonas = {
    isNormalUser = true;
    initialPassword = "nixos";
    extraGroups = [
      "networkmanager"
      "wheel"
      "input" # moonlight to use input devices in console mode
      "video" # needed for `/dev/dri/...` access (moonlight uses direct rendering)
      "render"
    ];
    openssh.authorizedKeys.keys = [
      "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAID0kw624nwuYMt0rju0MRfkvT6Opmww52kmlnQC71LtB joonas@work"
    ];
  };

  home-manager = {
    useGlobalPkgs = true;
    users."joonas" = {
      programs = {
        git = {
          enable = true;
        };

        yazi = {
          enable = true;

          # show in status bar the currently hovered item's modification date.
          # status bar customization: https://yazi-rs.github.io/docs/tips/#symlink-in-status
          initLua = ''
            Status:children_add(function()
              local h = cx.active.current.hovered
              if not h or not h.cha.mtime then
                return ""
              end

              local mtime = math.floor(h.cha.mtime)

              return ui.Line {
                ui.Span("Modified: "):fg("cyan"),
                ui.Span(os.date("%Y-%m-%d %H:%M", mtime)):fg("blue"),
                ui.Span(" "),
              }
            end, 500, Status.RIGHT)
          '';

          settings = {
            mgr = {
              show_hidden = true;
            };
          };
        };
      };

      # You should not change this value, even if you update Home Manager. If you do
      # want to update the value, then make sure to first check the Home Manager
      # release notes.
      home.stateVersion = "25.11"; # Please read the comment before changing.
    };
  };

  environment = {
    # these are by default present in `$PATH`. you can use programs without having them as system packages but then you need
    # to use Nix to manage any config files that reference them (see use of `tuigreet` as an example)
    systemPackages = with pkgs; [
      age
      bat
      binutils
      config.boot.kernelPackages.usbip
      curl
      hexyl
      htop
      # lf
      nvme-cli
      pik
      procs
      pv
      ripgrep
      tree
      unrar-free # free version of unrar
      unzip
      usbtop
      usbutils # lsusb
      viddy
      wget
      zoxide # needs to be added as system package (referencing from `interactiveShellInit` not enough for it to fully work)
    ];

    shellAliases = {
      ".." = "cd ..";
      "..." = "cd ../..";
      "...." = "cd ../../..";
      "....." = "cd ../../../..";
      "b" = "yazi"; # = "browse"
      "e" = "\$EDITOR"; # = "edit"
    };
  };

  programs = {
    starship = {
      enable = true;
      # https://starship.rs/config/
      settings = {
        add_newline = true;
        format = "┌─── $username$hostname$localip$shlvl$singularity$kubernetes$directory$vcsh$fossil_branch$fossil_metrics$git_branch$git_commit$git_state$git_metrics$git_status$hg_branch$pijul_channel$docker_context$package$c$cmake$cobol$daml$dart$deno$dotnet$elixir$elm$erlang$fennel$gleam$golang$guix_shell$haskell$haxe$helm$java$julia$kotlin$gradle$lua$nim$nodejs$ocaml$opa$perl$php$pulumi$purescript$python$quarto$raku$rlang$red$ruby$rust$scala$solidity$swift$terraform$typst$vlang$vagrant$zig$buf$nix_shell$conda$meson$spack$memory_usage$aws$gcloud$openstack$azure$nats$direnv$env_var$crystal$custom$sudo$cmd_duration$line_break$jobs$battery$time$status$os$container$shell└─$character";
        character = {
          success_symbol = "[❯](bold green)";
          error_symbol = "[✗](bold red) ";
        };
      };
    };

    bash.interactiveShellInit = ''
      eval "$(${pkgs.starship}/bin/starship init bash)"
      eval "$(${pkgs.zoxide}/bin/zoxide init bash)"
      # eval "$(${pkgs.atuin}/bin/atuin init bash)"
    '';
  };

  # don't ask for passwords. this is also needed by nixos-rebuild switch to not need asking sudo passwords.
  security.sudo.wheelNeedsPassword = false;

  nix.settings = {
    trusted-users = [
      "root"
      "@wheel" # otherwise nixos-rebuild switch gives `cannot add path ... because it lacks a signature by a trusted key` errors
    ];
    # nix requires all users to enable experimental features to use Nix as the way that is idiomatic......
    experimental-features = [
      "nix-command"
      "flakes"
    ];
  };

  system.stateVersion = "25.11";
}
