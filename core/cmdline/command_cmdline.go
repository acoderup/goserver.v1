package cmdline

import (
	"github.com/acoderup/goserver.v1/core/basic"
)

type cmdlineCommand struct {
	exec cmdExecuter
	args []string
}

func (cmd *cmdlineCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	cmd.exec.Execute(cmd.args)
	return nil
}

func PostCmd(p *basic.Object, exec cmdExecuter, args []string) bool {
	return p.SendCommand(&cmdlineCommand{exec: exec, args: args}, true)
}
