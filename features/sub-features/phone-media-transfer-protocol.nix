{
  pkgs,
  ...
}:
{
  services = {
    gvfs.enable = true;
  };

  environment.systemPackages = with pkgs; [
    glib # gio
    libmtp # mtp-detect, mtp-files, mtp-folders, mtp-connect
  ];
}
