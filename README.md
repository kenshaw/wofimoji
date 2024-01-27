# wofimoji

Command `wofimoji` is a [`wofi`][wofi] emoji picker.

<p align="center">
  <a href="#installing" title="Installing">Installing</a> |
  <a href="#building" title="Building">Building</a> |
  <a href="#using" title="Using">Using</a> |
  <a href="https://github.com/kenshaw/wofimoji/releases" title="Releases">Releases</a>
</p>

[![Releases][release-status]][Releases]
[![Discord Discussion][discord-status]][discord]

[releases]: https://github.com/kenshaw/wofimoji/releases "Releases"
[release-status]: https://img.shields.io/github/v/release/kenshaw/wofimoji?display_name=tag&sort=semver "Latest Release"
[discord]: https://discord.gg/WDWAgXwJqN "Discord Discussion"
[discord-status]: https://img.shields.io/discord/829150509658013727.svg?label=Discord&logo=Discord&colorB=7289da&style=flat-square "Discord Discussion"

## Installing

`wofimoji` can be installed [via Release][], [via AUR][] or [via Go][]:

[via Release]: #installing-via-release
[via AUR]: #installing-via-aur-arch-linux
[via Go]: #installing-via-go

### Installing via Release

1. [Download a release for your platform][releases]
2. Extract the `wofimoji` file from the `.tar.bz2` file
3. Move the extracted executable to somewhere on your `$PATH`

### Installing via AUR (Arch Linux)

Install `wofimoji` from the [Arch Linux AUR][aur] in the usual way with the [`yay`
command][yay]:

```sh
# install
$ yay -S wofimoji
```

Alternately, build and [install using `makepkg`][arch-makepkg]:

```sh
# clone package repo and make/install package
$ git clone https://aur.archlinux.org/wofimoji.git && cd wofimoji
$ makepkg -si
==> Making package: wofimoji 0.2.0-1 (Sat 11 Nov 2023 02:30:02 PM WIB)
==> Checking runtime dependencies...
==> Checking buildtime dependencies...
==> Retrieving sources...
...
```

### Installing via Go

Install `wofimoji` in the usual Go fashion:

```sh
# install latest wofimoji version
$ go install github.com/kenshaw/wofimoji@latest
```

## Using

```sh
$ wofimoji --help
wofimoji, the wofi emoji picker

Usage:
  wofimoji [flags]

Flags:
  -a, --action action           action (default copy)
      --clipboarder string      clipboarder command (default "wl-copy")
  -c, --config string           config file
  -h, --help                    help for wofimoji
      --prompt string           wofi prompt (default "emoji")
      --selector string         selector command (default "wofi")
  -f, --selector-args strings   selector args
  -t, --skin-tone tone          skin tone (default neutral)
  -T, --template file           template file
      --typer string            typer command (default "wtype")
  -v, --version                 version for wofimoji
```

### Sway

With `sway`, or other similar windowing system:

```conf
bindsym {
  # launch wofimoji using light skin tones, and passing a unique cache file to wofi
  Mod4+e exec wofimoji --skin-tone light --selector-args "--cache-file=$HOME/.cache/wofi-wofimoji"
}
```

Same, but shorter:

```
bindsym {
  Mod4+e exec wofimoji -t light -f "--cache-file=$HOME/.cache/wofi-wofimoji"
}
```

[wofi]: https://sr.ht/~scoopta/wofi/
[aur]: https://aur.archlinux.org/packages/wofimoji
[arch-makepkg]: https://wiki.archlinux.org/title/makepkg
[yay]: https://github.com/Jguer/yay
