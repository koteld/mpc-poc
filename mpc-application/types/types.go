package types

import "github.com/koteld/multi-party-sig/protocols/cmp"

type Config struct {
	Config    *cmp.Config `json:"config"`
	SessionID string      `json:"sessionID"`
}
