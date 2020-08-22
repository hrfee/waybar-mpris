## waybar-mpris

a custom waybar component for displaying info from MPRIS2 players. It automatically focuses on currently playing music players, and can easily be customized.

## Install
`go get github.com/hrfee/waybar-mpris`

or just grab the binary from here.

## Usage
When running, the program will pipe out json in waybar's format. Make a custom component in your configuration and set `return-type` to `json`, and `exec` to the path to the program.
```
Usage of ./waybar-mpris:
      --order string       Element order. (default "SYMBOL:ARTIST:ALBUM:TITLE")
      --pause string       Pause symbol/text to use. (default "")
      --play string        Play symbol/text to use. (default "▶")
      --separator string   Separator string to use between artist, album, and title. (default " - ")
```
* Modify the order of components with `--order`. `SYMBOL` is the play/paused icon or text, other options are self explanatory.
* `--play/--pause` specify the symbols or text to display when music is paused/playing respectively.
* --separator specifies a string to separate the artist, album and title text.

