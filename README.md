# Slack Playground

A playground to experiment with Slack apps. This repository contains a basic
Slack app that responds to "ping" messages when mentioned in a channel.

## Setup

Follow [these instructions](https://api.slack.com/authentication/basics) to
create a new app in your workspace. Configure the app as described below.

Scopes:

- `app_mentions:read`
- `chat:write`

Events:

- `app_mention`

The app exposes an event handler at `/event`. If you are using
[ngrok](https://ngrok.com/) to run the app locally, the request URL for events
will look like `https://0ebda1ca202b.ngrok.io/event`.

## Running

The app requires the following environment variables:

- `SLACK_ACCESS_TOKEN` - The access token for the app.
- `SLACK_SIGNING_SECRET` - The signing secret to verify incoming requests.

You can find both the access token and the signing secret in the app's basic
information page.

You can customize the address the app will listen to with the `-addr` flag. By
default, the app listens on the port 8080. If you are using
[ngrok](https://ngrok.com/) to run the app locally, remember to start the tunnel
with `ngrok http 8080` and setup the request URL for events accordingly.
