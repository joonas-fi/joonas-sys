Rationale - why this approach?
==============================

This might seem like much added complexity to you, and that's a fair argument. But I'm thinking I'm
just paying the price beforehand, because it's easier to do it now than later ("Leaning in to the
pain"). Let me explain..


### Maintaining multiple computers

An unexpected side effect came from building my system programmatically: I can provision it on multiple
computers. **It feels really nice** moving from my desktop to my laptop and it behaving exactly the
same. Same programs, same settings, same keyboard shortcuts. Same SSH accessibility etc.

I plan to have my third computer (media PC) added soon.


### Intermingled state

The thing that has always bothered me is that computers tend to intermingle (or at least make it too
easy to):

- Interesting state
	* Worth preserving, backing up
- Totally **un**interesting state

This makes migrating to a new computer or mobile device hard. Systems aren't forever, and especially
with Windows it's common wisdom to
[re-install every couple of years](https://twitter.com/joonas_fi/status/1356122493019426816) because
the system has just accumulated too much crud.

As a Windows user I used to obsess over how the data was laid out in my `C:` and `D:` drives. I used
to get angry and anxious if some badly behaving software wrote its data/log files/directories to the
root of the partition:

![](windows-c-drive-unnecessary-crap.png)

The above doesn't exactly scream "this is a well-organized system". I especially like
[good old 6749525315573233238](https://www.reddit.com/r/Amd/comments/8pzm63/what_is_the_purpose_of_c6749525315573233238/).

Linux is not immune to this. Here's how my freshly installed home directory looks:

```
/home/joonas/
├── .Xauthority
├── .bash_history
├── .bash_logout
├── .bashrc
├── .cache
├── .config
├── .dbus
├── .dmrc
├── .gnupg
├── .lesshst
├── .local
├── .mozilla
├── .mplayer
├── .profile
├── .selected_editor
├── .ssh
├── .sudo_as_admin_successful
├── .thunderbird
├── .vim
├── .viminfo
├── .wget-hsts
├── .wine
├── .xsession-errors
├── Desktop
├── Downloads
├── snap
└── work -> /persist/work
```

The only entries I placed there were `.config/`, `.ssh/` and `work/`. Most of `./config/` is also
filled up with stuff I didn't choose to put there.

Without the drastic approach I'm taking, I don't think there is other way to manage one's system
state in a way that doesn't leave you with dread on data loss ("did I backup everything I care about?").

You can of course backup your entire system but then you're left with countless unnecessary files
in your backups that you've to keep forever unless you take the time to dig into the backup to inspect
if there were interesting files to recover before deleting the backups of your long-gone system.

Would you say identifying interesting state would be easier to do now (or at most a week after the state
was created), than to leave it for you do do ten years from now?

Now my backup strategy is easy: back up everything under `/persist`. That's it. All other state
(applications to install & their configuration) is defined by this repository, and I consider Git (a
distributed VCS) a sufficient backup - if GitHub loses my data, I still have a local copy (and vice versa).

[Graham Christensen](https://grahamc.com/blog/erase-your-darlings) put it eloquently:

> Over time, a system collects state on its root partition. This state lives in assorted directories
> like /etc and /var, and represents every under-documented or out-of-order step in bringing up the services.
>
>> “Right, run myapp-init.”
>
> These small, inconsequential “oh, oops” steps are the pieces that get lost and don’t appear in your runbooks.
>
>> “Just download ca-certificates to … to fix …”
>
> Each of these quick fixes leaves you doomed to repeat history in three years when you’re finally
> doing that dreaded RHEL 7 to RHEL 8 upgrade.


### Advantages of storing system state in a VCS

Now I get some kickass
[visibility into my system](https://github.com/joonas-fi/joonas-sys/commit/5c82245c04a42b8e9bd6353d7eb098700d0f558f)
& how it's evolved.

As an additional bonus, now I can be more intentional on the system's state changes: it's harder to
accidentally commit a change than just testing some config change traditionally and forgetting it there.


### Unused software gets removed automatically

Because each week you start with a fresh system with your previous uncommitted changes removed, and
there's no remnants from complicated update processes that try to keep a
[system running for decades](https://www.youtube.com/watch?v=t0rCTZ_3TQ4), there's no more anxiety
around thinking if a re-install would fix your problems!

You're always at most a week away from what amounts to 100 % legit fresh install.