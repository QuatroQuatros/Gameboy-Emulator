package gb

type Button byte

type ButtonInput struct {
	// Pressed and Released are inputs on this frame
	Pressed, Released []Button
}

type IOBinding interface {
	// RenderScreen renders a frame of the game.
	Render(screen *[160][144][3]uint8)

	// ButtonInput returns which buttons were pressed and released
	//ButtonInput() ButtonInput

	// SetTitle sets the title of the window.
	SetTitle(title string)
	// IsRunning returns if the monitor is still running.
	IsRunning() bool

	//DEBUG
	RenderMemory(gb *Gameboy)
}
