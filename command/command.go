package command

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"github.com/tidwall/resp"
)

const (
	commandSet   = "SET"
	commandGet   = "GET"
	commandDel   = "DEL"
	commandExist = "EXIST"
	commandIncr  = "INCR"
	commandDecr  = "DECR"
)

var commandsHandlers = map[string]func([]resp.Value) (Command, error){
	commandSet:   SetCommandHandler,
	commandGet:   GetCommandHandler,
	commandDel:   DelCommandHandler,
	commandExist: ExistCommandHandler,
	commandIncr:  IncrCommandHandler,
	commandDecr:  DecrCommandHandler,
}

type Command interface {
	// todo
}

func ParseRawMsg(msg string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(msg))
	for {
		msg, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("parse command err: %d", "err", err)
		}
		if msg.Type() == resp.Array {
			cmd, err := parseCommandType(msg)
			if err != nil {
				return nil, err
			}
			return cmd, nil
		}
	}
	return nil, fmt.Errorf("unknown command")
}

func parseCommandType(val resp.Value) (Command, error) {
	commandType := val.Array()[0].String()
	if handler, ok := commandsHandlers[commandType]; ok {
		return handler(val.Array())
	}
	return nil, fmt.Errorf("unknown command")
}
