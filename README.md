# redibot [![CircleCI](https://circleci.com/gh/orcaman/redibot.svg?style=svg)](https://circleci.com/gh/orcaman/redibot) [![GoDoc](https://godoc.org/github.com/orcaman/redibot?status.svg)](https://godoc.org/github.com/orcaman/redibot) 
  
`redibot` is a redis bot for Slack I hacked on the other day. Hopefully we could turn this into something useful.  

![Alt text](/img/Bot-05.png?raw=true "redibot")

Features:
- Pub-sub support: the neat thing about this is that you can basically start sending notifications into Slack from any of your servers via redis publishing. Subscribe to any channel from Slack and publish from any backend that has access to publish on redis.  
-  Collabarative: bring redis into the conversation! you can fetch, set, etc. from redis and discuss with your team mates.
- `redis-cli` in your Slack. No need to leave Slack in order to run redis commands. 

## usage

In order to use `redibot`, you need to run the `redibot` server and setup a new bot on your slack account. See *Adding a new bot to your slack* below. Once installed, use the [redis commands](http://redis.io/commands/) syntax same as you would use `redis-cli`. A few examples:

Connect to a redis server instance:
```
@redibot connect my.redis.host.com:6379 MY REDIS PASSWORD
```

Set a value on redis:
```
orcaman [12:19 PM]  
@redibot INCR counter

redibotBOT [12:19 PM]  
1
```

![Alt text](/img/Screen Shot 2016-10-06 at 21.10.04.png?raw=true "redibot")

You can also publish and subscribe! the neat thing about this is that you can basically start sending notifications from all your servers via redis publishing, without having to create a designated webhook on Slack.

```
@redibot subscribe my_channel
@redibot publish my_channel hello there!

---- output on slack -----
my_channel: message: hello there!
```

![Alt text](/img/Screen Shot 2016-10-06 at 21.15.16.png?raw=true "redibot")

### Adding a new bot to your slack

1. Browser over here: https://my.slack.com/services/new/bot
2. Create "redibot" username (or any other name that suites you)
3. Grab tha auth token - you need this to run the server


### Running the server 
It's easiest to run the server using docker:

```
docker run -e "redibot_token=YOUR SLACK BOT TOKEN" orcaman/redibot
```

Otherwise you can build and run the binary (see development).

## development

CD to the redibot directory, get the deps and go build:

```
go get github.com/garyburd/redigo/redis
go get github.com/gorilla/websocket
go build
```

To update the docker container after a new build:
```bash
docker build -t orcaman/redibot .
```