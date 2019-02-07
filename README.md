# Disgord [![Documentation](https://godoc.org/github.com/andersfylling/disgord?status.svg)](http://godoc.org/github.com/andersfylling/disgord)
[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)[![forthebadge](https://forthebadge.com/images/badges/contains-technical-debt.svg)](https://forthebadge.com)[![forthebadge](https://forthebadge.com/images/badges/for-you.svg)](https://forthebadge.com)

## Health
| Branch       | Status  | Standard | Code coverage |
| ------------ |:-------------:|:---------------:|:----------------:|
| develop     | [![CircleCI](https://circleci.com/gh/andersfylling/disgord/tree/develop.svg?style=shield)](https://circleci.com/gh/andersfylling/disgord/tree/develop) | [![Go Report Card](https://goreportcard.com/badge/github.com/andersfylling/disgord)](https://goreportcard.com/report/github.com/andersfylling/disgord) | [![Test Coverage](https://api.codeclimate.com/v1/badges/687d02ca069eba704af9/test_coverage)](https://codeclimate.com/github/andersfylling/disgord/test_coverage) |

## About
GoLang module for interacting with the Discord API. Supports socketing and REST functionality. Discord object will also have implemented helper functions such as `Message.RespondString(session, "hello")`, or `Session.SaveToDiscord(&Emoji)` for simplicity/readability.

Disgord has complete implementation for Discord's documented REST API. It lacks comprehensive testing, hence the "contains technical debt" label, so any bug report/feedback is greatly appreciated!

To get started see the examples in [docs](docs/examples)

Discord channel/server: [Discord Gophers#Disgord](https://discord.gg/qBVmnq9)
You can find a live chats for DisGord in Discord. We exist in both the Gopher server and the Discord API server:
 - [Discord Gophers](https://discord.gg/qBVmnq9)
 - [Discord API](https://discord.gg/HBTHbme)

## Issues and behavior you must be aware of
Currently the caching focuses on being very configurable instead of as optimal as possible. This will change in the future, such that you will need to use build constraints if you want to tweak your cache configuration. But for now, or users that don't worry about this should know that all of your requests runs through the cache layer. You can overwrite the cache depending on REST method:
 - For those with a builder pattern, you can simply call `.IgnoreCache()` before you call `.Execute()`
 - The remaining methods will require you to use the exported package functions (see /docs/examples) for how to in detail.

The develop branch is under continuous breaking changes, so please use releases or be prepared to have a breaking codebase. A release branch will be introduced later when DisGord gets close to its v1.0.0 release. This is one of the reasons go modules is the only official supported way to use DisGord.

## Logging
DisGord allows you to inject your own logger, use the default one for DisGord (Zap), do not use any logging at all, you decide. To inject your own logger you must comply with the interface `disgord.Logger` (see logging.go).

## Package structure
None of the sub-packages should be used outside the library. If there exists a requirement for that, please create an issue or pull request.
```Markdown
github.com/andersfylling/disgord
└──.circleci    :CircleCI configuration
└──cache        :Different cache replacement algorithms
└──constant     :Constants such as version, GitHub URL, etc.
└──docs         :Examples, templates, (documentation)
└──logger       :Logger interface and Zap wrapper
└──endpoint     :All the REST endpoints of Discord
└──ratelimit    :All the ratelimit keys for the REST endpoints
└──event        :All the Discord event identifiers
└──generate     :All go generate scripts for "generic" code
└──httd         :Deals with rate limits and http calls
└──testdata     :Holds all test data for unit tests (typically JSON files)
└──websocket    :Discord Websocket logic (reconnect, resume, etc.)
```
The root pkg (disgord) holds all the data structures and the main client. Essentially all the features that should be used by the developer for creating bots. If you need access to the Snowflake type used by Disgord, then you should use `github.com/andersfylling/snowflake`.

### Dependencies
```Markdown
github.com/andersfylling/disgord
└──github.com/andersfylling/snowflake  :The snowflake ID designed for Discord
└──github.com/json-iterator/go         :For faster JSON decoding/encoding
└──github.com/sergi/go-diff            :Unit testing for checking JSON encoding/decoding of structs
└──github.com/uber-go/zap              :Logging (optional)
```

### Build constraints
> For **advanced users only**.

If you do not wish to use json-iterator, you can pass `-tags=json-std` to switch to `"encoding/json"`.
However, json-iterator is the recommended default for this library.

Disgord has the option to use mutexes (sync.RWMutex) on Discord objects. By default, methods of Discord objects are not locked as this is
not needed in our event driven architecture unless you create a parallel computing environment.
If you want the internal methods to deal with read-write locks on their own, you can pass `-tags=disgord_parallelism`, which will activate built-in locking.
Making all methods thread safe.

If you want to remove the extra memory used by mutexes, or you just want to completely avoid potential deadlocks by disabling
mutexes you can pass `-tags=disgord_removeDiscordMutex` which will replace the RWMutex with an empty struct, causing mutexes (in Discord objects only) to be removed at compile time.
This cannot be the default behaviour, as it creates confusion whether or not a mutex exists and leads to more error prone code. The developer has to be aware themselves whether or not
their code can be run without the need of mutexes. This option is not affected by the `disgord_parallelism` tag.

## Setup / installation guide
As this is a go module, it is expected that your project utilises the module concept (minimum Go version: 1.11). If you do not, then there is no guarantee that using this will work. To get this, simply use go get: `go get github.com/andersfylling/disgord`. I have been using this project in none module projects, so it might function for you as well. But official, this is not supported.

Read more about modules here: [https://github.com/golang/go/wiki/Modules](https://github.com/golang/go/wiki/Modules)

### Creating a fresh project using Disgord
To use the install script: `wget https://github.com/andersfylling/disgord/disgord.sh && ./disgord.sh` and follow the guide (please use go v1.11 minimum), or you can do it manually:

So if you haven't used modules before and you just want to create a Bot using Disgord, this is how it's done (Linux):
 1. Create a folder with your project name: `mkdir my-bot && cd my-bot` (outside the go path!)
 2. Create a main.go file, and add the following:
    ```go
    package main

    import "github.com/andersfylling/disgord"
    import "fmt"

    func main() {
        session, err := disgord.NewClient(&disgord.Config{
            BotToken: "DISGORD_TOKEN",
            Logger: disgord.DefaultLogger(false), // optional logging, debug=false
        })
        if err != nil {
            panic(err)
        }

        // note that the .Execute() is not on every REST method of DisGord (yet)
        myself, err := session.GetCurrentUser().Execute()
        if err != nil {
            panic(err)
        }

        fmt.Printf("Hello, %s!\n", myself.String())
    }
    ```
 3. Make sure you have activated go modules: `export GO111MODULE=auto`
 4. Initiate the project as a module: `go mod init my-bot` (you should now see a `go.mod` file)
 5. Start building, this will find all your dependencies and store them in the go.mod file: `go build .`
 6. You can now start the bot, and see the greeting: `go run .`

If you experience any issues with this guide, please create a issue.

## Contributing
Please see the [CONTRIBUTING.md file](CONTRIBUTING.md) (Note that it can be useful to read this regardless if you have the time)

## Git branching model
The branch:develop holds the most recent changes, as it name implies. There is no master branch as there will never be a "stable latest branch" except the git tags (or releases).

## Mental model
#### Caching
The cache can be either immutable (recommended) or mutable. When the cache is mutable you will share the memory space with the cache, such that if you change your data structure you might also change the cache directly. However, by using the immutable option all incoming data is deep copied to the cache and you will not be able to directly access the memory space, this should keep your code less error-prone and allow for concurrent cache access in case you want to use channels or other long-running tasks/processes.

#### Requests
For every REST API request the request is rate limited and cached auto-magically by DisGord. This means that when you utilize the Session interface you won't have to worry about rate limits and data is cached to improve performance. See the GoDoc for how to bypass the caching.

#### Events
> Note: if requested, a build flag can be added to use pro-actor pattern instead of reactor for handlers.

The reactor pattern is used. This will always be the default behavior, however channels will ofcourse work more as a pro-actor system as you deal with the data parallel to other functions.
Incoming events from the discord servers are parsed into respective structs and dispatched to either a) handlers, or b) through channels. Both are dispatched from the same place, and the arguments share the same memory space. Pick handlers (register them using Session.On method) simplicity as they run in sequence, while channels are executed in a parallel setting (it's expected you understand how channels work so I won't go in-depth here).

## Quick example
> **NOTE:** To see more examples go visit the docs/examples folder.
See the GoDoc for a in-depth introduction on the various topics (or disgord.go package comment). Below is an example of the traditional ping-pong bot and then some.
```go
package main 

import (
	"github.com/andersfylling/disgord"
	"os"
)

func replyPongToPing(session disgord.Session, data *disgord.MessageCreate) {
    msg := data.Message
    
    // whenever the message written is "ping", the bot replies "pong"
    if msg.Content == "ping" {
        msg.RespondString(session, "pong")
    }
}

func main() {
	var err error 
	
	botConfig := &disgord.Config{
        BotToken: os.Getenv("DISGORD_TOKEN"),
        Logger: disgord.DefaultLogger(false), // optional logging, debug=false
    }
	
    // create a Disgord session
    var client *disgord.Client
    if client, err = disgord.NewClient(botConfig); err != nil {
        panic(err)
    }
    
    // create a handler and bind it to new message events
    client.On(disgord.EventMessageCreate, replyPongToPing)
    
    // connect to the discord gateway to receive events
    if err = client.Connect(); err != nil {
        panic(err)
    }
    
    // Keep the socket connection alive, until you terminate the application (eg. Ctrl + C)
    if err = client.DisconnectOnInterrupt(); err != nil {
    	botConfig.Logger.Error(err) // reuse the logger from DisGord
    }
}
```

## Q&A

```Markdown
1. Is there an alternative Golang package?

Yes, it's called Discordgo (https://github.com/bwmarrin/discordgo). Its purpose is to provide low 
level bindings for Discord, while DisGord wants to provide a more configurable system with more 
features (channels, cache replacement strategies, build constraints, tailored unmarshal methods, etc.). 
Currently I do not have a comparison chart of DisGord and DiscordGo. But I do want to create one in the 
future, for now the biggest difference is that DisGord does not support self bots (as they aren't 
in the official documentation).
```

```Markdown
2. Reason for making another Discord lib in GoLang?

I'm trying to take over the world and then become a intergalactic war lord. Have to start somewhere.
```
