package cmdline

import (
	"fmt"

	"github.com/acoderup/goserver.v1/core/module"
)

type exitExecuter struct {
}

func (this exitExecuter) Execute(args []string) {
	module.Stop()
}

func (this exitExecuter) ShowUsage() {
	fmt.Println("usage: exit")
}

func init() {
	RegisteCmd("exit", &exitExecuter{})
}
