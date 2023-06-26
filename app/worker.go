package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"encoding/base64"
)

var uptime = time.Now().Unix()

func mapRooms() {

	data := make(map[string]*Info)

	for {
		select {
		case m := <-rooms.Add:
			data[m.room] = &Info{Rid: m.Rid, Key: m.Key, Proxy: m.Proxy, Start: m.Start, Last: m.Last, Online: m.Online, Income: m.Income, Dons: m.Dons, Tips: m.Tips, ch: m.ch}

		case s := <-rooms.Json:
			j, err := json.Marshal(data)
			if err == nil {
				s = string(j)
			}
			rooms.Json <- s

		case <-rooms.Count:
			rooms.Count <- len(data)

		case key := <-rooms.Del:
			delete(data, key)

		case room := <-rooms.Check:
			if _, ok := data[room]; !ok {
				room = ""
			}
			rooms.Check <- room

		case room := <-rooms.Stop:
			if _, ok := data[room]; ok {
				close(data[room].ch)
			}
		}
	}
}

func announceCount() {
	for {
		time.Sleep(30 * time.Second)
		rooms.Count <- 0
		l := <-rooms.Count
		msg, err := json.Marshal(struct {
			Chanel string `json:"chanel"`
			Count  int    `json:"count"`
		}{
			Chanel: "camsoda",
			Count:  l,
		})
		if err == nil {
			socketServer <- msg
		}
	}
}

func reconnectRoom(workerData Info) {
	time.Sleep(5 * time.Second)
	fmt.Println("reconnect:", workerData.room, workerData.Proxy)
	startRoom(workerData)
}

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

func getWS(workerData Info, key []byte) string {
	Dialer := *websocket.DefaultDialer
	if _, ok := conf.Proxy[workerData.Proxy]; ok {
		Dialer = websocket.Dialer{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http", // or "https" depending on your proxy
				Host:   conf.Proxy[workerData.Proxy],
				Path:   "/",
			}),
			HandshakeTimeout: 45 * time.Second, // https://pkg.go.dev/github.com/gorilla/websocket
		}
	}
	
	u := url.URL{Scheme: "wss", Host: "node2-ord.livemediahost.com:3000"}
	c, _, err := Dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error(), u.String(), workerData.room)
		return ""
	}
	
	defer c.Close()
	
	if err = c.WriteMessage(websocket.TextMessage, []byte(`["v3.authorize",{"token":"`+string(key)+`"}]`)); err != nil {
		fmt.Println(err.Error())
		return ""
	}
	
	for {
		c.SetReadDeadline(time.Now().Add(45 * time.Second))
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error())
			return ""
		}

		//fmt.Println(string(message))

		messageID := getMessageID(string(message))

		if messageID == 2 {
			if err = c.WriteMessage(websocket.TextMessage, []byte(`[5,{"room": "`+workerData.room+`"}]`)); err != nil {
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

func xWorker(workerData Info) {
	fmt.Println("Start", workerData.room, "proxy", workerData.Proxy)

	rooms.Add <- workerData

	defer func() {
		rooms.Del <- workerData.room
	}()


	if len(workerData.Key) < 50 {
		fmt.Println(workerData.Key, workerData.room)
		return
	}
	
	key, err := base64.StdEncoding.DecodeString(workerData.Key)
	if err != nil {
		fmt.Println(err, workerData.room)
		return
	}

	u, err := url.Parse(getWS(workerData, key))
	if err != nil {
		fmt.Println(err, workerData.room)
		return
	}

	Dialer := *websocket.DefaultDialer

	if _, ok := conf.Proxy[workerData.Proxy]; ok {
		Dialer = websocket.Dialer{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http", // or "https" depending on your proxy
				Host:   conf.Proxy[workerData.Proxy],
				Path:   "/",
			}),
			HandshakeTimeout: 45 * time.Second, // https://pkg.go.dev/github.com/gorilla/websocket
		}
	}

	c, _, err := Dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error(), u.String(), workerData.room)
		return
	}

	defer c.Close()
	
	if err = c.WriteMessage(websocket.TextMessage, []byte(`["v3.authorize",{"token":"`+string(key)+`"}]`)); err != nil {
		fmt.Println(err.Error(), workerData.room)
		return
	}

	dons := make(map[string]struct{})
	
	ticker := time.NewTicker(60 * 60 * 8 * time.Second)
	defer ticker.Stop()
	
	var income int64
	income = 0

	for {
		
		select {
		case <-ticker.C:
			fmt.Println("too_long exit:", workerData.room)
			return
		case <-workerData.ch:
			fmt.Println("Exit room:", workerData.room)
			return
		default:
		}
		
		c.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err.Error(), workerData.room)
			if income > 1 && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				go reconnectRoom(workerData)
			}
			return
		}

		now := time.Now().Unix()
		slog <- saveLog{workerData.Rid, now, string(message)}
		
		if now > workerData.Last+60*60 {
			fmt.Println("no_tips exit:", workerData.room)
			return
		}
		
		
		messageID := getMessageID(string(message))
		
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
			if input.Stealth || len(input.Username) < 3 {
				input.Username = "anon_tips"
			}
			
			if _, ok := dons[input.Username]; !ok {
				dons[input.Username] = struct{}{}
				workerData.Dons++
			}
			
			save <- saveData{workerData.room, strings.ToLower(input.Username), workerData.Rid, input.Value, now}
			
			income += input.Value
			
			workerData.Tips++
			workerData.Last = now
			workerData.Income += input.Value
			rooms.Add <- workerData
			
			fmt.Println(input.Username, "send", input.Value, "tokens to", workerData.room)
		}
	}
}
