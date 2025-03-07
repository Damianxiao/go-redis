package server

import (
	"fmt"
	"go-redis/command"
	"go-redis/pkg/utils"
	"go-redis/repo"
	"log/slog"
	"net"
	"reflect"
	"strings"
)

const (
	listenAddress = ":50001"
	OK            = "+OK\r\n"
)

type Config struct {
	listenAddress string
}
type Server struct {
	config Config
	ln     net.Listener
	quitCh chan struct{}
	peerCh chan Conn
	peers  map[Conn]bool
	msgCh  chan Message
}

type Message struct {
	Conn Conn
	Data []byte
}

func NewMessage(conn Conn, data []byte) Message {
	return Message{
		Conn: conn,
		Data: data,
	}
}

func NewServer(conf Config) *Server {
	if conf.listenAddress == "" {
		conf.listenAddress = listenAddress
	}
	return &Server{
		quitCh: make(chan struct{}, 1),
		config: conf,
		peerCh: make(chan Conn),
		peers:  make(map[Conn]bool),
		msgCh:  make(chan Message),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.config.listenAddress)
	if err != nil {
		return fmt.Errorf("fail to listen: %v", err)
	}
	s.ln = ln
	go s.loop()
	s.acceptLoop()
	return nil
}

func (s *Server) loop() {
	for {
		select {
		case message := <-s.msgCh:
			if err := s.HandleRawMsg(message); err != nil {
				slog.Error("Error handling raw message", "err", err)
			}
		case peer := <-s.peerCh:
			s.peers[peer] = true
		case <-s.quitCh:
			return
		}
	}
}

func (s *Server) set(key, val string, ex string) error {
	err := repo.KvString.Set(key, val, ex)
	if err != nil {
		return err
	}
	slog.Info("save to memroy success", "data", key+val)
	return nil
}

func (s *Server) get(key string) ([]byte, error) {
	bytes, err := repo.KvString.Get(key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s *Server) del(key string) error {
	err := repo.KvString.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) exist(key string) (bool, error) {
	ok, err := repo.KvString.Exist(key)
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (s *Server) incr(key, amount string) error {
	err := repo.KvString.Incr(key, amount)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) decr(key, amount string) error {
	err := repo.KvString.Decr(key, amount)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) push(t, key, value string) {
	switch t {
	case "LPUSH":
		repo.KvList.Lpush(key, value)
	case "RPUSH":
		repo.KvList.Rpush(key, value)
	}
}

func (s *Server) lrange(key, start, end string) ([]string, error) {
	res, err := repo.KvList.Lrange(key, start, end)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) commandHandlers() map[reflect.Type]func(Message, command.Command) error {
	return map[reflect.Type]func(Message, command.Command) error{
		reflect.TypeOf(command.SetCommand{}):    s.executeSetCommand,
		reflect.TypeOf(command.GetCommand{}):    s.executeGetCommand,
		reflect.TypeOf(command.DelCommand{}):    s.executeDelCommand,
		reflect.TypeOf(command.ExistCommand{}):  s.executeExistCommand,
		reflect.TypeOf(command.IncrCommand{}):   s.executeIncrCommand,
		reflect.TypeOf(command.DecrCommand{}):   s.executeDecrCommand,
		reflect.TypeOf(command.PushCommand{}):   s.executePushCommand,
		reflect.TypeOf(command.LrangeCommand{}): s.executeLrangeCommand,
	}
}

func (s *Server) executeLrangeCommand(message Message, c command.Command) error {
	cmd := c.(command.LrangeCommand)
	res, err := s.lrange(cmd.Key, cmd.Start, cmd.End)
	if err != nil {
		slog.Error("lrange err", "err", err)
		s.handleErr(message, err)
		return err
	}
	data := strings.Join(res, " ")
	respClient(message.Conn, []byte(data), "data")
	slog.Info("range list", "list:", data)
	return nil
}

func (s *Server) executePushCommand(message Message, c command.Command) error {
	cmd := c.(command.PushCommand)
	s.push(cmd.T, cmd.Key, cmd.Value)
	s.handleSuccess(message, []byte(OK))
	slog.Info("PUSH command executed", "push type", cmd.T, "key", cmd.Key, "value", cmd.Value)
	// slog.Info("now memory list", "list", repo.KvList.KvList["user"])
	return nil
}

func (s *Server) executeSetCommand(message Message, c command.Command) error {
	cmd := c.(command.SetCommand)
	err := s.set(cmd.Key, cmd.Val, cmd.EX)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	s.handleSuccess(message, []byte(OK))
	slog.Info("SET command executed", "key", cmd.Key, "value", cmd.Val)
	return nil
}

func (s *Server) executeGetCommand(message Message, c command.Command) error {
	cmd := c.(command.GetCommand)
	bytes, err := s.get(cmd.Key)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	respClient(message.Conn, bytes, "data")
	return nil
}

func (s *Server) executeDelCommand(message Message, c command.Command) error {
	cmd := c.(command.DelCommand)
	err := s.del(cmd.Key)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	s.handleSuccess(message, []byte(OK))
	return nil
}

func (s *Server) executeExistCommand(message Message, c command.Command) error {
	cmd := c.(command.ExistCommand)
	ok, err := s.exist(cmd.Key)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	respClient(message.Conn, []byte(utils.Btoi(ok)), "data")
	return nil
}

func (s *Server) executeIncrCommand(message Message, c command.Command) error {
	cmd := c.(command.IncrCommand)
	err := s.incr(cmd.Key, cmd.Amount)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	s.handleSuccess(message, []byte(OK))
	return nil
}

func (s *Server) executeDecrCommand(message Message, c command.Command) error {
	cmd := c.(command.DecrCommand)
	err := s.decr(cmd.Key, cmd.Amount)
	if err := s.handleErr(message, err); err != nil {
		return err
	}
	s.handleSuccess(message, []byte(OK))
	return nil
}

func (s *Server) handleUnknownCommand(message Message) {
	respClient(message.Conn, []byte("Unknown command"), "err")
}

func (s *Server) executeCommand(message Message, cmd command.Command) error {
	handler := s.commandHandlers()
	if handler == nil {
		s.handleUnknownCommand(message)
		return fmt.Errorf("unknown command: %v", cmd)
	}
	return handler[reflect.TypeOf(cmd)](message, cmd)
}

func (s *Server) HandleRawMsg(message Message) error {
	cmd, err := command.ParseRawMsg(string(message.Data))
	if err != nil {
		respClient(message.Conn, []byte(err.Error()), "err")
		return err
	}
	return s.executeCommand(message, cmd)
}

func respClient(conn Conn, data []byte, t string) error {
	switch t {
	case "err":
		data = []byte("-ERR " + string(data) + "\r\n")
	case "ok":
		data = []byte(OK)
	case "data":
		data = []byte("$" + string(data) + "\r\n")
	}
	return conn.Write(data)
}

func (s *Server) handleErr(message Message, err error) error {
	if err != nil {
		respClient(message.Conn, []byte(err.Error()), "err")
		return err
	}
	return nil
}

func (s *Server) handleSuccess(message Message, resp []byte) {
	respClient(message.Conn, resp, "ok")
}

// listen new conn
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("Error accepting connection", "err", err)
			continue

		}
		go s.handleConn(conn)
	}
}
func (s *Server) handleConn(conn net.Conn) {
	peer := NewConn(conn, s.msgCh)
	s.peerCh <- peer
	slog.Info("new connected: ", "add:", peer.addr)
	if err := peer.read(); err != nil {
		slog.Error("fail to read msg ", "err", err)
	}
}
