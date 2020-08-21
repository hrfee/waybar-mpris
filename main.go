package main

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
	"sort"
	"strings"
)

type Player struct {
	player                           dbus.BusObject
	playing, stopped                 bool
	playerName, title, artist, album string
	metadata                         map[string]dbus.Variant
	conn                             *dbus.Conn
}

func NewPlayer(conn *dbus.Conn, name string) *Player {
	p := &Player{
		player:     conn.Object(name, "/org/mpris/MediaPlayer2"),
		conn:       conn,
		playerName: strings.ReplaceAll(name, "org.mpris.MediaPlayer2.", ""),
	}
	p.Refresh()
	return p
}

func (p *Player) Refresh() {
	val, err := p.player.GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	if err != nil {
		panic(err)
	}
	if strings.Contains(val.String(), "Playing") {
		p.playing = true
		p.stopped = false
	} else if strings.Contains(val.String(), "Paused") {
		p.playing = false
		p.stopped = false
	} else {
		p.playing = false
		p.stopped = true
	}
	md, err := p.player.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return
	}
	p.metadata = md.Value().(map[string]dbus.Variant)
	p.artist = strings.Join(p.metadata["xesam:artist"].Value().([]string), ", ")
	p.title = p.metadata["xesam:title"].Value().(string)
	p.album = p.metadata["xesam:album"].Value().(string)
}

func (p *Player) JSON() string {
	data := map[string]string{}
	data["tooltip"] = fmt.Sprintf(
		"%s\nby %s\nfrom %s\n(%s)",
		p.title,
		p.artist,
		p.album,
		p.playerName)
	var symbol string
	if p.playing {
		data["class"] = "playing"
		symbol = ""
	} else {
		data["class"] = "paused"
		symbol = "▶"
	}
	data["text"] = fmt.Sprintf(
		"%s %s - %s - %s",
		symbol,
		p.artist,
		p.album,
		p.title)

	text, _ := json.Marshal(data)
	return string(text)
}

type Players []*Player

func (s Players) Len() int {
	return len(s)
}

func (s Players) Less(i, j int) bool {
	// x, err := s[i].GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	// y, err := s[j].GetProperty("org.mpris.MediaPlayer2.Player.PlaybackStatus")
	s[i].Refresh()
	s[j].Refresh()
	var states [2]int
	if s[i].playing {
		states[0] = 1
	}
	if s[j].playing {
		states[1] = 1
	}
	// if strings.Contains(x.String(), "Playing") {
	// 	states[0] = 1
	// }
	// if strings.Contains(y.String(), "Playing") {
	// 	states[1] = 1
	// }
	// fmt.Println(states[i])
	// reverse
	return states[0] > states[1]
}

func (s Players) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func main() {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	var fd []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&fd)
	if err != nil {
		panic(err)
	}
	var players Players
	for _, name := range fd {
		if strings.HasPrefix(name, "org.mpris.MediaPlayer2") {
			//players = append(players, conn.Object(name, "/org/mpris/MediaPlayer2"))
			players = append(players, NewPlayer(conn, name))
		}
	}
	sort.Sort(players)
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',path='/org/mpris/MediaPlayer2',interface='org.freedesktop.DBus.Properties'")
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	fmt.Println(players[0].JSON())
	for range c {
		players[0].Refresh()
		fmt.Println(players[0].JSON())
	}
	// fmt.Printf("%d: %s: %t\n", i, s.playerName, s.playing)
	// for key, val := range s.metadata {
	// 	fmt.Println(key, val)
	// }

	// fmt.Println(fd)
	// node, err := introspect.Call(conn.Object("org.mpris.MediaPlayer2", "/org/mpris/MediaPlayer2"))
	// if err != nil {
	// 	panic(err)
	// }
	// data, _ := json.MarshalIndent(node, "", "    ")
	// fmt.Println(data)

}
