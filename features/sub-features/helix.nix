{ pkgs, config, ... }:
{
  home-manager = {
    users."joonas" = {
      programs = {
        helix = {
          enable = true;
          # TODO: doesn't work
          # defaultEditor=true;
          settings = {
            theme = "zed_onedark";
            editor = {
              rulers = [ 120 ];
              mouse = false;
              # end-of-line-diagnostics = "hint";
              bufferline = "always";

              inline-diagnostics.cursor-line = "warning"; # show warnings and errors on the cursorline inline

              file-picker = {
                # show hidden "dotfiles" (the logic is pretty counter-intuitive as `hidden = false` means to show hidden files)
                hidden = false;
                # git-ignore = false;
                # e.g. `joonas-sys` has symlinks pointing to outside of the repo to current system, hence Helix would scan pretty much the entire system
                follow-symlinks = false;
              };
            };

            # https://docs.helix-editor.com/keymap.html#normal-mode
            keys.normal = {
              # familiarity with Vim
              ins = "insert_mode";
              # by default "r" is replace with a character and "R" is replace with yanked text
              # swap those around.
              r = "replace_with_yanked";
              R = "replace";
              # same swap for change (don't yank by default)
              c = "change_selection_noyank";
              "A-c" = "change_selection";
            };
          };

          # for list of languages:
          #   https://docs.helix-editor.com/lang-support.html
          #
          # if need to debug Helix's LSP problems:
          # 1. run Helix with `-v` flag
          # 2. tail -f ~/.cache/helix/helix.log
          #
          # NOTE: relies on the fact that Helix sets langserver's workdir to project root for Turbo Bob to resolve
          # correct project and from there which container to forward the langserver invokation to.
          languages = {
            language-server = {
              turbobob-lsp = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                ];
              };

              turbobob-lang-from-project-tsx = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-from-project=tsx"
                ];

                # `--` trick to outsmart typescript-language-server arg validation but enable deep down in dependencies `--clientProcessId=1` detection from CLI args
                # to fool the biblically stupid parent PID watchdog which is designed to cripple container support for no good reason.
                # https://github.com/microsoft/vscode-languageserver-node/blob/d58c00bbf8837b9fd0144924db5e7b1c543d839e/server/src/node/main.ts#L43-L74
              };

              turbobob-generic-json = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=json"
                ];

                # doesn't seem to work :(
                provideFormatter = true;
              };

              turbobob-generic-html = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=html"
                ];
              };

              turbobob-generic-css = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=css"
                ];
              };

              turbobob-generic-bash = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=bash"
                ];
              };

              turbobob-generic-yaml = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=yaml"
                ];
              };

              marksman = {
                # NOTE: rumdl might be better (with `enableSymbols=true`) once version gets updated to nixpkgs
                command = "${pkgs.marksman}/bin/marksman";
                args = [ "server" ];
                config = {
                  enableSymbols = true;
                };
                # command = "bob";
                # args = [
                #   "tools"
                #   "langserver"
                #   "--lang-generic=md"
                # ];
              };

              turbobob-generic-dockerfile = {
                command = "bob";
                args = [
                  "tools"
                  "langserver"
                  "--lang-generic=dockerfile"
                ];
              };

              nil = {
                command = "docker";
                args = [
                  "exec"
                  "--interactive"
                  "--workdir=/workspace"
                  "nix"
                  "nil"
                  "--stdio"
                ];
              };
            };

            language = [
              {
                name = "go";
                "language-servers" = [ "turbobob-lsp" ];

                # https://github.com/go-delve/delve/blob/master/Documentation/api/dap/README.md
                # https://github.com/golang/vscode-go/blob/master/docs/debugging.md
                # Helix supported DAP transports: tcp/stdio
                #   https://github.com/helix-editor/helix/blob/e7ac2fcdecfdbf43a4f772e7f7c163b43b3d0b9b/helix-dap/src/client.rs#L64-L68
                # delve upported DAP transports: tcp/unix
                #
                # $ dlv dap -l 127.0.0.1:6969 --only-same-user=false
                debugger = {
                  name = "Delve";
                  command = "nc";
                  args = [
                    "localhost"
                    "6969"
                  ];

                  # args = ["exec", "-i", "--workdir=/workspace", "tbdev-fog-default", "dlv"]
                  transport = "stdio";

                  # path = "/tmp/build/delve.sock"
                  # port = 44913
                  # env = { "ENV_VAR" = "value" }
                  # cwd = "/path/to/project"
                  templates = [
                    {
                      name = "Debug Joonas";
                      request = "launch";

                      # completion = [ { name = "entrypoint", completion = "filename", default = "." } ]
                      args = {
                        mode = "debug";
                        program = "./cmd/fog-debug";
                      };

                      # args = { mode = "exec", program = "/workspace/rel/fog-debug_linux-amd64"}
                    }

                    {
                      name = "Run all tests";
                      request = "launch";

                      # completion = [ { name = "entrypoint", completion = "filename", default = "." } ]
                      args = {
                        mode = "test";
                        program = "./pkg/scheduler";
                      };

                      # args = { mode = "exec", program = "/workspace/rel/fog-debug_linux-amd64"}
                    }
                  ];
                };
              }

              {
                name = "python";
                "language-servers" = [ "turbobob-lsp" ];
              }

              {
                name = "json";
                "language-servers" = [ "turbobob-generic-json" ];
              }

              {
                name = "html";
                "language-servers" = [ "turbobob-generic-html" ];
              }

              {
                name = "css";
                "language-servers" = [ "turbobob-generic-css" ];
              }

              {
                name = "bash";
                "language-servers" = [ "turbobob-generic-bash" ];
              }

              {
                name = "yaml";
                "language-servers" = [ "turbobob-generic-yaml" ];
              }

              {
                name = "docker-compose";
                "language-servers" = [ "turbobob-generic-yaml" ];
              }

              {
                name = "markdown";
                "language-servers" = [ "marksman" ];
              }

              {
                name = "tsx";
                "language-servers" = [ "turbobob-lang-from-project-tsx" ];
              }

              {
                name = "typescript";
                "language-servers" = [ "turbobob-lang-from-project-tsx" ];
              }

              {
                name = "nix";
                "language-servers" = [ "nil" ];
              }

              # doesn't work yet.
              # {
              #   name = "dockerfile";
              #   "language-servers" = [ "turbobob-generic-dockerfile" ];
              # }
            ];
          };
        };
      };

    };
  };

  environment = {
    sessionVariables = {
      # https://unix.stackexchange.com/q/4859 (VISUAL vs. EDITOR – what’s the difference?)
      EDITOR = "${pkgs.helix}/bin/hx";
      VISUAL = "${pkgs.helix}/bin/hx";
    };
  };
}
