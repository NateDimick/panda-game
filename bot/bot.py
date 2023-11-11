import argparse
import json
import logging
import random
import requests
from logging import Logger
from os import getenv
from requests.auth import HTTPBasicAuth
from typing import Dict
from websockets.exceptions import ConnectionClosed
from websockets.sync.client import connect, ClientConnection

class PandaGameBot:
    def __init__(self, name: str, host: str, game_id: str):
        """
        """
        self.name: str = name
        self.host = host
        self.game_id = game_id
        self.disconnected: bool = False
        self.in_game = False # flag if bot is in a game (or a pre-game lobby)
        self.game_in_progress = False # will turn true once the game starts and false once game is over
        self.Wait_for_event = False # for client mode, will not prompt until next event
        self.socket: ClientConnection
        self.logger: Logger = logging.Logger(name=self.name)


    def connect(self):
        """
        login as guest with provided name and random password
        """
        login_resp = requests.post(f"http://{self.host}/guest", auth=HTTPBasicAuth(self.name, "todorandompassword"))
        if not login_resp.status_code == 200:
            self.logger.error({"log_reason": "guest_login_error", "resp": login_resp.text})
            exit()

        connect_headers = {
            "Cookie": login_resp.get("Set-Cookie")
        }
        self.socket = connect(f"ws://{self.host}/ws", headers=connect_headers)  # todo wss for real connections over the web

    def run(self):
        """
        Joins a specific game and plays randomly
        """
        self.connect()
        self.send_event("JoinGame", self.game_id)
        while True:
            if self.disconnected:
                return
            event = ""
            try:
                event = self.socket.recv()
                self.logger.info({"log_reason": "websocket_event", "raw_event": event})
            except ConnectionClosed:
                break
            ej = json.loads(event)
            event_name = ej["messageType"]
            event_data = ej["message"]
            self.logger.info({"log_reason": "receive_event", "event_name": event_name, "event_data": event_data})

            self.handle_message(event_name, event_data)
                    

    def handle_message(self, event_name: str, event_data: str | Dict):
        """
        Bot Plays Game Randomly
        """

        match event_name:
            case "GameStart":
                self.logger.info({"log_reason": "game_start", "game_id": event_data})
                self.game_in_progress = True
            case "GameOver":
                self.logger.info({"log_reason": "game_over", "game_id": event_data})
                self.game_in_progress = False
                self.socket.close()
                self.disconnected = True
            case "LobbyUpdate":
                self.logger.info({"log_reason": "lobby_update", "update": event_data})
                self.game_id = event_data["Gid"]
            case "GameUpdate":
                self.logger.info({"log_reason": "game_update", "update": event_data})
            case "ActionPrompt":
                choice = random.choice(event_data["selectFrom"])
                resp = {
                    "pid": event_data["pid"],
                    "action": event_data["action"],
                    "selection": choice
                }
                self.logger.info({"log_reason": "bot_TakeAction", "sent_event": resp})
                self.send_event("TakeAction", resp)
            case "Warning":
                self.logger.warn({"log_reason": "server_warning", "warning": event_data})
            case "Goodbye":
                self.logger.warn({"log_reason": "server_goodbye"})
                self.socket.close()
                self.disconnected = True


    def send_event(self, event_name: str, event_data: str):
        """
        Send event. Event Data should be json encoded string.
        """
        self.logger.info({"log_reason": "send_event", "event_name": event_name, "event_data": event_data})
        event = {
            "messageType": event_name,
            "message": event_data
        }
        self.socket.send(json.dumps(event))
        

if __name__ == "__main__":
    # load config from args or env
    ap = argparse.ArgumentParser()
    ap.add_argument("--name")
    ap.add_argument("--host")
    ap.add_argument("--game-id")

    args = ap.parse_args()

    env_name = getenv("BOT_NAME")
    env_host = getenv("BOT_HOST")
    env_game_id = getenv("BOT_GAME_ID")

    default_host = "localhost:3000"
    default_name = "bot-todogeneraterandomname"
    default_game_id = ""

    name = args["name"] or env_name or default_name
    host = args["host"] or env_host or default_host
    game_id = args["game-id"] or env_game_id or default_game_id

    if not game_id:
        print("game id is required")
        exit()

    # run the bot
    client = PandaGameBot(name, host, game_id)
    client.run()