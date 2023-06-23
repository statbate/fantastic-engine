#!/usr/bin/env python3

import requests, json, websocket

token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJlbWFpbF9jb25maXJtZWQiOmZhbHNlLCJnZW5kZXIiOiJnIiwiaGFzX2NjIjpmYWxzZSwiaGFzX3B1cmNoYXNlIjpmYWxzZSwiaGFzX3Rva2VucyI6ZmFsc2UsImhpZGVfdmlwX2JhZGdlIjpmYWxzZSwiaWF0IjoxNjg3NTI5OTMyLCJpc19ndWVzdCI6dHJ1ZSwiaXNfbW9kZWwiOmZhbHNlLCJpc19tb2RlbF9hZG1pbiI6ZmFsc2UsImlzX3NpdGVfbW9kIjpmYWxzZSwiaXNfc3VwZXJfdmlwIjpmYWxzZSwiaXNfdmlwIjpmYWxzZSwiaXNfYm90IjpmYWxzZSwiaXNzIjoiaHR0cDpcL1wvY2Ftc29kYS5jb20iLCJyZWNlbnRfcHVyY2hhc2VzIjpmYWxzZSwidG9rZW5zIjowLCJ1c2VybmFtZSI6Imd1ZXN0XzM2MTk0In0.Ono1bECiI5rQEVUNzeEZ-ISU6Nb5u2-JIe47iZbrzhU"

def on_open(ws):
    welcome = json.dumps(["v3.authorize", {"token": token}])
    ws.send(welcome)
    print(welcome)

def on_message(ws, message):
    print("Received message: {}".format(message))

       
def on_error(ws, error):
    print("Error encountered: {}".format(error))

def on_close(ws, b, c):
    print(ws, b, c)
    print("WebSocket connection closed.")


ws = websocket.WebSocketApp("wss://node2-ord.livemediahost.com:5870", on_open=on_open, on_message=on_message, on_error=on_error, on_close=on_close)
ws.run_forever()  
