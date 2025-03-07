package command

import (
	"fmt"
	"go-redis/pkg/utils"
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

type SetCommand struct {
	Key string
	Val string
	EX  string
}

type GetCommand struct {
	Key string
}

type DelCommand struct {
	Key string
}

type ExistCommand struct {
	Key string
}

type IncrCommand struct {
	Key    string
	Amount string
}
type DecrCommand struct {
	Key    string
	Amount string
}

func GetCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 2 {
		cmd := GetCommand{
			Key: set[1].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid GET command")
	return nil, fmt.Errorf("invalid GET command")
}

func IncrCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 2 {
		cmd := IncrCommand{
			Key: set[1].String(),
		}
		return cmd, nil
	} else if len(set) == 3 && utils.IsNumeric(set[2].String()) {
		cmd := IncrCommand{
			Key:    set[1].String(),
			Amount: set[2].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Incr command")
	return nil, fmt.Errorf("invalid Incr command")
}

func DecrCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 2 {
		cmd := DecrCommand{
			Key: set[1].String(),
		}
		return cmd, nil
	} else if len(set) == 3 && utils.IsNumeric(set[2].String()) {
		cmd := DecrCommand{
			Key:    set[1].String(),
			Amount: set[2].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Decr command")
	return nil, fmt.Errorf("invalid Decr command")
}

func ExistCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 2 {
		cmd := ExistCommand{
			Key: set[1].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid Exist command")
	return nil, fmt.Errorf("invalid Exist command")
}

func DelCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 2 {
		cmd := DelCommand{
			Key: set[1].String(),
		}
		return cmd, nil
	}
	slog.Error("invalid DEL command")
	return nil, fmt.Errorf("invalid DEL command")
}

func SetCommandHandler(set []resp.Value) (Command, error) {
	if len(set) == 3 {
		// normal set
		cmd := SetCommand{
			Key: set[1].String(),
			Val: set[2].String(),
		}
		return cmd, nil
	} else if len(set) == 5 {
		if set[3].String() == "EX" || set[3].String() == "PX" {
			// verify is valid or not
			ex, err := utils.MsOrS(set[3].String(), set[4].String())
			if err != nil {
				return nil, err
			}
			cmd := SetCommand{
				Key: set[1].String(),
				Val: set[2].String(),
				EX:  ex,
			}
			return cmd, nil
		}
	}

	slog.Error("invalid SET command")
	return nil, fmt.Errorf("invalid SET command")
}
