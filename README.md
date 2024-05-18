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
   server        Start a server
   client        Start a client
   database, db  manage the application database
   help, h       Shows a list of commands or help for one command

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
Server listening on [::]:3000
ACCEPT guest user `guest94509` (127.0.0.1:53294)
Client connected: guest94509@127.0.0.1:53294
REJECT invalid credentials (127.0.0.1:53301)
ACCEPT authenticated user `alice` (127.0.0.1:53306)
Client connected: alice@127.0.0.1:53306
alice@127.0.0.1:53306: Packet(type=3, body='hello')
guest94509@127.0.0.1:53294: Packet(type=2, body='hello everyone!')
Client disconnected: guest94509@127.0.0.1:53294
```

### Client

Start a client (server must be running):

```
hello-go client -p 3333
```

Connecting with a client will enter a basic REPL for sending messages:

```
Connected to 127.0.0.1:3000
SERVER READY
!hello everyone!
BROADCAST> guest94509@127.0.0.1:53294: hello everyone!
QUIT
Goodbye!
```

By default, `client` will use `guest` as the username with no password. The server will generate a random username (`guestXXXXX`) for guests.

To authenticate, use `-u` and `-p` to provide a username and password.

## Learning Roadmap

The following are some goals to learn more about Go:

- [ ] Client features
    - [x] Broadcast messages (`!message`)
    - [ ] Slash commands (`/list`, `/join`, `/leave`, etc.)
    - [ ] Direct messages to peers or channels/rooms
- [ ] Server features
    - [x] Database connections
    - [x] Client connection map
    - [x] Permissions for authenticated users vs. guests
    - [ ] Server channels/rooms
- [ ] Application features
    - [x] User authentication
    - [ ] Websockets API
    - [ ] REST API
- [ ] Language/misc features
    - [ ] Buffered streams for reading long messages
    - [x] Structured messages (JSON, binary encoding/decoding, etc.)
    - [ ] Parsing/routing different message types (e.g., tagged union with `json.RawMessage`)
    - [x] Channels for message passing
    - [x] Password hashing
    - [ ] Interfaces for various items (messages, database connections, etc.)
    - [x] Color output
    - [ ] Unit tests

References:

- [Anthony GG | How To Build A Custom TCP Server In Golang?](https://www.youtube.com/watch?v=qJQrrscB1-4)
- [david hwang | Go Realtime Chat Part 1](https://www.youtube.com/watch?v=W9SuX9c40s8)
- [ProgrammingPercy | Mastering WebSockets With Go](https://www.youtube.com/watch?v=pKpKv9MKN-E)
