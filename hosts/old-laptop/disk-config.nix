{ lib, ... }:

{
  disko.devices = {
    disk = {
      # means partlabel will be `disk-main-...`
      main = {
        device = lib.mkDefault "/dev/disk/by-id/usb-Samsung_Flash_Drive_FIT_0374221020000616-0:0";
        type = "disk";
        content = {
          type = "gpt";
          partitions = {
            ESP = {
              type = "EF00";
              size = "512M";
              content = {
                type = "filesystem";
                format = "vfat";
                mountpoint = "/boot";
                mountOptions = [ "umask=0077" ];
              };
            };
            # partlabel will be `disk-main-root`
            root = {
              size = "100%";
              content = {
                type = "filesystem";
                format = "ext4";
                mountpoint = "/";
              };
            };
          };
        };
      };
    };
  };
}
