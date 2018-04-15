package main

import (
	"encoding/json"
	"fmt"
	"github.com/thoj/go-ircevent"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Truncate(str string, length int) string {
	if len(str) > length {
		return str[:length]
	} else {
		return str
	}
}

func PrivmsgTruncate(irccon *irc.Connection, target, message string) {
	irccon.Privmsg(target, Truncate(message, 200))
}

func PrivmsgWrapper(irccon *irc.Connection, target, message string) {
	irccon.Privmsg(target, fmt.Sprintf("%s -- %d", Truncate(message, 200), time.Now().Unix()))
}

func GetURLContents(url string) []byte {
	if response, err := http.Get(url); err != nil {
		panic(err)
	} else {
		defer response.Body.Close()
		if content, err := ioutil.ReadAll(response.Body); err != nil {
			panic(err)
		} else {
			return content
		}

	}
}

type Board []struct {
	Threads []struct {
		No      int    `json:"no"`
		Com     string `json:"com"`
		Name    string `json:"name"`
		Time    int    `json:"time"`
		Replies int    `json:"replies"`
		Sticky  int    `json:"sticky"`
		Locked  int    `json:"locked"`
		Sub     string `json:"sub,omitempty"`
	} `json:"threads"`
	Page int `json:"page"`
}

func Cmd8ChanCatalog(irccon *irc.Connection, target string, arguments []string) {
	catalog := GetURLContents(fmt.Sprintf("https://8ch.net/%s/catalog.json", arguments[0]))
	unmarshaled := Board{}
	if err := json.Unmarshal(catalog, &unmarshaled); err != nil {
		panic(err)
	} else {
		// This is where the fun starts
		for _, thread := range unmarshaled[0].Threads {
			if thread.Sticky != 1 && thread.Locked != 1 {
				PrivmsgTruncate(irccon, target, fmt.Sprintf("http://8ch.net/%s/res/%d.html Replies:%d Subject:%s",
					arguments[0], thread.No, thread.Replies, thread.Sub))
				time.Sleep(1500 * time.Millisecond)
			}
		}
	}
}

func main() {
	var (
		Server   = "irc.rizon.net:6667"
		Nick     = "[k00l]gobot"
		Channels = []string{"#/tech/", "#/pone/", "#templeos"}
	)

	commands := map[string]func(*irc.Connection, string, []string){
		"8ChanCatalog": Cmd8ChanCatalog,
	}

	irccon := irc.IRC(Nick, Nick)

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range Channels {
			irccon.Join(channel)
		}
	})

	irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
		if strings.HasPrefix(e.Nick, "[k00l]") && strings.HasPrefix(e.Message(), ",") {
			args := strings.Split(e.Message()[1:], " ")
			if command, exists := commands[args[0]]; exists {
				command(irccon, e.Arguments[0], args[1:])
			}
		}
	})

	if err := irccon.Connect(Server); err != nil {
		panic(err)
	}

	irccon.Loop()
}
