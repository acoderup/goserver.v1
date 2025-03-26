package task

import (
	"github.com/acoderup/goserver.v1/core/basic"
	"github.com/acoderup/goserver.v1/core/utils"
)

type taskExeCommand struct {
	t Task
}

func (ttc *taskExeCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	defer utils.DumpStackIfPanic("taskExeCommand")
	ttc.t.setAfterQueCnt(o.GetPendingCommandCnt())
	return ttc.t.run(o)
}

// SendTaskExe 将任务发送给一个worker处理
func SendTaskExe(o *basic.Object, t Task) bool {
	t.setBeforeQueCnt(o.GetPendingCommandCnt())
	return o.SendCommand(&taskExeCommand{t: t}, true)
}
