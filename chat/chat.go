package chat

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamtraining/chat/auth"
	"github.com/iamtraining/chat/logger"
)

const (
	socketReadBufferSize  = 1024
	socketWriteBufferSize = 1024
)

var (
	OUT = os.Stdout
)

type ChatServer struct {
	rooms      map[string]*Room
	roomsMu    sync.RWMutex
	wg         sync.WaitGroup
	bufferSize uint
	rootCtx    context.Context
	rootCancel context.CancelFunc
	upgrader   websocket.Upgrader
}

type Room struct {
	name    string
	send    chan *Message
	join    chan *Client
	leave   chan *Client
	clients map[*Client]bool
	cliMu   sync.RWMutex
	srvwg   *sync.WaitGroup
	rwg     *sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	logger  logger.Logger
}

type Client struct {
	socket *websocket.Conn
	data   string
	send   chan *Message
	room   *Room
}

type Message struct {
	Name    string
	Message string
	Time    time.Time
}

func NewChatServer(ctx context.Context, bufferSize uint) ChatServer {
	ctx, cancel := context.WithCancel(ctx)

	return ChatServer{
		rooms:      make(map[string]*Room),
		bufferSize: bufferSize,
		rootCtx:    ctx,
		rootCancel: cancel,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  socketReadBufferSize,
			WriteBufferSize: socketWriteBufferSize,
		},
	}
}

func (srv *ChatServer) Close(ctx context.Context) error {
	wg := make(chan struct{})
	go func() {
		srv.wg.Wait()
		wg <- struct{}{}
	}()

	srv.rootCancel()

	select {
	case <-wg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) read() {
	defer c.socket.Close()

	for {
		var msg *Message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}

		msg.Time = time.Now()
		msg.Name = c.data

		c.room.send <- msg
	}

}

func (c *Client) write() {
	defer c.socket.Close()

	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			break
		}
	}

}

func NewRoom(ctx context.Context, srvwg *sync.WaitGroup, name string, bufferSize uint) *Room {
	ctx, cancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	return &Room{
		name:    name,
		send:    make(chan *Message, bufferSize),
		join:    make(chan *Client),
		leave:   make(chan *Client),
		clients: make(map[*Client]bool),
		srvwg:   srvwg,
		rwg:     wg,
		ctx:     ctx,
		cancel:  cancel,
		logger:  logger.Silent(),
	}
}

func (r *Room) Run() {
	for {
		select {
		case cli := <-r.join:
			r.cliMu.RLock()
			r.logger.Log("user joined")
			r.clients[cli] = true
			r.cliMu.RUnlock()
		case cli := <-r.leave:
			r.cliMu.RLock()
			delete(r.clients, cli)
			close(cli.send)
			r.logger.Log("user left the room")
			r.cliMu.RUnlock()
		case msg := <-r.send:
			r.cliMu.RLock()
			r.logger.Log("message received ", "[\"", msg.Message, "\"]")
			for cli := range r.clients {
				cli.send <- msg
				r.logger.Log("[was sent to user]")
			}
			r.cliMu.RUnlock()
		case <-r.ctx.Done():
			r.rwg.Done()
		}
	}
}

func (r *Room) Close(ctx context.Context) error {
	wg := make(chan struct{})
	go func() {
		r.rwg.Wait()
		wg <- struct{}{}
	}()

	r.cancel()
	defer r.srvwg.Done()

	select {
	case <-wg:
		close(r.send)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (srv *ChatServer) Join(w http.ResponseWriter, r *http.Request) {
	name, ok := mux.Vars(r)["room"]
	if name == "" || !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.roomsMu.RLock()
	room, ok := srv.rooms[name]
	if !ok {
		room = NewRoom(srv.rootCtx, &srv.wg, name, srv.bufferSize)
		room.logger = logger.New(OUT)

		srv.rooms[name] = room

		srv.wg.Add(1)
		go room.Run()

		defer srv.wg.Done()
	}
	srv.roomsMu.RUnlock()

	socket, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("room", err)
	}

	cookie, err := r.Cookie("credentials")
	if err != nil {
		log.Fatal("getting credentials cookie failure", err.Error())
		return
	}

	cli := NewClient(socket, room, auth.Decoder(cookie.Value))

	defer func() {
		room.leave <- cli
	}()

	room.cliMu.RLock()
	room.join <- cli
	room.cliMu.RUnlock()

	room.rwg.Add(1)
	defer room.rwg.Done()

	go cli.write()

	cli.read()
}

func NewClient(socket *websocket.Conn, r *Room, email string) *Client {
	return &Client{
		socket: socket,
		data:   email,
		room:   r,
		send:   make(chan *Message),
	}
}
