# hello-go

A very basic server-client application for learning Go.

## Usage

Running the application with no arguments will display a usage menu:


```
NAME:
   hello-go - a basic client-server application

USAGE:
   hello-go [global options] command [command options]

COMMANDS:
   server   Start a server
   client   Start a client
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --port PORT  PORT to serve or connect on (default: 3000)
   --help, -h   show help

```


### Server

Start a server with (example to change default port):

```
hello-go server -p 3333
```

Messages will be printed to the terminal when clients interact:

```
Server listening on: [::]:3333
Client connected: 127.0.0.1:53390
127.0.0.1:53390: hello, world
Client disconnected: 127.0.0.1:53390
```

### Client

Start a client (server must be running):

```
hello-go client -p 3333
```

Connecting with a client will enter a basic REPL for sending messages:

```
Connected to 127.0.0.1:3333
Waiting for input, use `QUIT` to exit
> hello, world
> QUIT
Goodbye!
```

## Learning Roadmap

The following are some goals to learn more about Go:

- [ ] Client features
    - [x] Username and authentication
    - [x] Broadcast or direct messaging
    - [ ] Slash commands (`/join`, `/leave`, etc.)
    - [ ] Direct messaging to peers
- [ ] Server features
    - [x] Database connections for user data
    - [x] Client connection map
    - [ ] Message history (ring buffer, database, etc.)
    - [ ] Server channels/rooms
    - [ ] Implement REST API (e.g., CRUD operations for users, server status, etc.)
- [ ] Language features
    - [ ] Unit tests
    - [x] Structured messages (JSON, binary encoding/decoding, etc.)
    - [ ] Interfaces for different types of databases
    - [x] Channels for message passing
    - [ ] Buffered streams for reading long messages
