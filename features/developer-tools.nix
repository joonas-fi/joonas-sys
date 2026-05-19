{ lib, pkgs, ... }:

let
  inherit (lib) concatMapStringsSep escapeShellArg optionalString;

  shellArgs = args: concatMapStringsSep " " escapeShellArg args;
  flattenMap = f: xs: builtins.concatLists (map f xs);

  appendStaticArgs =
    args: concatMapStringsSep "\n" (arg: "docker_args+=(${escapeShellArg arg})") args;

  appendEnv =
    env:
    concatMapStringsSep "\n" (
      name:
      let
        value = toString env.${name};
      in
      "docker_args+=(--env ${escapeShellArg "${name}=${value}"})"
    ) (builtins.attrNames env);

  appendPassEnv =
    names: concatMapStringsSep "\n" (name: "docker_args+=(--env ${escapeShellArg name})") names;

  mkContainerTool =
    {
      name,
      image,
      network,
      command ? [ ],
      appendUserArgs ? true,
      mountCurrentDir ? false,
      cwdMode ? "rw",
      containerCwd ? "/workspace",
      workdir ? null,
      entrypoint ? null,
      runAsCurrentUser ? true,
      readOnlyRoot ? true,
      tmpfs ? [ "/tmp:rw,nosuid,nodev,size=256m" ],
      privileged ? false,
      pidHost ? false,
      hostRootReadOnly ? false,
      hostRootWorkdirFromPwd ? false,
      env ? { },
      passEnv ? [ ],
      volumes ? [ ],
      devices ? [ ],
      extraDockerArgs ? [ ],
      extraDockerArgSetup ? "",
      beforeDocker ? "",
      afterDocker ? "",
    }:
    let
      cwdSuffix = optionalString (cwdMode == "ro") ":ro";
      commandText = optionalString (command != [ ]) " ${shellArgs command}";
      userArgsText = optionalString appendUserArgs ''"$@"'';
    in
    pkgs.writeShellApplication {
      inherit name;
      runtimeInputs = [
        pkgs.coreutils
        pkgs.docker
      ];
      text = ''
        # _pwd_symlinksresolved() {
        #   readlink --canonicalize-existing .
        # }

        docker_args=(
          run
          --rm
          --network=${escapeShellArg network}
        )

        if [ -t 0 ] && [ -t 1 ]; then
          docker_args+=(-it)
        else
          docker_args+=(-i)
        fi

        ${optionalString (!privileged) ''
          docker_args+=(--cap-drop ALL)
          docker_args+=(--security-opt no-new-privileges)
        ''}

        ${optionalString privileged ''
          docker_args+=(--privileged)
        ''}

        ${optionalString pidHost ''
          docker_args+=(--pid host)
        ''}

        ${optionalString readOnlyRoot ''
          docker_args+=(--read-only)
        ''}

        ${appendStaticArgs (
          flattenMap (mount: [
            "--tmpfs"
            mount
          ]) tmpfs
        )}

        ${optionalString runAsCurrentUser ''
          docker_args+=(--user "$(id -u):$(id -g)")
          docker_args+=(--env HOME=/tmp)
        ''}

        ${appendEnv env}
        ${appendPassEnv passEnv}
        ${appendStaticArgs (
          flattenMap (volume: [
            "-v"
            volume
          ]) volumes
        )}
        ${appendStaticArgs (
          flattenMap (device: [
            "--device"
            device
          ]) devices
        )}
        ${appendStaticArgs extraDockerArgs}

        ${optionalString mountCurrentDir ''
          docker_args+=(-v "$PWD:${containerCwd}${cwdSuffix}")
          docker_args+=(--workdir ${escapeShellArg containerCwd})
        ''}

        ${optionalString hostRootReadOnly ''
          docker_args+=(-v "/:/host:ro")
        ''}

        ${optionalString hostRootWorkdirFromPwd ''
          docker_args+=(--workdir "/host$(_pwd_symlinksresolved)")
        ''}

        ${optionalString (workdir != null) ''
          docker_args+=(--workdir ${escapeShellArg workdir})
        ''}

        ${optionalString (entrypoint != null) ''
          docker_args+=(--entrypoint ${escapeShellArg entrypoint})
        ''}

        ${extraDockerArgSetup}
        ${beforeDocker}

        set +e
        docker "''${docker_args[@]}" ${escapeShellArg image} ${commandText} ${userArgsText}
        status=$?
        set -e

        ${afterDocker}

        exit "$status"
      '';
    };

  mkTool =
    args:
    mkContainerTool (
      args
      // {
        network = "none";
      }
    );

  mkNetworkTool =
    {
      network ? "bridge",
      ...
    }@args:
    mkContainerTool (
      args
      // {
        inherit network;
      }
    );
  turbobob = pkgs.callPackage ../packages/turbobob.nix { };
in
{
  # NOTE: virtualization is spelled wrong 😬
  virtualisation = {
    docker.enable = true;
    containerd.enable = true;
  };

  programs = {
    ssh = {
      startAgent = true;
    };
  };

  home-manager = {
    users."joonas" =
      { config, ... }:
      {
        home.file.".ssh/id_rsa".source = config.lib.file.mkOutOfStoreSymlink "/sto/id/eaPAN6iqlMM/id_rsa";

        programs = {
          lazygit = {
            enable = true;
            settings = {
              disableStartupPopups = true;
              notARepository = "quit"; # do not offer to do `$ git init` or select from recent repos
              gui = {
                useHunkModeInStagingView = false;
              };
              # https://github.com/jesseduffield/lazygit/blob/master/docs/Config.md
              keybinding = {
                universal = {
                  # https://github.com/jesseduffield/lazygit/blob/40d6800fd35ad2d5c6d96d6f08ffa42d4c764ad3/pkg/i18n/english.go#L1838
                  toggleRangeSelect = "V";
                };
                commits = {
                  cherryPickCopy = "c";
                  pasteCommits = "v";
                  # this was taken over by "c" so our main purpose is to set it to something else. lazygit doesn't warn about conflicting keybinds.... 🫠
                  setFixupMessage = "C";
                };
              };
            };
          };

          git = {
            enable = true;

            settings = {
              user = {
                name = "Joonas Loppi";
                email = "joonas+code@joonas.fi";
              };

              init.defaultBranch = "main";

              pull.rebase = true;

              # push.autoSetupRemote = true;
            };
          };
        };
      };
  };

  environment = {
    shellAliases = {
      "alpine" =
        "docker run --rm -it --network=host -v \"$(pwd):/workspace\" --workdir=/workspace alpine \"$@\"";
      "j" = "just --choose --unsorted";
      "lg" = "lazygit";
      "pubkey" = "ssh-keygen -y -f ~/.ssh/id_rsa";
    };

    variables = {
      FZF_DEFAULT_OPTS = "--no-mouse --layout=reverse";
    };

    systemPackages = with pkgs; [
      # lowcharts
      android-tools # adb
      bpftrace
      fzf
      gdb
      gum
      jnv
      jq
      just
      mkcert
      oras
      qemu_kvm # qemu-host-cpu-only
      rclone
      tcpdump
      tealdeer
      turbobob
      up
      wireguard-tools

      (mkTool {
        name = "exiftool";
        image = "ghcr.io/r-xs-fi/exiftool:latest";
        mountCurrentDir = true;
      })

      (mkTool {
        name = "convert";
        image = "ghcr.io/r-xs-fi/imagemagick:latest";
        command = [ "convert" ];
        mountCurrentDir = true;
      })

      (mkNetworkTool {
        name = "croc";
        image = "ghcr.io/r-xs-fi/croc:latest";
        mountCurrentDir = true;
      })

      (mkNetworkTool {
        name = "nslookup";
        image = "alpine:latest";
        command = [ "nslookup" ];
      })

      (mkNetworkTool {
        name = "dig";
        image = "ghcr.io/r-xs-fi/dig:latest";
      })

      (mkTool {
        name = "htop";
        image = "joonas/htop:latest";
        pidHost = true;
        extraDockerArgSetup = ''
          docker_args+=(-v "$HOME/.config/htop/htoprc:/etc/htoprc:ro")
          docker_args+=(-v "/etc/passwd:/etc/passwd:ro")
        '';
      })

      (mkTool {
        name = "iotop";
        image = "joonas/iotop:latest";
        privileged = true;
        pidHost = true;
        runAsCurrentUser = false;
      })

      (mkNetworkTool {
        name = "nethogs";
        image = "joonas/nethogs:latest";
        network = "host";
        privileged = true;
        pidHost = true;
        runAsCurrentUser = false;
      })

      (mkNetworkTool {
        name = "ethtool";
        image = "ghcr.io/r-xs-fi/ethtool:latest";
        network = "host";
      })

      (mkTool {
        name = "pstree";
        image = "joonas/psmisc:latest";
        command = [ "pstree" ];
        pidHost = true;
      })

      (mkTool {
        name = "killall";
        image = "joonas/psmisc:latest";
        command = [ "killall" ];
        privileged = true;
        pidHost = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "lspci";
        image = "ghcr.io/r-xs-fi/lspci:latest";
      })

      (mkTool {
        name = "fdisk";
        image = "joonas/fdisk:latest";
        privileged = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "gdisk";
        image = "joonas/fdisk:latest";
        entrypoint = "/usr/bin/gdisk";
        privileged = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "hollywood";
        image = "bcbcarl/hollywood:latest";
      })

      (mkTool {
        name = "asciiquarium";
        image = "danielkraic/asciiquarium:latest";
      })

      (mkTool {
        name = "radeontop";
        image = "joonas/radeontop:latest";
        privileged = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "pwdgen";
        image = "joonas/pwdgen:20211113_1928_0fea588c";
      })

      (mkNetworkTool {
        name = "nmap";
        image = "joonas/nmap:latest";
        network = "host";
      })

      (mkTool {
        name = "stress-ng";
        image = "joonas/stress-ng:latest";
      })

      (mkNetworkTool {
        name = "netstat";
        image = "alpine:latest";
        command = [ "netstat" ];
        network = "host";
      })

      (mkNetworkTool {
        name = "arp";
        image = "alpine:latest";
        command = [ "arp" ];
        network = "host";
      })

      (mkNetworkTool {
        name = "iperf";
        image = "joonas/iperf:latest";
        network = "host";
      })

      (mkTool {
        name = "file";
        image = "joonas/file:latest";
        hostRootReadOnly = true;
        hostRootWorkdirFromPwd = true;
      })

      (mkTool {
        name = "rpm2cpio";
        image = "joonas/rpm2cpio:latest";
        hostRootReadOnly = true;
        hostRootWorkdirFromPwd = true;
      })

      (mkTool {
        name = "smartctl";
        image = "joonas/smartmontools:latest";
        privileged = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "qrencode";
        image = "joonas/qrencode:latest";
      })

      (mkTool {
        name = "figlet";
        image = "ghcr.io/r-xs-fi/figlet:latest";
      })

      (mkTool {
        name = "ncdu";
        image = "ghcr.io/r-xs-fi/ncdu:latest";
        passEnv = [ "LANG" ];
        extraDockerArgSetup = ''
          if [ "$#" -eq 0 ]; then
            docker_args+=(-v "$PWD:/workspace" --workdir /workspace)
          else
            docker_args+=(-v "/:/host:ro")
          fi
        '';
      })

      (mkNetworkTool {
        name = "awscli";
        image = "amazon/aws-cli:latest";
        entrypoint = "";
        command = [ "bash" ];
        appendUserArgs = false;
        readOnlyRoot = false;
        workdir = "/aws";
        extraDockerArgSetup = ''
          docker_args+=(-v "$PWD:/aws")
        '';
      })

      (mkNetworkTool {
        name = "hey";
        image = "joonas/hey:latest";
        network = "host";
      })

      (mkTool {
        name = "lshw";
        image = "joonas/lshw:latest";
        privileged = true;
        runAsCurrentUser = false;
      })

      (mkTool {
        name = "sqlite";
        image = "joonas/sqlite:latest";
        mountCurrentDir = true;
      })

      (mkTool {
        name = "pdfcpu";
        image = "ghcr.io/r-xs-fi/pdfcpu:latest";
        mountCurrentDir = true;
        env = {
          HOME = "/workspace";
        };
      })

      (mkTool {
        name = "uuidgen";
        image = "joonas/uuidgen:latest";
      })

      (mkTool {
        name = "cmatrix";
        image = "ghcr.io/r-xs-fi/cmatrix:latest";
      })

      (mkNetworkTool {
        name = "telnet";
        image = "joonas/telnet:latest";
      })

      (mkNetworkTool {
        name = "whois";
        image = "ghcr.io/r-xs-fi/whois:latest";
      })

      (mkNetworkTool {
        name = "mapscii";
        image = "joonas/telnet:latest";
        command = [ "mapscii.me" ];
        appendUserArgs = false;
      })

      (mkTool {
        name = "cowsay";
        image = "ghcr.io/r-xs-fi/cowsay:latest";
      })

      (mkTool {
        name = "doge";
        image = "ghcr.io/r-xs-fi/doge:latest";
      })

      (mkTool {
        name = "tokei";
        image = "ghcr.io/r-xs-fi/tokei:latest";
        mountCurrentDir = true;
      })

      (mkTool {
        name = "neofetch";
        image = "joonas/neofetch:latest";
        volumes = [ "/etc/os-release:/etc/os-release:ro" ];
      })

      (mkTool {
        name = "ffmpeg";
        image = "ghcr.io/r-xs-fi/ffmpeg:latest";
        mountCurrentDir = true;
      })

      (mkTool {
        name = "yulelog";
        image = "joonas/yulelog:latest";
      })

      (mkTool {
        name = "pipes.sh";
        image = "joonas/pipes.sh:latest";
      })

      (mkTool {
        name = "go-life";
        image = "ghcr.io/r-xs-fi/go-life:latest";
      })

      # The original wrapper used /usr/bin/shelldon. This assumes you provide a matching image.
      (mkNetworkTool {
        name = "shelldon";
        image = "ghcr.io/r-xs-fi/shelldon:latest";
        readOnlyRoot = false;
        extraDockerArgs = [
          "--env-file"
          "/sto/id/I5CzPYOBxNs/openai.env"
        ];
        command = [
          "sh"
          "-c"
          ''
            if [ "$#" -eq 0 ]; then
              verb=$(gum choose -- Exec Ask --help)
              if [ "$verb" = "Exec" ]; then
                prompt=$(gum input --placeholder='Prompt')
                exec shelldon exec "$prompt"
              elif [ "$verb" = "--help" ]; then
                exec shelldon --help
              else
                prompt=$(gum input --placeholder='Prompt')
                exec shelldon ask "$prompt"
              fi
            fi

            exec shelldon "$@"
          ''
          "shelldon"
        ];
      })

      (mkTool {
        name = "lolcat";
        image = "joonas/lolcat:latest";
        command = [ "--force" ];
      })

      (mkTool {
        name = "fortune";
        image = "ghcr.io/r-xs-fi/fortune:latest";
      })

      (mkTool {
        name = "scooter";
        image = "ghcr.io/r-xs-fi/scooter:latest";
        mountCurrentDir = true;
      })

      (mkNetworkTool {
        name = "codex";
        image = "ghcr.io/r-xs-fi/codex:latest";
        entrypoint = "";
        command = [
          "sh"
          "-c"
          "umask 017; exec codex \"$@\""
          "codex"
        ];
        mountCurrentDir = true;
        readOnlyRoot = false;
        runAsCurrentUser = false;
        env = {
          HOME = "/root";
        };
        extraDockerArgSetup = ''
          docker_args+=(-v "/persist/apps/codex:/root/.codex")
          docker_args+=(--user "0:$(id -g)")
        '';
      })

      (mkNetworkTool {
        name = "torrent";
        image = "ghcr.io/r-xs-fi/torrent:latest";
        command = [ "download" ];
        network = "host";
        privileged = true;
        mountCurrentDir = true;
        runAsCurrentUser = false;
      })

      (mkNetworkTool {
        name = "onvif";
        image = "ghcr.io/r-xs-fi/onvif-cli:latest";
        network = "host";
      })
    ];
  };
}
