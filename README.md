# Industry Job Notification Service

A service to notify you of completed industry jobs in EVE Online.

## What does it do?

IJNS will monitor your characters industry jobs and will send a message to you over Slack a minute before they are due to be completed.
This will give you time to log in so that you can deliver your jobs and begin new ones.

## How does it work?

IJNS uses the EVE XML API to look for jobs which are nearing completion, then uses the Slack Bot API to send you a direct message when it is nearing completion.

IJNS is only designed to work for POS manufacturing, and so requires a corp-level API key.

## Usage

IJNS ships with a Dockerfile which will build an image suitable for running with Docker.
To build the image simply run `docker build -t ijns .` in the project directory.
You can also use the author's public image, `robhaswell/ijns`.

Alternatively you can build the software if you have a suitable Go environment setup:

```
go get ./...
go build
```

You can then run the image directly with `docker run`.

IJNS requires configuration.
Please see below.

## Configuration

IJNS is configured via environment variables, and these must be present for the program to start.

| Variable | Description | Example
| --- | --- | --- |
| IJNS_KEYID | An EVE API key ID with the `IndustryJobs` permission. This must be a **Corporation** API key. [Click here to create one](https://community.eveonline.com/support/api-key/CreatePredefined?accessMask=128) | 1234567 |
| IJNS_VCODE | An EVE API verification code. See above. | abcdefghijklmnopqrtsuvwxyz1234567890 |
| IJNS_SLACK_TOKEN | A token for the Slack Bot API. [Click here to create one](https://my.slack.com/services/new/bot) | abcd-123456789-abcdefghhijk1234567890 |
| IJNS_CHARACTERS | A JSON array which maps Slack userames (to be notified) onto a list of EVE Character names. The Slack username will be notified about those characters' jobs. | `{"slack_username":["EVE Character 1", "EVE Character 2"], "slack_other_username":["EVE Character 3"]}` |

# Bug reports & contributing

Please use [GitHub Issues](https://github.com/robhaswell/ijns/issues) to report bugs or feature requests.
Pull requests are welcome.

CI is configured using the Docker Hub: https://hub.docker.com/r/robhaswell/ijns/builds/
