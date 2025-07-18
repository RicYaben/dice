package sdk

import (
	"encoding/json"

	"github.com/dice/shared"
)

func AsJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

type Labels map[string]shared.Label

func (l Labels) MakeLabel(name string, hostID uint) *shared.Label {
	lab := l[name]
	lab.HostID = hostID
	return &lab
}
