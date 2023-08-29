# About

Command `wofimoji` is a [`wofi`][wofi] emoji picker.

## Usage

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
