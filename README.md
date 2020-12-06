# discord-meetingmanager
> MeetingManager bot, for logging meetings in Discord voice channels

## Add to your server

[Click here](https://discord.com/api/oauth2/authorize?client_id=782730468156112957&permissions=8&scope=bot) to add the MeetingManager bot to your server.

## Functionality

The bot logs meeting times, durations, and members (along with how much time each member spent in the meeting.)

Meetings with only one member or which last less than 2 minutes will not be logged by default.

The bot will output to the "meeting-logs" text channel if one is located within the server where a meeting took place. Other channel names may also be used, eg. "voice-logs" or "meetings".

The bot will generate personal voice channels when a user joins a generator channel. To create a generator channel, make a voice channel named "Join/Click to Create Channel/Room/Meeting". It is not compatible with other bots which offer this functionality.

Send `ml!help` to see an informational panel.

## Run

Set environment variable DISCORD_BOT_TOKEN to your bot token.

Clone repo, then `go run`.

## Docker

A Dockerfile is included in this repository which can be used for deployment.
