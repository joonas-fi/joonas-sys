
// https://superuser.com/a/130052
pref('browser.ctrlTab.recentlyUsedOrder', false);

// disable automatic updates, because we're always installing the latest version
// doesn't work anymore, needed to do it with policies instead: https://support.mozilla.org/en-US/kb/customizing-firefox-using-policiesjson
// pref('app.update.auto', false);
// pref('app.update.enabled', false);

// welcome to Firefox
pref('trailhead.firstrun.didSeeAboutWelcome', true);

// disable nag when entering about:config
pref('browser.aboutConfig.showWarning', false);

// https://stackoverflow.com/a/47353456
pref('datareporting.policy.firstRunURL', '');
