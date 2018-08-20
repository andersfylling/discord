# Contributing

The following is a set of guidelines for contributing to Disgord.
These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

> **Note:** This CONTRIBUTIONS guideline is heavily inspired by the one created by the [Atom Organization](https://github.com/atom) on GitHub.

#### Table Of Contents

[Code of Conduct](#code-of-conduct)

[I don't want to read this whole thing, I just have a question!!!](#i-dont-want-to-read-this-whole-thing-i-just-have-a-question)

[What should I know before I get started?](#what-should-i-know-before-i-get-started)
  * [Introduction to Disgord](#introduction)
  * [Design Decisions](#design-decisions)

[How Can I Contribute?](#how-can-i-contribute)
  * [Reporting Bugs](#reporting-bugs)
  * [Suggesting Enhancements](#suggesting-enhancements)
  * [Your First Code Contribution](#your-first-code-contribution)
  * [Pull Requests](#pull-requests)

[Styleguides](#styleguides)
  * [Git Commit Messages](#git-commit-messages)

[Additional Notes](#additional-notes)
  * [Issue and Pull Request Labels](#issue-and-pull-request-labels)

## Code of Conduct
Use the GoLang formatter tool. Regarding commenting on REST functionality, do follow the commenting guide, which should include:
 1. The date it was reviewed/created/updated (just write reviewed)
 2. A description of what the endpoint does (copy paste from the discord docs if possible)
 3. The complete endpoint
 4. The rate limiter
 5. optionally, comments. (Comment#1, Comment#2, etc.)

Example (use spaces):
```GoLang
// CreateGuild [POST]       Create a new guild. Returns a guild object on success. Fires a Guild Create 
//                          Gateway event.
// Endpoint                 /guilds
// Rate limiter             /guilds
// Discord documentation    https://discordapp.com/developers/docs/resources/guild#create-guild
// Reviewed                 2018-08-16
// Comment                  This endpoint can be used only by bots in less than 10 guilds. Creating channel
//                          categories from this endpoint is not supported.
```

When creating a function, make sure that it has one purpose. It shouldn't hold any hidden functionality, and it could be wise to take dependencies as a parameter:
```GoLang
func CreateGuild(client httd.Poster, params *CreateGuildParams) (ret *Guild, err error) {
    ...
}
```

I won't accept pull requests where the author has created a singleton structure. I do not want package singletons either, as I'm worried it might cause technical debt. If you disagree you are welcome to create a discussion (not about the pattern, but why your implementation requires a singleton).

## I don't want to read this whole thing I just have a question!!!

> **Note:** While you are free to ask questions, given that you add the [help] prefix. You'll get faster results by using the resources below.

This repository has it's own discord guild/server: https://discord.gg/KjbVrak
Using the live chat application will most likely give you a faster result.

## What should I know before I get started?

### Introduction

## How Can I Contribute?

### Reporting Bugs
Reporting a bug should help the community improving Disgord. We need you to be specific and give enough information such that it can be reproduced by others. You must use the Bug template which can be found here: [TEMPLATE_BUG.md](TEMPLATE_BUG.md).

### Suggesting enhancement
We don't currently have a template for this. Provide benchmarks or demonstrations why your suggestion is an improvement or how it can help benefit this project is of great apriciation.


## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* When only changing documentation, include `[ci skip]` in the commit title
* Consider starting the commit message with an applicable emoji:
    * :art: `:art:` when improving the format/structure of the code
    * :racehorse: `:racehorse:` when improving performance
    * :non-potable_water: `:non-potable_water:` when plugging memory leaks
    * :memo: `:memo:` when writing docs
    * :penguin: `:penguin:` when fixing something on Linux
    * :apple: `:apple:` when fixing something on macOS
    * :checkered_flag: `:checkered_flag:` when fixing something on Windows
    * :bug: `:bug:` when fixing a bug
    * :fire: `:fire:` when removing code or files
    * :green_heart: `:green_heart:` when fixing the CI build
    * :white_check_mark: `:white_check_mark:` when adding tests
    * :lock: `:lock:` when dealing with security
    * :arrow_up: `:arrow_up:` when upgrading dependencies
    * :arrow_down: `:arrow_down:` when downgrading dependencies
    * :shirt: `:shirt:` when removing linter warnings
    

## Additional notes

Add the following prefixes to your issues to help categorize them better:
* [help] when asking a question about functionality.
