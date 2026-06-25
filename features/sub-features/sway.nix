{
  pkgs,
  lib,
  ...
}:

let
  mod = "Mod4";
  alt = "Mod1";
  wallpaper = pkgs.callPackage ./wallpaper.nix { };
  dim-screen = pkgs.callPackage ../../packages/dim-screen.nix { };
  myCustomFonts = pkgs.stdenvNoCC.mkDerivation {
    pname = "my-custom-fonts";
    version = "1.0";

    src = ../../config/fonts;

    installPhase = ''
      mkdir -p $out/share/fonts/truetype
      cp *.ttf $out/share/fonts/truetype/
    '';
  };
in
{
  services = {
    greetd = {
      enable = true;
      settings = {
        default_session = {
          command = "${pkgs.tuigreet}/bin/tuigreet --remember --cmd=sway --theme=\"border=magenta;text=cyan;prompt=green;time=red;action=blue;button=yellow;container=black;input=red\"";
          user = "greeter";
        };
      };
    };
  };

  home-manager = {
    users."joonas" = {
      services.swayidle =
        let
          screenOffTimeout = 7 * 60; # [s]
          dimScreenBeforeOff = 30; # [s]
          # Lock command
          # lock = "${pkgs.swaylock}/bin/swaylock --daemonize";
          # TODO: modify "display" function based on your window manager
          # Sway
          display = status: "${pkgs.sway}/bin/swaymsg 'output * power ${status}'";
          # Hyprland
          # display = status: "hyprctl dispatch dpms ${status}";
          # Niri
          # display = status: "${pkgs.niri}/bin/niri msg action power-${status}-monitors";
        in
        {
          enable = true;
          timeouts = [
            # {
            #   timeout = 20;
            #   command = lock;
            # }
            {
              timeout = screenOffTimeout - dimScreenBeforeOff; # in seconds
              command = "${dim-screen}/bin/dim --alpha=1.0 --fade=${toString (dimScreenBeforeOff)} --duration=${toString (dimScreenBeforeOff)}";
            }
            {
              timeout = screenOffTimeout;
              command = display "off";
              resumeCommand = display "on";
            }
            # {
            #   timeout = 30;
            #   command = "${pkgs.systemd}/bin/systemctl suspend";
            # }
          ];
          events = {
            # multiple commands because adding duplicated entries for the same event may not work
            # before-sleep = (display "off") + "; " + lock;
            after-resume = display "on";
            # lock = (display "off") + "; " + lock;
            unlock = display "on";
          };
        };

      wayland.windowManager.sway = {
        enable = true;

        # Remove these two lines if you want Home Manager to install Sway itself.
        package = null;
        # checkConfig = false;

        extraConfigEarly = ''
          # necessary NixOS integration surface
          include /etc/sway/config.d/nixos.conf

          set $alt ${alt}
          set $mod ${mod}

          # `GTK_IM_MODULE` to overcome dead keys (like tilde) not working on GTK: https://github.com/ghostty-org/ghostty/discussions/8899
          set $term env GTK_IM_MODULE=simple ghostty

          seat * hide_cursor 3000
          seat * hide_cursor when-typing enable
        '';

        config = {
          modifier = mod;
          # don't know why we've to specify the terminal if this isn't anyway usable as variable (without introducing explicit variable)
          terminal = "$term";
          menu = "fuzzel";

          input = {
            "*" = {
              xkb_layout = "fi";
            };
            "type:touchpad" = {
              tap = "enabled";
            };
          };

          output = {
            "*" = {
              bg = "${wallpaper}/share/backgrounds/wallpaper.jpg fill";
            };
          };

          workspaceAutoBackAndForth = true;
          focus.newWindow = "smart";

          fonts = {
            names = [ "monospace" ];
            size = 8.0;
          };

          floating.modifier = "Ctrl";

          window = {
            titlebar = false;
            border = 0;
            commands = [
              {
                criteria = {
                  window_role = "pop-up";
                };
                command = "floating enable, border normal";
              }
              {
                criteria = {
                  window_role = "task_dialog";
                };
                command = "floating enable, border normal";
              }
              {
                criteria = {
                  class = "Qemu-system-x86_64";
                };
                command = "floating enable";
              }
              {
                criteria = {
                  class = "Qemu-system-i386";
                };
                command = "floating enable";
              }
              {
                criteria = {
                  class = "vlc";
                };
                command = "floating enable";
              }
              {
                criteria = {
                  app_id = "showmethekey-gtk";
                  title = "Floating Window - Show Me The Key";
                };
                command = "floating enable, sticky enable";
              }
              {
                criteria = {
                  instance = ".*_9dc82_dialog";
                };
                command = "floating enable";
              }
              {
                criteria = {
                  class = ".*_9dc82_dialog";
                };
                command = "floating enable";
              }
              {
                criteria = {
                  app_id = "at.yrlf.wl_mirror";
                };
                command = "floating enable";
              }
            ];
          };

          keybindings = lib.mkForce {
            "$mod+Return" = "exec $term";
            "$alt+F4" = "kill";
            "$mod+d" = ''exec "fuzzel"'';
            "$mod+l" = "exec swaylock --image=${wallpaper}/share/backgrounds/wallpaper.jpg";

            "$mod+Left" = "focus left";
            "$mod+Down" = "focus down";
            "$mod+Up" = "focus up";
            "$mod+Right" = "focus right";

            "$mod+Shift+j" = "move left";
            "$mod+Shift+k" = "move down";
            "$mod+Shift+l" = "move up";
            "$mod+Shift+semicolon" = "move right";
            "$mod+Shift+Left" = "move left";
            "$mod+Shift+Down" = "move down";
            "$mod+Shift+Up" = "move up";
            "$mod+Shift+Right" = "move right";

            "$mod+h" = "split h";
            "$mod+v" = "split v";
            "$mod+f" = "fullscreen toggle";
            "$mod+s" = "layout stacking";
            "$mod+w" = "layout tabbed";
            "$mod+e" = "layout toggle split";
            "$mod+Shift+space" = "floating toggle";
            "$mod+a" = "focus parent";

            "$mod+F1" = "exec www-browser --new-window";
            "$mod+F2" = "exec thunar";
            "$mod+F3" = "exec /usr/bin/password.sh";

            "$mod+n" = "exec i3-input -F 'rename workspace to \"%s\"' -P 'New name: '";

            "$mod+1" = "workspace number 1";
            "$mod+2" = "workspace number 2";
            "$mod+3" = "workspace number 3";
            "$mod+4" = "workspace number 4";
            "$mod+5" = "workspace number 5";
            "$mod+6" = "workspace number 6";
            "$mod+7" = "workspace number 7";
            "$mod+8" = "workspace number 8";
            "$mod+9" = "workspace number 9";
            "$mod+0" = "workspace number 10";

            "$mod+Shift+1" = "move container to workspace number 1";
            "$mod+Shift+2" = "move container to workspace number 2";
            "$mod+Shift+3" = "move container to workspace number 3";
            "$mod+Shift+4" = "move container to workspace number 4";
            "$mod+Shift+5" = "move container to workspace number 5";
            "$mod+Shift+6" = "move container to workspace number 6";
            "$mod+Shift+7" = "move container to workspace number 7";
            "$mod+Shift+8" = "move container to workspace number 8";
            "$mod+Shift+9" = "move container to workspace number 9";
            "$mod+Shift+0" = "move container to workspace number 10";

            "$mod+Shift+c" = "reload";
            "$mod+Shift+r" = "restart";
            "$mod+Shift+e" = "exec xfce4-session-logout";
            "$mod+r" = ''mode "resize"'';

            "XF86AudioRaiseVolume" = "exec pactl set-sink-volume @DEFAULT_SINK@ +5%";
            "XF86AudioLowerVolume" = "exec pactl set-sink-volume @DEFAULT_SINK@ -5%";
            # "XF86AudioRaiseVolume" = "exec --no-startup-id hautomo-client audioctl vol-increase -- 5";
            # "XF86AudioLowerVolume" = "exec --no-startup-id hautomo-client audioctl vol-increase -- -5";
            "XF86AudioMute" = "exec --no-startup-id amixer -D pulse set Master 1+ toggle";
            "Shift+XF86AudioMute" =
              "exec --no-startup-id hautomo-client audioctl toggle-between-speakers-and-headphones";
            "XF86AudioPlay" = "exec ${pkgs.playerctl}/bin/playerctl play-pause";
            # "XF86AudioPause" = "exec --no-startup-id hautomo-client playerctl play-pause";
            "XF86AudioNext" = "exec --no-startup-id hautomo-client playerctl next";
            "XF86AudioPrev" = "exec --no-startup-id hautomo-client playerctl previous";

            "XF86HomePage" = "exec www-browser --new-window";
            "Shift+XF86HomePage" = "exec www-browser --private-window";
            "XF86Mail" = "exec thunar";
            "XF86Search" = "exec thunar";
            "XF86Explorer" = "exec thunar";
            "XF86Calculator" = "exec /usr/bin/password.sh";
            "XF86MonBrightnessUp" = "exec brightnessctl set 10%+";
            "XF86MonBrightnessDown" = "exec brightnessctl set 10%-";

            "$mod+space" = "exec jsys backlight keyboard cycle";
            "Print" = "exec ${pkgs.flameshot}/bin/flameshot gui";
            # Kooha otherwise looked good BUT it only supports webm & gif
            "$mod+Print" =
              "exec ${pkgs.wf-recorder}/bin/wf-recorder --filename=record.mp4 --geometry=\"\$(${pkgs.slurp}/bin/slurp)\"";
          };

          modes = {
            resize = {
              "Left" = "resize shrink width 10 px or 10 ppt";
              "Down" = "resize grow height 10 px or 10 ppt";
              "Up" = "resize shrink height 10 px or 10 ppt";
              "Right" = "resize grow width 10 px or 10 ppt";
              "j" = "resize shrink width 10 px or 10 ppt";
              "k" = "resize grow height 10 px or 10 ppt";
              "l" = "resize shrink height 10 px or 10 ppt";
              "semicolon" = "resize grow width 10 px or 10 ppt";
              "Return" = ''mode "default"'';
              "Escape" = ''mode "default"'';
            };
          };

          bars = [
            {
              command = "${pkgs.waybar}/bin/waybar";
              # fonts = {
              #   names = [ "monospace" ];
              #   size = 11.0;
              # };
              # colors = {
              #   background = "#00000060";
              #   activeWorkspace = {
              #     border = "#333333";
              #     background = "#00000060";
              #     text = "#ffffff";
              #   };
              #   focusedWorkspace = {
              #     border = "#00000000";
              #     background = "#00000060";
              #     text = "#ffffffff";
              #   };
              #   inactiveWorkspace = {
              #     border = "#00000060";
              #     background = "#00000060";
              #     text = "#ffffff20";
              #   };
              # };
              # position = "top";
              # statusCommand = "i3status -c ${config.xdg.configHome}/i3status/config";
              # extraConfig = ''
              #   separator_symbol " | "
              # '';
            }
          ];

          # startup = [
          #   {
          #     command = ''setxkbmap -option "nbsp:none"'';
          #   }
          # ];

          gaps = {
            inner = 20;
          };
        };

        extraConfig = ''
          bindcode --release 133 exec "fuzzel"
        '';
      };
      programs = {
        # nicer than swaybar
        waybar = {
          enable = true;

          settings.main = {
            spacing = 4;

            modules-left = [ "sway/workspaces" ];
            modules-center = [ "sway/window" ];
            modules-right = [
              "privacy"
              "battery"
              "network"
              "disk"
              "memory"
              "cpu"
              "temperature" # CPU temp
              "clock"
              "tray"
            ];

            # https://github.com/Alexays/Waybar/wiki/Module:-Privacy
            privacy = {
              icon-spacing = 4;
              icon-size = 18;
              transition-duration = 1000;

              modules = [
                {
                  type = "screenshare";
                  tooltip = true;
                  tooltip-icon-size = 24;
                }
                # audio being played is not very privacy invasive
                # {
                #   type = "audio-out";
                #   tooltip = true;
                #   tooltip-icon-size = 24;
                # }
                {
                  type = "audio-in";
                  tooltip = true;
                  tooltip-icon-size = 24;
                }
              ];

              ignore-monitor = true;

              ignore = [
                {
                  type = "audio-in";
                  name = "cava";
                }
                {
                  type = "screenshare";
                  name = "obs";
                }
              ];
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Battery
            battery = {
              format = "🔋 {capacity} %";
              # don't display anything when plugged and fully charged
              format-plugged = "";
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Clock
            clock = {
              # the zero before the format string seems to be required: https://github.com/Alexays/Waybar/discussions/2538#discussioncomment-9209399
              format = "󰸗  {0:%d.%m} 󰅐  {0:%H:%M}";
              tooltip-format = "<big>{:%Y %B}</big>\n<tt><small>{calendar}</small></tt>";
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-CPU
            cpu = {
              format = "󰍛 {usage:02d}%"; # pad to two digits so width of the content doesn't bounce around
              interval = 5;
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Disk
            disk = {
              format = "󰋊 {specific_free:0.2f} GB";
              unit = "GB";
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Memory
            memory = {
              format = "󰑭  {avail:0.1f} GB";
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Network
            network = {
              format-wifi = "{icon} {essid} ({signalStrength}%)";
              format-ethernet = " {bandwidthTotalBits}";
              tooltip-format = "{ifname} via {gwaddr} 󰊗";
              format-linked = "{ifname} (No IP) 󰊗";
              format-disconnected = "Disconnected ⚠";
              format-alt = "{ifname}: {ipaddr}/{cidr}";
              format-icons = [
                "󰤯"
                "󰤟"
                "󰤢"
                "󰤥"
                "󰤨"
              ];
              interval = 5; # [s]
            };

            # https://github.com/Alexays/Waybar/wiki/Module:-Temperature
            temperature = {
              hwmon-path = "/dev/cpu_temp"; # symlink set up by udev rule
              format = ""; # hiden when good
              format-critical = "{temperatureC} °C";
              critical-threshold = 60;
            };
          };

          style = ''
            * {
                /* `otf-font-awesome` is required to be installed for icons */
                font-family: FontAwesome, Roboto, Helvetica, Arial, sans-serif;
                font-size: 13px;
            }

            window#waybar {
                background-color: rgba(43, 48, 59, 0.5);
                border-bottom: 3px solid rgba(100, 114, 125, 0.5);
                color: #ffffff;
                transition-property: background-color;
                transition-duration: .5s;
            }

            window#waybar.hidden {
                opacity: 0.2;
            }

            window#waybar.termite {
                background-color: #3F3F3F;
            }

            window#waybar.chromium {
                background-color: #000000;
                border: none;
            }

            button {
                /* Use box-shadow instead of border so the text isn't offset */
                box-shadow: inset 0 -3px transparent;
                /* Avoid rounded borders under each button name */
                border: none;
                border-radius: 0;
            }

            /* https://github.com/Alexays/Waybar/wiki/FAQ#the-workspace-buttons-have-a-strange-hover-effect */
            button:hover {
                background: inherit;
                box-shadow: inset 0 -3px #ffffff;
            }

            /* you can set a style on hover for any module like this */
            #pulseaudio:hover {
                background-color: #a37800;
            }

            #workspaces button {
                padding: 0 5px;
                background-color: transparent;
                color: #ffffff;
            }

            #workspaces button:hover {
                background: rgba(0, 0, 0, 0.2);
            }

            #workspaces button.focused, #workspaces button.active {
                background-color: #64727D;
                box-shadow: inset 0 -3px #ffffff;
            }

            #workspaces button.urgent {
                background-color: #eb4d4b;
            }

            #mode {
                background-color: #64727D;
                box-shadow: inset 0 -3px #ffffff;
            }

            #clock,
            #battery,
            #cpu,
            #memory,
            #disk,
            #temperature,
            #backlight,
            #network,
            #pulseaudio,
            #wireplumber,
            #custom-media,
            #tray,
            #mode,
            #idle_inhibitor,
            #scratchpad,
            #power-profiles-daemon,
            #mpd {
                padding: 0 10px;
                color: #ffffff;
                background-color: rgba(0, 0, 0, 0.25);
            }

            #window,
            #workspaces {
                margin: 0 4px;
            }

            /* If workspaces is the leftmost module, omit left margin */
            .modules-left > widget:first-child > #workspaces {
                margin-left: 0;
            }

            /* If workspaces is the rightmost module, omit right margin */
            .modules-right > widget:last-child > #workspaces {
                margin-right: 0;
            }

            @keyframes blink {
                to {
                    background-color: #ffffff;
                    color: #000000;
                }
            }

            /* Using steps() instead of linear as a timing function to limit cpu usage */
            #battery.critical:not(.charging) {
                background-color: #f53c3c;
                color: #ffffff;
                animation-name: blink;
                animation-duration: 0.5s;
                animation-timing-function: steps(12);
                animation-iteration-count: infinite;
                animation-direction: alternate;
            }

            #power-profiles-daemon {
                padding-right: 15px;
            }

            #power-profiles-daemon.performance {
                background-color: #f53c3c;
                color: #ffffff;
            }

            #power-profiles-daemon.balanced {
                background-color: #2980b9;
                color: #ffffff;
            }

            #power-profiles-daemon.power-saver {
                background-color: #2ecc71;
                color: #000000;
            }

            label:focus {
                background-color: #000000;
            }

            #scratchpad {
                background: rgba(0, 0, 0, 0.2);
            }

            #scratchpad.empty {
                background-color: transparent;
            }

            #privacy {
                padding: 0;
            }

            #privacy-item {
                padding: 0 5px;
                color: white;
            }

            #privacy-item.screenshare {
                background-color: #cf5700;
            }

            #privacy-item.audio-in {
                background-color: #1ca000;
            }

            #privacy-item.audio-out {
                background-color: #0069d4;
            }
          '';
        };
      };
    };
  };

  programs = {
    # waybar.enable = true;
    sway = {
      enable = true;
      wrapperFeatures.gtk = true;
    };
  };

  environment.systemPackages = with pkgs; [
    fuzzel
    libnotify # notify-send
    mako # notification daemon (will run automatically when a notification is emitted. This happens via D-Bus activation)
    showmethekey
    wev
    wl-clipboard
  ];

  fonts.packages = with pkgs; [
    font-awesome_4 # required by waybar
    myCustomFonts
  ];
}
