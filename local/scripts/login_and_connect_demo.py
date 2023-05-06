
import requests
import socketio
from requests.auth import HTTPBasicAuth

sio = socketio.Client()

@sio.on("*")
def catch_all(event, data):
    print("EVENT", event)
    print("DATA", data)

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
    # login
    resp = requests.post("http://localhost:3000/login", auth=HTTPBasicAuth("script", "tpircs"))
    if resp.status_code != 200:
        print(resp.text)
        exit()
    # connect
    print("RESP COOKIE", resp.cookies.items())
    print("RESP HEADERS", resp.headers.items())
    uname = resp.cookies.get("UserName")
    pid = resp.cookies.get("PlayerId")
    sio_headers = {
        "Cookie": resp.headers.get("Set-Cookie")
    }
    sio.connect("http://localhost:3000", socketio_path="socket", namespaces="/", headers=sio_headers)
    print(sio.sid)

    sio.disconnect()


if __name__ == "__main__":
    demo()