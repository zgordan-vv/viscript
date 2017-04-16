package process

import (
	"github.com/corpusc/viscript/hypervisor"
	extTask "github.com/corpusc/viscript/hypervisor/task_ext"
	"github.com/corpusc/viscript/msg"
	"strconv"
	"strings"
)

func (st *State) commandStart(args []string) {
	if len(args) < 1 {
		st.PrintError("Must pass a command into Start!")
	} else {
		detached := args[0] != "-a"

		if !detached {
			args = args[1:]
		}

		newExtProc, err := extTask.MakeNewTaskExternal(args, detached)
		if err != nil {
			st.PrintError(err.Error())
			return
		}

		err = newExtProc.Start()
		if err != nil {
			st.PrintError(err.Error())
			return
		}

		extProcInterface := newExtProc.GetExtProcessInterface()
		procId := hypervisor.AddExtProcess(extProcInterface)

		if !detached {
			st.proc.AttachExternalProcess(extProcInterface)
		}

		st.PrintLn("Added External Process (ID: " +
			strconv.Itoa(int(procId)) + ", Command: " +
			newExtProc.CommandLine + ")")
	}
}

func (st *State) commandAttach(args []string) {
	if len(args) < 1 {
		st.PrintError("No task id passed! eg: attach 1")
		return
	}

	passedId, err := strconv.Atoi(args[0])
	if err != nil {
		st.PrintError("Task id must be an integer.")
		return
	}

	extProcId := msg.ExtProcessId(passedId)

	extProc, err := hypervisor.GetExtProcess(extProcId)
	if err != nil {
		st.PrintError(err.Error())
		return
	}

	st.PrintLn(extProc.GetFullCommandLine())
	st.proc.AttachExternalProcess(extProc)
}

func (st *State) commandListExternalTasks(args []string) {
	extTaskMap := hypervisor.ExtProcessListGlobal.ProcessMap
	if len(extTaskMap) == 0 {
		st.PrintLn("No external tasks running.\n" +
			"Try starting one with \"start\" command (\"help\" or \"h\" for help).")
		return
	}

	fullPrint := false

	if len(args) > 0 && args[0] == "-f" {
		fullPrint = true
	}

	for procId, extProc := range extTaskMap {
		procCommand := ""

		if fullPrint {
			procCommand = extProc.GetFullCommandLine()
		} else {
			procCommand = strings.Split(
				extProc.GetFullCommandLine(), " ")[0]
		}

		st.Printf("[ %d ] -> [ %s ]\n", int(procId), procCommand)
	}
}
