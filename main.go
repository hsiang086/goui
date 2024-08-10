package main

import (
	"runtime"

	engine "goui/engine"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	var game engine.Game
	if game.Initialize() {
		game.RunLoop()
	}
	game.Shutdown()
}
