
// virtual trees whose state we are not concerned about
// TODO: this is empty directory - mountpoint - ignore those automatically?
# dir_always_ignore { key = "/sto" }

# dir_always_ignore { key = "/etc" }
# dir_always_ignore { key = "/boot" }
dir_always_ignore { key = "/root/.config" } // we shouldn't use root user much, so no use in persisting root's config

# ---------------- sublime text  ----------------

# sessions keep state of recent sessions. don't need that.
file_always_ignore { key = "/home/joonas/.config/sublime-text-3/Local/Auto Save Session.sublime_session" }
file_always_ignore { key = "/home/joonas/.config/sublime-text-3/Local/Session.sublime_session" }

# this file is empty
file_always_ignore { key = "/home/joonas/.config/sublime-text-3/Packages/User/Package Control.user-ca-bundle" }

# ---------------- browsers  ----------------

dir_always_ignore { key = "/home/joonas/.mozilla" }
dir_always_ignore { key = "/home/joonas/.config/BraveSoftware" }

# ---------------- VLC  ----------------

file_always_ignore { key = "/home/joonas/.config/vlc/vlc-qt-interface.conf" } # never contains anything interesting
file_always_ignore { key = "/home/joonas/.local/share/vlc/ml.xspf" } # media library

# ---------------- flameshot  ----------------

file_always_ignore { key = "/home/joonas/.config/Dharkael/flameshot.ini" } # last saved folder

# ---------------- Thunar  ----------------

file_always_ignore { key = "/home/joonas/.config/Thunar/accels.scm" } # automated key shortcut dump
file_always_ignore { key = "/home/joonas/.config/Thunar/uca.xml" } # "Example for a custom action" ??
file_always_ignore { key = "/home/joonas/.config/xfce4/xfconf/xfce-perchannel-xml/thunar.xml" } # Window sizes etc. state

# ---------------- xfce4-notifyd  ----------------

file_always_ignore { key = "/home/joonas/.config/xfce4/xfconf/xfce-perchannel-xml/xfce4-notifyd.xml" } # known applications etc. notification state

# ---------------- mousepad  ----------------

file_always_ignore { key = "/home/joonas/.config/Mousepad/accels.scm" } # automated key shortcut dump

# ---------------- thunderbolt state  ----------------

dir_always_ignore { key = "/var/lib/boltd/bootacl" }
dir_always_ignore { key = "/var/lib/boltd/domains" }

# ---------------- recently used files / saved files / bookmarks (Sublime Text / Brave / Mousepad)  ----------------

file_always_ignore { key = "/home/joonas/.local/share/recently-used.xbel" }

# ---------------- CUPS  ----------------

dir_always_ignore { key = "/var/spool/cups/tmp" }
file_always_ignore { key = "/etc/cups/printers.conf.O" }
file_always_ignore { key = "/etc/cups/subscriptions.conf.O" }
dir_always_ignore { key = "/etc/cups/ppd" } # automatically discovered printers?
file_always_ignore { key = "/etc/cups/printers.conf" }
file_always_ignore { key = "/etc/cups/subscriptions.conf" } # subscriptions (like printers-changed) state?

# ---------------- generic cache places  ----------------

# caches by definition don't contain state we can't lose

dir_always_ignore { key = "/tmp" }
dir_always_ignore { key = "/var/cache" }
dir_always_ignore { key = "/root/.cache" }
dir_always_ignore { key = "/home/joonas/.cache" }
dir_always_ignore { key = "/home/joonas/.config/sublime-text-3/Cache" }
file_always_ignore { key = "/etc/ld.so.cache" }
file_always_ignore { key = "/usr/share/applications/mimeinfo.cache" }

# Qt file dialog's history, shortcuts etc.
file_always_ignore { key = "/home/joonas/.config/QtProject.conf" }

# console-setup cache files: console font and encoding

file_always_ignore { key = "/etc/console-setup/Uni2-Fixed16.psf.gz" }
file_always_ignore { key = "/etc/console-setup/cached_UTF-8_del.kmap.gz" }
file_always_ignore { key = "/etc/console-setup/cached_Uni2-Fixed16.psf.gz" }
file_always_ignore { key = "/etc/console-setup/cached_setup_font.sh" }
file_always_ignore { key = "/etc/console-setup/cached_setup_keyboard.sh" }
file_always_ignore { key = "/etc/console-setup/cached_setup_terminal.sh" }

# ---------------- vim ----------------

file_always_ignore { key = "/home/joonas/.viminfo" }
file_always_ignore { key = "/home/joonas/.vim/.netrwhist" } # directory history etc.

# ---------------- these things are supposed to change ----------------

dir_always_ignore { key = "/home/joonas/.config/pulse" }

file_always_ignore { key = "/home/joonas/.Xauthority" }
file_always_ignore { key = "/home/joonas/.bash_history" }
file_always_ignore { key = "/root/.bash_history" }
file_always_ignore { key = "/home/joonas/.lesshst" }
file_always_ignore { key = "/root/.lesshst" }
file_always_ignore { key = "/home/joonas/.wget-hsts" }
file_always_ignore { key = "/home/joonas/.xsession-errors" }
file_always_ignore { key = "/home/joonas/.xsession-errors.old" }

# could be important in regular systems, but not relevant in our use (AFAIK we aren't be using grub)
file_always_ignore { key = "/boot/grub/grubenv" }

file_always_ignore { key = "/etc/machine-id" } // This ID uniquely identifies the host. It should be considered "confidential"
file_always_ignore { key = "/etc/.pwd.lock" } // lock file for /etc/passwd and /etc/shadow

dir_always_ignore { key = "/var/crash" }         // crash dumps
dir_always_ignore { key = "/var/backups" }       // I don't keep backups here
dir_always_ignore { key = "/var/log" }                    // runtime log files

# netplan writes Python cache files in its own dir, because how else software could work!
dir_always_ignore { key = "/usr/share/netplan/netplan/cli/commands/__pycache__" }
dir_always_ignore { key = "/usr/share/netplan/netplan/cli/__pycache__" }
dir_always_ignore { key = "/usr/share/netplan/netplan/__pycache__" }

dir_always_ignore { key = "/var/lib/AccountsService/users" } // remembers last used window manager for each user
dir_always_ignore { key = "/var/lib/lightdm" } // session manager runtime files
file_always_ignore { key = "/var/lib/logrotate/status" }           // logrotate status
file_always_ignore { key = "/var/lib/PackageKit/transactions.db" } // abstraction over package managers (apt, yum,  ..)
dir_always_ignore { key = "/var/lib/systemd/timers" } // "stamp files" - for tracking last runs of systemd timers
file_always_ignore { key = "/var/lib/systemd/random-seed" }
file_always_ignore { key = "/var/lib/systemd/timesync/clock" }
file_always_ignore { key = "/var/lib/NetworkManager/NetworkManager-intern.conf" } // internal state file
file_always_ignore { key = "/var/lib/NetworkManager/NetworkManager.state" }
file_always_ignore { key = "/var/lib/NetworkManager/seen-bssids" }
file_always_ignore { key = "/var/lib/NetworkManager/timestamps" }
file_always_ignore { key = "/var/lib/alsa/asound.state" } # Alsa's volume levels & other state I guess
file_always_ignore { key = "/home/joonas/.dmrc" } # last used graphical session (e.g. "i3")

// "an abstraction for enumerating power devices, listening to device events and querying history and statistics"
dir_always_ignore { key = "/var/lib/upower" }

// Samba runtime files
dir_always_ignore { key = "/var/lib/samba/private/msg.sock" }

// Wireshark runtime files
file_always_ignore { key = "/home/joonas/.config/wireshark/recent" }
file_always_ignore { key = "/home/joonas/.config/wireshark/recent_common" }

// Snap runtime files
file_always_ignore { key = "/snap/README" }
file_always_ignore { key = "/var/lib/snapd/apparmor/snap-confine/overlay-root" }
file_always_ignore { key = "/var/lib/snapd/assertions/asserts-v0/model/16/generic/generic-classic/active" }
file_always_ignore { key = "/var/lib/snapd/features/classic-preserves-xdg-runtime-dir" }
file_always_ignore { key = "/var/lib/snapd/features/robust-mount-namespace-updates" }
file_always_ignore { key = "/var/lib/snapd/maintenance.json" }
file_always_ignore { key = "/var/lib/snapd/seccomp/bpf/global.bin" }
file_always_ignore { key = "/var/lib/snapd/state.json" }
file_always_ignore { key = "/var/lib/snapd/system-key" }

// Lazygit
file_always_ignore { key = "/home/joonas/.config/jesseduffield/lazygit/state.yml" }

// Libreoffice writes a bunch of stuff
dir_always_ignore { key = "/home/joonas/.config/libreoffice" }
file_always_ignore { key = "/usr/lib/libreoffice/share/fonts/truetype/.uuid" }

# ---------------- package management meta type dirs ----------------

# install packages -> files get added here. in our context these are not necessary.

dir_always_ignore { key = "/var/lib/dpkg" }
dir_always_ignore { key = "/var/lib/apt" } 
dir_always_ignore { key = "/usr/share/metainfo" } // package metadata in new cross-distro XML format https://wiki.debian.org/AppStream
dir_always_ignore { key = "/usr/share/dpkg" } // C-build related makefiles?
dir_always_ignore { key = "/usr/share/lintian" } // "checks Debian software packages for common inconsistencies and errors"

# ---------------- pretty sure there is never interesting state here (should WARN though) ----------------

dir_always_ignore { key = "/etc/ssl/certs" }
dir_always_ignore { key = "/usr/include" }               // never interested about shared development C code..
dir_always_ignore { key = "/usr/share/build-essential" } // never interested about shared development C code..
dir_always_ignore { key = "/usr/share/bug" }             // how to share bugreports
dir_always_ignore { key = "/usr/share/doc" }             // documentation
dir_always_ignore { key = "/usr/share/fonts" }
dir_always_ignore { key = "/usr/share/fontconfig" }
dir_always_ignore { key = "/usr/share/mime" }  // content type database
dir_always_ignore { key = "/usr/share/icons" }
dir_always_ignore { key = "/usr/share/locale" }       // translations
file_always_ignore { key = "/usr/lib/locale/locale-archive" }
file_always_ignore { key = "/etc/locale.gen" }
dir_always_ignore { key = "/usr/share/man" }          // program user manuals
dir_always_ignore { key = "/usr/share/doc-base" }     // "each package that installs online manuals (any format) should register its manuals to doc-base."
dir_always_ignore { key = "/usr/share/info" }
dir_always_ignore { key = "/usr/share/libexttextcat" } // N-Gram-Based Text Categorization library primarily intended for language guessing
dir_always_ignore { key = "/usr/share/liblangtag" }    // "access/deal with tags for identifying languages"
dir_always_ignore { key = "/usr/share/librevenge" }    // "library for writing document import filters"
dir_always_ignore { key = "/usr/share/perl5" }
dir_always_ignore { key = "/usr/share/perl-openssl-defaults" }
dir_always_ignore { key = "/usr/share/glib-2.0/schemas" } // dconf settings schemas per each app. I don't put interesting stuff here. diff here indicates new app installation
dir_always_ignore { key = "/var/lib/colord" }             // "manage, install and generate color profiles to accurately color manage input and output devices"
dir_always_ignore { key = "/var/lib/ucf" }                // "preserve user changes to config files" (joonas-sys makes this unnecessary)
dir_always_ignore { key = "/usr/share/bash-completion/completions" }
dir_always_ignore { key = "/usr/share/zsh/vendor-completions" }

// "designed to provide a highly sophisticated, innovative and integrated desktop" (indexer, tag & metadata DB)
dir_always_ignore { key = "/usr/share/tracker" }
dir_always_ignore { key = "/usr/share/tracker-miners/extract-rules" }

// "subset of systemctl for machines not running systemd"
// changes in here signal installation of new software
dir_always_ignore { key = "/var/lib/systemd/deb-systemd-helper-enabled" }
dir_always_ignore { key = "/var/lib/systemd/deb-systemd-user-helper-enabled" }

// "Scopes are search engines" } like application results for a search keyword can be searched from these backends..
// this is a discovery file, so if there's a scope backend somewhere it'll show up in a diff anyways
file_always_ignore { key = "/usr/share/unity/client-scopes.json" }

// core dump
file_always_ignore { key = "/core" }

# generic backups of passwd etc. files. we don't need backups with our approach.
#   https://unix.stackexchange.com/questions/53128/difference-between-passwd-and-passwd-file
file_always_ignore { key = "/etc/fstab.orig" }
file_always_ignore { key = "/etc/group-" }
file_always_ignore { key = "/etc/gshadow-" }
file_always_ignore { key = "/etc/passwd-" }
file_always_ignore { key = "/etc/shadow-" }

# https://askubuntu.com/questions/813942/is-it-possible-to-stop-sudo-as-admin-successful-being-created
# file /home/joonas/.sudo_as_admin_successful
