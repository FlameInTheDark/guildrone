# GuilDrone

This is a Guilded chat library based on [discordgo](https://github.com/bwmarrin/discordgo) by [bwmarrin](https://github.com/bwmarrin).

This is basically [bwmarrin's](https://github.com/bwmarrin) code, so please leave all thanks to him.

## Getting Started

### Installing

This assumes you already have a working Go environment, if not please see
[this page](https://golang.org/doc/install) first.

`go get` *will always pull the latest tagged release from the master branch.*

```sh
go get github.com/FlameInTheDark/guildrone
```

### Usage

Import the package into your project.

```go
import "github.com/FlameInTheDark/guildrone"
```

Construct a new Guilded client which can be used to access the variety of
Guilded API functions and to set callback functions for Guilded events.

```go
guilded, err := guildrone.New("authentication token")
```

See Documentation and Examples below for more detailed information.

- [![Go Reference](https://pkg.go.dev/badge/github.com/FlameInTheDark/guildrone.svg)](https://pkg.go.dev/github.com/FlameInTheDark/guildrone)
- [GuilDrone Examples](https://github.com/FlameInTheDark/guildrone/tree/master/examples) - A collection of example programs written with GuilDrone