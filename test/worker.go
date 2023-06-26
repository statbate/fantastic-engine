package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func getMessageID(s string) int {
	i := strings.IndexByte(s, ',')
	if i != -1 {
		if v, err := strconv.Atoi(s[1:i]); err == nil {
			return v
		}
	}
	return 0
}

func getMessageData(s string) string {
	a := strings.IndexByte(s, '{')
	b := strings.LastIndexByte(s, '}')
	if a != -1 && b != -1 {
		return s[a : b+1]
	}
	return ""
}

func getWS(room, key string) string {
	u := url.URL{Scheme: "wss", Host: "node2-ord.livemediahost.com:3000"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}

	defer c.Close()

	if err = c.WriteMessage(websocket.TextMessage, []byte(`["v3.authorize",{"token":"`+key+`"}]`)); err != nil {
		fmt.Println(err.Error())
		return ""
	}

	for {
		c.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}

		//fmt.Println(string(message))

		messageID := getMessageID(string(message))

		if messageID == 2 {
			if err = c.WriteMessage(websocket.TextMessage, []byte(`[5,{"room": "`+room+`"}]`)); err != nil {
				fmt.Println(err.Error())
				return ""
			}
			fmt.Println("send message!")
			continue
		}

		if messageID == 6 {
			input := struct {
				Url  string `json:"url"`
				Room string `json:"room"`
			}{}
			if err := json.Unmarshal([]byte(getMessageData(string(message))), &input); err != nil {
				fmt.Println(err)
			}
			return input.Url
		}
	}

	return ""
}

func statRoom(room, key string) {
	fmt.Println("Start", room, "key", key)

	if len(key) < 50 {
		fmt.Println(key, room)
		return
	}

	xkey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Println(err, room)
		return
	}

	wsUrl, _ := url.Parse(getWS(room, string(xkey)))

	u := url.URL{Scheme: "wss", Host: wsUrl.Host}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer c.Close()

	if err = c.WriteMessage(websocket.TextMessage, []byte(`["v3.authorize",{"token":"`+string(xkey)+`"}]`)); err != nil {
		fmt.Println(err.Error())
		return
	}

	for {
		c.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		//fmt.Println(string(message))

		messageID := getMessageID(string(message))

		//fmt.Println(messageID)

		if messageID == 27 {
			input := struct {
				Value    int64  `json:"value"`
				Username string `json:"subject_username"`
				Tip      bool   `json:"is_tip"`
				Stealth  bool   `json:"is_stealth"`
			}{}
			//[27,{"recent_tips":88,"value":1,"subject_username":"max4you","is_tip":true,"is_stealth":false},1,["<b>max4you</b> tipped <b>1</b> tokens"]]
			if err := json.Unmarshal([]byte(getMessageData(string(message))), &input); err != nil {
				//fmt.Println(string(message))
				//fmt.Println(err)
				continue
			}
			if !input.Tip || input.Value < 1 {
				//fmt.Println(input.Tip, input.Value, "continue")
				continue
			}
			if input.Stealth {
				input.Username = "anon_tips"
			}
			fmt.Println(input.Username, "send", input.Value, "to", room)
		}
	}
}
