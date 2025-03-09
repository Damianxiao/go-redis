package command

import (
	"fmt"
	"log/slog"

	"github.com/tidwall/resp"
)

const (
	commandZadd   = "ZADD"
	commandZscore = "ZSCORE"
	commandZrank   = "ZRANK"
)

type ZaddCommand struct {
	Key    string
	Member string
	Score  string
}

type ZscoreCommand struct {
	Key    string
	Member string
}

type ZrankCommand struct {
	Key    string
	Member string
}

func ZaddCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 4 {
		cmd := ZaddCommand{
			Key:    set[1].String(),
			Member: set[2].String(),
			Score:  set[3].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Zadd command")
	return nil, fmt.Errorf("invalid Zadd command")
}

func ZscoreCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 3 {
		cmd := ZscoreCommand{
			Key:    set[1].String(),
			Member: set[2].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Zscore command")
	return nil, fmt.Errorf("invalid Zscore command")
}

func ZrankCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 3 {
		cmd := ZrankCommand{
			Key:    set[1].String(),
			Member: set[2].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Zrank command")
	return nil, fmt.Errorf("invalid Zrank command")
}
