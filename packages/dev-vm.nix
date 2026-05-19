{
  stdenvNoCC,
  lib,
  bash,
  coreutils,
  findutils,
  gawk,
  iproute2,
  nix,
  openssh,
  procps,
  util-linux,
}:

let
  runtimePath = lib.makeBinPath [
    bash
    coreutils
    findutils
    gawk
    iproute2
    nix
    openssh
    procps
    util-linux
  ];
in
stdenvNoCC.mkDerivation {
  pname = "dev-vm";
  version = "0-unstable-2026-07-24";

  src = ./dev-vm.sh;
  dontUnpack = true;

  installPhase = ''
    runHook preInstall

    install -Dm755 "$src" "$out/libexec/dev-vm/dev-vm"
    patchShebangs "$out/libexec/dev-vm/dev-vm"

    mkdir -p "$out/bin"
    cat > "$out/bin/dev-vm" <<EOF
    #!${bash}/bin/bash
    set -euo pipefail

    export PATH=${lib.escapeShellArg runtimePath}\''${PATH:+:\$PATH}
    exec -a dev-vm ${bash}/bin/bash -c 'script="\$1"; shift; source "\$script"' dev-vm "$out/libexec/dev-vm/dev-vm" "\$@"
    EOF
    chmod +x "$out/bin/dev-vm"

    runHook postInstall
  '';
}
