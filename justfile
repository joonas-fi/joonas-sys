_default:
    @just --list --unsorted

shell-in-build-container: (run-in-container "bash")

rebuild-build-only nixosConfig: (run-in-container "nixos-rebuild" "build" "--flake=.#" + nixosConfig)

rebuild-switch-bedroom-pi: (run-in-container "nixos-rebuild" "switch" "--flake=.#bedroom-pi" "--target-host=joonas@192.168.1.210" "--sudo")
    
rebuild-switch-garage: (run-in-container "nixos-rebuild" "switch" "--flake=.#garage" "--target-host=joonas@192.168.1.234" "--sudo")

rebuild-switch-old-laptop: (run-in-container "nixos-rebuild" "switch" "--flake=.#old-laptop" "--target-host=joonas@192.168.2.105" "--sudo")
    
fmt: (run-in-container "nix" "fmt")

flake-show: (run-in-container "nix" "flake" "show")

# test-in-vm: (run-in-container "nixos-rebuild build-image" "--image-variant=qemu" "--flake=.#old-laptop")

[positional-arguments]
run-in-container *dummyArgs:
    #!/bin/bash -eu
    if [ -z "${SSH_AUTH_SOCK+x}" ]; then
        echo "Need SSH agent (\$SSH_AUTH_SOCK)"
        exit 1
    fi
    
    if [[ "$SSH_AUTH_SOCK" == *"gpg-agent"* ]]; then
        # I just don't have gpg-agent configured.
        echo "gpg-agent detected. (expected OpenSSH's ssh-agent)"
        exit 1
    fi
    
    if docker container inspect nix 2>/dev/null 1>&2; then
        docker exec -it nix "$@"
    else
        docker run --rm -it --name=nix \
            --volume=nix:/nix \
            --volume=nix-cache:/root/.cache/nix \
            --volume=$(pwd):/workspace \
            --workdir=/workspace \
            --volume="$SSH_AUTH_SOCK:/ssh-agent:ro" \
            --volume=/tmp:/tmp \
            --env=SSH_AUTH_SOCK=/ssh-agent \
            --pid=host \
            --entrypoint= \
            nix "$@"
    fi
