package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/url"
	"reflect"
	"time"
)

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

		var input []interface{}

		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println(err)
		}

		if reflect.TypeOf(input[0]).String() != "float64" {
			continue
		}

		if int(input[0].(float64)) == 2 {
			if err = c.WriteMessage(websocket.TextMessage, []byte(`[5,{"room": "`+room+`"}]`)); err != nil {
				fmt.Println(err.Error())
				return ""
			}
			fmt.Println("send message!")
			continue
		}

		if int(input[0].(float64)) == 6 {
			return input[1].(map[string]interface{})["url"].(string)
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

	xurl := getWS(room, string(xkey))

	wsUrl, _ := url.Parse(xurl)

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

		var input []interface{}

		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println(err)
		}

		if reflect.TypeOf(input[0]).String() == "float64" {
			if int(input[0].(float64)) == 27 && input[1].(map[string]interface{})["subject_username"] != nil && input[1].(map[string]interface{})["value"] != nil {
				//[27,{"recent_tips":88,"value":1,"subject_username":"max4you","is_tip":true,"is_stealth":false},1,["<b>max4you</b> tipped <b>1</b> tokens"]]
				fmt.Println(input[1].(map[string]interface{})["subject_username"].(string), "send", int64(input[1].(map[string]interface{})["value"].(float64)), "to", room)
			}
		}
	}
}
