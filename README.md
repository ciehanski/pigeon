# pigeon [![Build Status](https://github.com/ciehanski/pigeon/workflows/build/badge.svg)](https://github.com/ciehanski/pigeon/actions) [![Coverage Status](https://coveralls.io/repos/github/ciehanski/pigeon/badge.svg?branch=master)](https://coveralls.io/github/ciehanski/pigeon?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/ciehanski/pigeon)](https://goreportcard.com/report/github.com/ciehanski/pigeon)

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

## Contributing:

Any pull request submitted must meet the following requirements:
- Have included tests applicable to the relevant PR.
- Attempt to adhere to the standard library as much as possible.

You can get started by either forking or cloning the repo. After, you can get started
by running:

```bash
make run
```

This will go ahead and build everything necessary to interface with Tor. After compose
has completed building, you will have a new `pigeon` container which will be your
dev environment.

Anytime a change to a .go or .mod file is detected the container will rerun with
the changes you have made. You must save in your IDE or text editor for the 
changes to be picked up. It takes roughly ~30 seconds for pigeon to restart after 
you have made changes.

You can completely restart the build with:
```bash
make restart
```

Run tests:
```bash
make exec
make test
```

Get container logs:
```bash
make logs
```

Shell into docker container:
```bash
make exec
```

Lint the project:
```bash
make lint
```

## License
- MIT