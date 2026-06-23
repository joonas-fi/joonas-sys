{ pkgs }:

pkgs.writeShellApplication {
  name = "seedbox-mount";

  runtimeInputs = [
    pkgs.rclone
  ];

  text = ''
    set -eu

    mountpoint=/tmp/seedbox
    password="$(< /sysroot/apps/seedbox/password)"

    mkdir -p "$mountpoint"
    echo "Mounting at $mountpoint"

    rclone mount :sftp:downloads/manual/ "$mountpoint" \
      --sftp-host voyager.myseedbox.site \
      --sftp-user evoseedb_p34320 \
      --sftp-pass "$(rclone obscure "$password")"
  '';
}
