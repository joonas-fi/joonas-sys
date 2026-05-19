# https://wiki.nixos.org/wiki/Firefox/en#Advanced
{
  home-manager.users."joonas" =
    { config, ... }:
    {
      programs.firefox = {
        enable = true;

        # silence migration nag by introducing this.
        configPath = "${config.xdg.configHome}/mozilla/firefox";

        languagePacks = [ "en-US" ];

        profiles.default = {
          settings = {
            # https://bsky.app/profile/rag.pub/post/3mmsc56fqhk27
            "gfx.font_rendering.freetype.enhanced_contrast" = 100;
            "gfx.font_rendering.freetype.gamma" = 0;

            "browser.translations.neverTranslateLanguages" = "fi";

            # doesn't actually control anything
            # "browser.tabs.firefox-view" = false; # icon from top left (left of tab bar)

            # force dark theme
            "layout.css.prefers-color-scheme.content-override" = 0;

            # sponsored stuff from new tab page
            "browser.newtabpage.activity-stream.showSponsoredTopSites" = false;
            "browser.newtabpage.activity-stream.showSponsored" = false;

            "signon.generation.enabled" = false; # do not ask to autogenerate passwords for "create password" -type fields (we have Bitwarden)
          };

          search = {
            # force = true;
            # default = "DuckDuckGo";
            # privateDefault = "DuckDuckGo";
          };
        };

        policies = {
          # Updates & Background Services
          AppAutoUpdate = false;
          BackgroundAppUpdate = false;

          # Feature Disabling
          # DisableBuiltinPDFViewer = true;
          DisableFirefoxStudies = true;
          DisableFirefoxAccounts = true;
          DisableFirefoxScreenshots = true;
          DisableForgetButton = true;
          DisableMasterPasswordCreation = true;
          DisableProfileImport = true;
          DisableProfileRefresh = true;
          DisableSetDesktopBackground = true;
          DisablePocket = true;
          DisableTelemetry = true;
          DisableFormHistory = true;

          # Completely disable Firefox's built-in password manager. (this conflicts with Bitwarden's UI for generating password for example)
          # This also blocks about:logins.
          PasswordManagerEnabled = false;

          # Extensions
          ExtensionSettings =
            let
              moz = short: "https://addons.mozilla.org/firefox/downloads/latest/${short}/latest.xpi";
            in
            {
              "*".installation_mode = "blocked";

              "uBlock0@raymondhill.net" = {
                install_url = moz "ublock-origin";
                installation_mode = "force_installed";
                updates_disabled = true;
              };

              "{446900e4-71c2-419f-a6a7-df9c091e268b}" = {
                install_url = moz "bitwarden-password-manager";
                installation_mode = "force_installed";
                updates_disabled = true;
              };
            };
        };
      };
    };
}
