package command

import (
	"fmt"
	"log/slog"

	"github.com/tidwall/resp"
)

const (
	commandLpush  = "LPUSH"
	commandRpush  = "RPUSH"
	commandLRange = "LRANGE"
)

type PushCommand struct {
	T     string
	Key   string
	Value string
}

type LrangeCommand struct {
	Key   string
	Start string
	End   string
}

func PushCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 3 {
		cmd := PushCommand{
			T:     set[0].String(),
			Key:   set[1].String(),
			Value: set[2].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid PUSH command")
	return nil, fmt.Errorf("invalid PUSH command")
}

func LrangeCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 4 {
		cmd := LrangeCommand{
			Key:   set[1].String(),
			Start: set[2].String(),
			End:   set[3].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid LRANGE command")
	return nil, fmt.Errorf("invalid LRANGE command")
}
