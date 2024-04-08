package mailwatcher

import (
	"encoding/json"
	"errors"
)

type Action int32

const (
	Add             Action = 1
	Remove          Action = 2
	Watch           Action = 3
	WatchAll        Action = 4
	Stop            Action = 5
	StopAll         Action = 6
	Code            Action = 7
	ConnectionError Action = 8
)

type Message struct {
	Cmd    Action
	Params map[string]interface{}
}

func (a *Action) ToString() (string, error) {
	switch *a {
	case Add:
		return "Add", nil
	case Remove:
		return "Remove", nil
	case Watch:
		return "Watch", nil
	case WatchAll:
		return "WatchAll", nil
	case Stop:
		return "Stop", nil
	case StopAll:
		return "StopAll", nil
	case Code:
		return "Code", nil
	case ConnectionError:
		return "ConnectionError", nil
	default:
		return "", errors.New("Unknown message action")
	}
}

func Parse(msg []byte) (Message, error) {
	m := Message{}
	err := json.Unmarshal(msg, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func Serialize(msg *Message) ([]byte, error) {
	return json.Marshal(*msg)
}
