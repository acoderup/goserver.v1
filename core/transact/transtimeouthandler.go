package transact

import (
	"github.com/acoderupacoderup/goserver.v1/core/timer"
)

type transactTimerAction struct {
}

func (t transactTimerAction) OnTimer(h timer.TimerHandle, ud interface{}) bool {
	if trans, ok := ud.(*TransNode); ok {
		trans.timeout()
		return true
	}
	return false
}
