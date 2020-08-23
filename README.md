## waybar-mpris
<p align="center">
    <img src="images/cropped.gif" style="width: 100%;" alt="bar gif"></img>
</p>

a waybar component/utility for displaying and controlling MPRIS2 compliant media players individually, inspired by [waybar-media](https://github.com/yurihs/waybar-media). 

MPRIS2 is widely supported, so this component should work with:
* Chrome/Chromium
* Other browsers (with kde plasma integration installed)
* VLC
* Spotify
* Noson
* mpd (with [mpDris2](https://github.com/eonpatapon/mpDris2))
* Most other music/media players

## Install
`go get github.com/hrfee/waybar-mpris` will install the program, as well as the go dbus bindings and pflags for command-line arguments.

or just grab the `waybar-mpris` binary from here and place it in your PATH.

## Usage
When running, the program will pipe out json in waybar's format. Add something like this to your waybar `config.json`:
```
"custom/waybar-mpris": {
    "return-type": "json",
    "exec": "waybar-mpris",
    "on-click": "waybar-mpris --send toggle",
    // This option will switch between players on right click.
        "on-click-right": "waybar-mpris --send player-next",
    // The options below will switch the selected player on scroll
        // "on-scroll-up": "waybar-mpris --send player-next",
        // "on-scroll-down": "waybar-mpris --send player-prev",
    // The options below will go to next/previous track on scroll
        // "on-scroll-up": "waybar-mpris --send next",
        // "on-scroll-down": "waybar-mpris --send prev",
    "escape": true,
},
```


```
Usage of waybar-mpris:
      --autofocus          Auto switch to currently playing music players.
      --order string       Element order. (default "SYMBOL:ARTIST:ALBUM:TITLE")
      --pause string       Pause symbol/text to use. (default "\uf8e3")
      --play string        Play symbol/text to use. (default "▶")
      --send string        send command to already runnning waybar-mpris instance. (options: player-next/player-prev/next/prev/toggle)
      --separator string   Separator string to use between artist, album, and title. (default " - ")
```

* Modify the order of components with `--order`. `SYMBOL` is the play/paused icon or text, other options are self explanatory.
* `--play/--pause` specify the symbols or text to display when music is paused/playing respectively.
* `--separator` specifies a string to separate the artist, album and title text.
* `--autofocus` makes waybar-mpris automatically focus on currently playing music players.
* `--send` sends commands to an already running waybar-mpris instance via a unix socket. Commands:
  * `player-next`: Switch to displaying and controlling next available player.
  * `player-prev`: Same as `player-next`, but for the previous player.
  * `next/prev`: Next/previous track on the selected player.
  * `toggle`: Play/pause.
  * *Note: you can also bind these commands to keys in your sway/other wm config.*

