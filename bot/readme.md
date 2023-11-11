# Panda Game Bot

The panda game bot is built to be a simple testing tool for multiplayer matches.

The bot is designed to connect to a specific game and play randomly.

## Configuration

the bot has 3 configurable variables:

| value | cli flag | env var | description | default |
| ---   | ---      | ---     | ---         | ---     |
| name  | `--name` | `BOT_NAME`| username of the bot | `bot-<randomnumber>` |
| host  | `--host` | `BOT_HOST` | hostnme of the panda game server | `localhost:3000` |
| game id | `--game-id` | `BOT_GAME_ID` | id of a current game for the bot to join | None |

The bot will not run without a provided `gameId`

cli flags have priority over env vars if both are set.

## Docker

run the `task bot-docker` command to build the bot image
