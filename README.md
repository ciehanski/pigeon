# pigeon [![Build Status](https://github.com/ciehanski/pigeon/workflows/build/badge.svg)](https://github.com/ciehanski/pigeon/actions) [![Build status](https://ci.appveyor.com/api/projects/status/c69cpo8syshw7xlj?svg=true)](https://ci.appveyor.com/project/ciehanski/pigeon) [![Coverage Status](https://coveralls.io/repos/github/ciehanski/pigeon/badge.svg?branch=master)](https://coveralls.io/github/ciehanski/pigeon?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ciehanski/pigeon)](https://goreportcard.com/report/github.com/ciehanski/pigeon)

pigeon is an instant messaging service built utilizing WebSockets 
and Tor hidden services as the transport mechanism. pigeon also
works without issue on Tor's strictest settings. The frontend is 
built with [Bulma](https://bulma.io/) and vanilla JavaScript. All 
JavaScript code is well-documented and can be found [here](https://github.com/ciehanski/pigeon/blob/master/templates/chatroom.go).

## Flags

#### Tor Version

Modify if pigeon will utilize Tor version 3. By default is true.

```bash
pigeon -torv3 false
```

#### Remote Port

Modify the port used to connect to the Tor hidden service. By
default is 80.

```bash
pigeon -port 8080
```

#### Debug

Runs pigeon in debug mode. If you want to stare at endless prompts
make sure to set this flag. If you're contributing this may come
in handy. :)

```bash
pigeon -debug
```

## License
- MIT