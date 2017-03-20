package viewport

import (
	"runtime"

	"github.com/corpusc/viscript/app"
	"github.com/corpusc/viscript/viewport/terminal"

	igl "github.com/corpusc/viscript/viewport/gl" //internal gl
)

//glfw
//glfw.PollEvents()
//only remaining

var (
	CloseWindow bool                   = false
	Terms       terminal.TerminalStack = terminal.TerminalStack{}
)

func Init() {
	println("viewport.Init()")
	app.MakeHighlyVisibleLogHeader(app.Name, 15)
	igl.InitCanvas()
	// GLFW event handling must run on the main OS thread
	// See documentation for functions that are only allowed to be called from the main thread.
	runtime.LockOSThread()

	initScreen()
	initEvents()
	initTerms()
}

func initScreen() {
	println("Viewport: init screen")
	igl.WindowInit()
	igl.LoadTextures()
	igl.InitRenderer()
}

func initEvents() {
	println("Viewport: init events")
	igl.InitInputEvents(igl.GlfwWindow)
	igl.InitMiscEvents(igl.GlfwWindow)
}

func initTerms() {
	println("Viewport: init terminals")
	Terms.Init()
	Terms.AddTerminal("Meshnet Server", "srv")
	Terms.AddTerminal("Meshnet Cli", "cli")
	// Terms.AddTerminal()
}

func TeardownScreen() {
	igl.ScreenTeardown()
}

func PollUiInputEvents() {
	igl.PollEvents() //move to gl
}

//could be in messages
func DispatchEvents() []byte {
	message := []byte{}

	for len(igl.InputEvents) > 0 {
		v := <-igl.InputEvents
		message = UnpackEvent(v)
	}

	return message
}

func Tick() {
	igl.Curs.Tick()
	Terms.Tick()
}

func UpdateDrawBuffer() {
	igl.DrawBegin()
	Terms.Draw()
	igl.DrawEnd()
}

func SwapDrawBuffer() {
	igl.SwapDrawBuffer()
}
