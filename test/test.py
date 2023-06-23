#!/usr/bin/env python3

import requests, json, websocket

def on_open(ws):
    r = requests.get("https://www.camsoda.com/api/v1/user/current")
    dist = json.loads(r.text)
    welcome = json.dumps(["v3.authorize", {"token": dist["user"]["node_token"]}])
    ws.send(welcome)
    print(welcome)

def on_message(ws, message):
    print("Received message: {}".format(message))
    dist = json.loads(message)
    print(dist[0])
    if dist[0] == 2:
        join = json.dumps([5, {"room": "irisallen"}])
        ws.send(join)
        print(join)

def on_error(ws, error):
    print("Error encountered: {}".format(error))

def on_close(ws, b, c):
    print(ws, b, c)
    print("WebSocket connection closed.")


ws = websocket.WebSocketApp("wss://node2-ord.livemediahost.com:3000", on_open=on_open, on_message=on_message, on_error=on_error, on_close=on_close)
ws.run_forever()  
