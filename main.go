package main

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
	flag "github.com/spf13/pflag"
	"log"
	"net"
	"os"
	"sort"
	"strings"
)

var knownPlayers = map[string]string{
	"plasma-browser-integration": "Browser",
	"noson":                      "Noson",
}

type Player struct {
	player                               dbus.BusObject
	fullName, name, title, artist, album string
	playing, stopped                     bool
	metadata                             map[string]dbus.Variant
	conn                                 *dbus.Conn
}

const (
	INTERFACE = "org.mpris.MediaPlayer2"
	PATH      = "/org/mpris/MediaPlayer2"
	// NameOwnerChanged
	MATCH_NOC = "type='signal',path='/org/freedesktop/DBus',interface='org.freedesktop.DBus',member='NameOwnerChanged'"
	// PropertiesChanged
	MATCH_PC = "type='signal',path='/org/mpris/MediaPlayer2',interface='org.freedesktop.DBus.Properties'"
	SOCK     = "/tmp/waybar-mpris.sock"
)

var (
	PLAY      = "▶"
	PAUSE     = ""
	SEP       = " - "
	ORDER     = "SYMBOL:ARTIST:ALBUM:TITLE"
	AUTOFOCUS = false
	COMMANDS  = []string{"player-next", "player-prev", "next", "prev", "toggle"}
)

// NewPlayer returns a new player object.
func NewPlayer(conn *dbus.Conn, name string) (p *Player) {
	playerName := strings.ReplaceAll(name, INTERFACE+".", "")
	for key, val := range knownPlayers {
		if strings.Contains(name, key) {
			playerName = val
			break
		}
	}
	p = &Player{
		player:   conn.Object(name, PATH),
		conn:     conn,
		name:     playerName,
		fullName: name,
	}
	p.Refresh()
	return
}

// Refresh grabs playback info.
func (p *Player) Refresh() (err error) {
	val, err := p.player.GetProperty(INTERFACE + ".Player.PlaybackStatus")
	if err != nil {
		p.playing = false
		p.stopped = false
		p.metadata = map[string]dbus.Variant{}
		p.title = ""
		p.artist = ""
		p.album = ""
		return
	}
	strVal := val.String()
	if strings.Contains(strVal, "Playing") {
		p.playing = true
		p.stopped = false
	} else if strings.Contains(strVal, "Paused") {
		p.playing = false
		p.stopped = false
	} else {
		p.playing = false
		p.stopped = true
	}
	metadata, err := p.player.GetProperty(INTERFACE + ".Player.Metadata")
	if err != nil {
		p.metadata = map[string]dbus.Variant{}
		p.title = ""
		p.artist = ""
		p.album = ""
		return
	}
	p.metadata = metadata.Value().(map[string]dbus.Variant)
	switch artist := p.metadata["xesam:artist"].Value().(type) {
	case []string:
		p.artist = strings.Join(artist, ", ")
	case string:
		p.artist = artist
	default:
		p.artist = ""
	}
	switch title := p.metadata["xesam:title"].Value().(type) {
	case string:
		p.title = title
	default:
		p.title = ""
	}
	switch album := p.metadata["xesam:album"].Value().(type) {
	case string:
		p.album = album
	default:
		p.album = ""
	}
	return nil
}

func (p *Player) JSON() string {
	data := map[string]string{}
	symbol := PLAY
	data["class"] = "paused"
	if p.playing {
		symbol = PAUSE
		data["class"] = "playing"
	}
	var items []string
	order := strings.Split(ORDER, ":")
	for _, v := range order {
		if v == "SYMBOL" {
			items = append(items, symbol)
		} else if v == "ARTIST" {
			if p.artist != "" {
				items = append(items, p.artist)
			}
		} else if v == "ALBUM" {
			if p.album != "" {
				items = append(items, p.album)
			}
		} else if v == "TITLE" {
			if p.title != "" {
				items = append(items, p.title)
			}
		}
	}
	if len(items) == 0 {
		return "{}"
	}
	text := ""
	for i, v := range items {
		right := ""
		if v == symbol && i != len(items)-1 {
			right = " "
		} else if i != len(items)-1 && items[i+1] != symbol {
			right = SEP
		} else {
			right = " "
		}
		text += v + right
	}

	data["tooltip"] = fmt.Sprintf(
		"%s\nby %s\n",
		p.title,
		p.artist)
	if p.album != "" {
		data["tooltip"] += "from " + p.album + "\n"
	}
	data["tooltip"] += "(" + p.name + ")"
	data["text"] = text
	out, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(out)
}

type PlayerList struct {
	list    List
	current uint
	conn    *dbus.Conn
}

type List []*Player

func (ls List) Len() int {
	return len(ls)
}

func (ls List) Less(i, j int) bool {
	var states [2]uint8
	for i, p := range []bool{ls[i].playing, ls[j].playing} {
		if p {
			states[i] = 1
		}
	}
	// Reverse order
	return states[0] > states[1]
}

func (ls List) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

// Doesn't retain order since sorting if constantly done anyway
func (pl *PlayerList) Remove(fullName string) {
	currentName := pl.list[pl.current].fullName
	var i int
	found := false
	for ind, p := range pl.list {
		if p.fullName == fullName {
			i = ind
			found = true
			break
		}
	}
	if found {
		pl.list[0], pl.list[i] = pl.list[i], pl.list[0]
		pl.list = pl.list[1:]
		found = false
		for ind, p := range pl.list {
			if p.fullName == currentName {
				pl.current = uint(ind)
				found = true
				break
			}
		}
		if !found {
			pl.current = 0
			pl.Refresh()
			fmt.Println(pl.JSON())
		}
	}
	// ls[len(ls)-1], ls[i] = ls[i], ls[len(ls)-1]
	// ls = ls[:len(ls)-1]
}

func (pl *PlayerList) Reload() error {
	var buses []string
	err := pl.conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&buses)
	if err != nil {
		return err
	}
	for _, name := range buses {
		if strings.HasPrefix(name, INTERFACE) {
			pl.New(name)
		}
	}
	return nil
}

func (pl *PlayerList) New(name string) {
	pl.list = append(pl.list, NewPlayer(pl.conn, name))
	if AUTOFOCUS {
		pl.current = uint(len(pl.list) - 1)
	}
}

func (pl *PlayerList) Sort() {
	sort.Sort(pl.list)
	pl.current = 0
}

func (pl *PlayerList) Refresh() {
	for i := range pl.list {
		pl.list[i].Refresh()
	}
}

func (pl *PlayerList) JSON() string {
	if len(pl.list) != 0 {
		return pl.list[pl.current].JSON()
	}
	return "{}"
}

func (pl *PlayerList) Next() {
	pl.list[pl.current].player.Call(INTERFACE+".Player.Next", 0)
}

func (pl *PlayerList) Prev() {
	pl.list[pl.current].player.Call(INTERFACE+".Player.Previous", 0)
}

func (pl *PlayerList) Toggle() {
	pl.list[pl.current].player.Call(INTERFACE+".Player.PlayPause", 0)
}

func main() {
	flag.StringVar(&PLAY, "play", PLAY, "Play symbol/text to use.")
	flag.StringVar(&PAUSE, "pause", PAUSE, "Pause symbol/text to use.")
	flag.StringVar(&SEP, "separator", SEP, "Separator string to use between artist, album, and title.")
	flag.StringVar(&ORDER, "order", ORDER, "Element order.")
	flag.BoolVar(&AUTOFOCUS, "autofocus", AUTOFOCUS, "Auto switch to currently playing music players.")
	var command string
	flag.StringVar(&command, "send", "", "send command to already runnning waybar-mpris instance. (options: "+strings.Join(COMMANDS, "/")+")")
	flag.Parse()

	if command != "" {
		conn, err := net.Dial("unix", SOCK)
		if err != nil {
			log.Fatalln("Couldn't dial:", err)
		}
		_, err = conn.Write([]byte(command))
		if err != nil {
			log.Fatalln("Couldn't send command")
		}
		fmt.Println("Sent.")
	} else {
		conn, err := dbus.SessionBus()
		if err != nil {
			panic(err)
		}
		players := &PlayerList{
			conn: conn,
		}
		players.Reload()
		players.Sort()
		players.Refresh()
		fmt.Println(players.JSON())
		lastLine := ""
		// fmt.Println("New array", players)
		go func() {
			os.Remove(SOCK)
			listener, err := net.Listen("unix", SOCK)
			if err != nil {
				log.Fatalln("Couldn't establish socket connection at", SOCK)
			}
			defer listener.Close()
			for {
				con, err := listener.Accept()
				if err != nil {
					log.Println("Couldn't accept:", err)
					continue
				}
				buf := make([]byte, 512)
				nr, err := con.Read(buf)
				if err != nil {
					log.Println("Couldn't read:", err)
					continue
				}
				command := string(buf[0:nr])
				if command == "player-next" {
					if players.current < uint(len(players.list)-1) {
						players.current += 1
					} else {
						players.current = 0
					}
					players.Refresh()
					fmt.Println(players.JSON())
				} else if command == "player-prev" {
					if players.current != 0 {
						players.current -= 1
					} else {
						players.current = uint(len(players.list) - 1)
					}
					players.Refresh()
					fmt.Println(players.JSON())
				} else if command == "next" {
					players.Next()
				} else if command == "prev" {
					players.Prev()
				} else if command == "toggle" {
					players.Toggle()
				} else {
					fmt.Println("Invalid command")
				}
			}
		}()
		conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, MATCH_NOC)
		conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, MATCH_PC)
		c := make(chan *dbus.Signal, 10)
		conn.Signal(c)
		for v := range c {
			// fmt.Printf("SIGNAL: Sender %s, Path %s, Name %s, Body %s\n", v.Sender, v.Path, v.Name, v.Body)
			if strings.Contains(v.Name, "NameOwnerChanged") {
				switch name := v.Body[0].(type) {
				case string:
					var pid uint32
					conn.BusObject().Call("org.freedesktop.DBus.GetConnectionUnixProcessID", 0, name).Store(&pid)
					if strings.Contains(name, INTERFACE) {
						if pid == 0 {
							// fmt.Println("Removing", name)
							players.Remove(name)
						} else {
							// fmt.Println("Adding", name)
							players.New(name)
						}
					}
				}
			} else if strings.Contains(v.Name, "PropertiesChanged") && strings.Contains(v.Body[0].(string), INTERFACE+".Player") {
				players.Refresh()
				if AUTOFOCUS {
					players.Sort()
				}
				if l := players.JSON(); l != lastLine {
					lastLine = l
					fmt.Println(l)
				}
			}
		}
	}
}
