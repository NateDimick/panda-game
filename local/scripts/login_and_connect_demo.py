
import requests
from requests.auth import HTTPBasicAuth
from websockets.sync.client import connect

def demo():
    # register
    body = "username=script&password=tpircs"
    headers = {
        "Content-Type": "application/x-www-form-urlencoded",
        "Content-Length": f"{len(body)}"
    }
    resp = requests.post("http://localhost:3000/register", data=body, headers=headers)
    if resp.status_code != 201 and resp.status_code != 409:
        print(resp.text)
        exit()
    # login as new user
    resp = requests.post("http://localhost:3000/login", auth=HTTPBasicAuth("script", "tpircs"))
    if resp.status_code != 200:
        print(resp.text)
        exit()
    # switch to pre-created user by create-mongo-db.py
    
    # connect
    print("RESP COOKIE", resp.cookies.items())
    print("RESP HEADERS", resp.headers.items())
    sio_headers = {
        "Cookie": resp.headers.get("Set-Cookie")
    }
    websocket = connect("ws://localhost:3000/ws", additional_headers=sio_headers)
    websocket.send("{\"messageType\": \"blah\", \"message\": \"foobar\"}")
    resp = websocket.recv()
    print(resp)
    


if __name__ == "__main__":
    demo()