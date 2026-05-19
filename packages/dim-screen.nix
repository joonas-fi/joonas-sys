{
  pkgs ? import <nixpkgs> { },
  lib ? pkgs.lib,
  rustPlatform ? pkgs.rustPlatform,
  fetchFromGitHub ? pkgs.fetchFromGitHub,
  installShellFiles ? pkgs.installShellFiles,
  libxkbcommon ? pkgs.libxkbcommon,
  pkg-config ? pkgs.pkg-config,
  scdoc ? pkgs.scdoc,
  wayland ? pkgs.wayland,
}:

rustPlatform.buildRustPackage rec {
  # package name different than project name ("dim") because:
  # > dim packages are titled as dim-screen to avoid naming conflicts.
  pname = "dim-screen";
  version = "0.4.1-unstable-2026-01-11";

  src = fetchFromGitHub {
    owner = "marcelohdez";
    repo = "dim";
    rev = "84034e1c7de8ff7679e633c92deabe4d00201bc8";
    hash = "sha256-BwUBCRmNI3WfX7aTX7W8XnDX6u9NDBWz9CMaAyvbDY4=";
  };

  cargoHash = "sha256-IS86bGdf5jykZg5FI76ChD5wzoNJ6lLct2I7qylLs58=";

  nativeBuildInputs = [
    installShellFiles
    pkg-config
    scdoc
  ];

  buildInputs = [
    libxkbcommon
    wayland
  ];

  postInstall = ''
    scdoc < man/dim.1.scd > dim.1
    installManPage dim.1

    completions_dir="$(mktemp -d)"
    "$out/bin/dim" --gen-completions "$completions_dir"
    installShellCompletion \
      --bash "$completions_dir/dim.bash" \
      --fish "$completions_dir/dim.fish" \
      --zsh "$completions_dir/_dim"
  '';

  meta = {
    description = "Native Wayland screen dimming tool";
    homepage = "https://github.com/marcelohdez/dim";
    license = lib.licenses.gpl3Only;
    maintainers = with lib.maintainers; [ ];
    mainProgram = "dim";
    platforms = lib.platforms.linux;
  };
}
