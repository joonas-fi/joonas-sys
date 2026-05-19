{
  lib,
  stdenv,
  dockerTools,
  runCommand,
  undocker,
  gnutar,
  coreutils,
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
      imageDigest = "sha256:ec54ba770aa4195d317240a21d291857f0fa7068c109567ab0b2a381d624754c";
      sha256 = "sha256-xut8DbyqD28RQ3EfxeSb9RdJn/FwFEF3ggdduXm2APw=";
      arch = "amd64";
    };

    aarch64-linux = {
      imageDigest = "sha256:ec54ba770aa4195d317240a21d291857f0fa7068c109567ab0b2a381d624754c";
      sha256 = lib.fakeHash;
      arch = "arm64";
    };
  };

  containerImageForCurrentSystem =
    containerImages.${stdenv.hostPlatform.system}
      or (throw "turbobob: unsupported system ${stdenv.hostPlatform.system}");

  containerImage = dockerTools.pullImage (
    {
      imageName = "ghcr.io/function61/turbobob";
      os = "linux";
    }
    // containerImageForCurrentSystem
  );
in

runCommand "turbobob"
  {
    meta = with lib; {
      description = "Modern, minimal container-based build/development tool to make any project´s dev easy and frictionless.";
      homepage = "https://github.com/function61/turbobob";
      license = licenses.asl20;
      platforms = builtins.attrNames containerImages;
      mainProgram = "bob";
    };

    nativeBuildInputs = [
      undocker
      gnutar
      coreutils
    ];
  }
  ''
    set -euo pipefail

    mkdir -p "$out/bin"

    undocker ${containerImage} - | tar -xf - --strip-components=2 -C "$out/bin/" usr/bin/bob

  ''
