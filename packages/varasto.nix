{
  lib,
  stdenv,
  dockerTools,
  runCommand,
  undocker,
  gnutar,
  coreutils,
  makeWrapper,
  fuse,
}:

let
  # when you want to update this:
  # 1. put new `imageDigest` in the matching `containerImages` entry
  # 2. change the entry's `sha256` just a bit (to have something valid-looking but different)
  # 3. run this monstrosity:
  #
  # NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 NIXPKGS_ALLOW_UNFREE=1 nix flake check --all-systems --impure
  #
  # you then get specified-vs-got hashes. you will then put the new hash to the `sha256`
  containerImages = {
    x86_64-linux = {
      imageDigest = "sha256:d797adacb7f44e24e9abedcc1a9e2526aa6d0e88b6d3fc666792b207152796de";
      sha256 = "sha256-bCtHso1gI5ft3b+g/QhBeY9Yl0GX5sILQhUwrypUW50=";
      arch = "amd64";
    };

    aarch64-linux = {
      imageDigest = "sha256:d797adacb7f44e24e9abedcc1a9e2526aa6d0e88b6d3fc666792b207152796de";
      sha256 = "sha256-rV1RglbTQvp/oAfSbVRp7K159Pa5mv3jXoFtAHJvgyM=";
      arch = "arm64";
    };
  };

  containerImageForCurrentSystem =
    containerImages.${stdenv.hostPlatform.system}
      or (throw "varasto: unsupported system ${stdenv.hostPlatform.system}");

  containerImage = dockerTools.pullImage (
    {
      imageName = "fn61/varasto";
      os = "linux";
    }
    // containerImageForCurrentSystem
  );
in

runCommand "varasto"
  {
    meta = with lib; {
      description = "All your files in one simple, replicated, encrypted place - with built-in backups and configurable ransomware protection.";
      homepage = "https://github.com/function61/varasto";
      license = licenses.agpl3Only;
      platforms = builtins.attrNames containerImages;
      mainProgram = "sto";
    };

    nativeBuildInputs = [
      undocker
      gnutar
      coreutils
      makeWrapper
    ];
  }
  ''
    set -euo pipefail

    mkdir -p "$out/bin"

    # NOTE: executable is inside the image in unorthodox location
    undocker ${containerImage} - | tar -xf - --strip-components=1 -C "$out/bin/" varasto/sto

    # make `fuse` executables available to Varasto
    wrapProgram  "$out/bin/sto" \
      --prefix PATH : ${lib.makeBinPath [ fuse ]}
  ''
