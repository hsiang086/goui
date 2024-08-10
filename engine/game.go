package engine

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mat/besticon/ico"
)

var (
	IsRunning    = false
	IsFullscreen = false
	WindowWidth  = 1024
	WindowHeight = 768
	cameraPos    = mgl32.Vec3{0.0, 0.0, 3.0}
	cameraFront  = mgl32.Vec3{0.0, 0.0, -1.0}
	cameraUp     = mgl32.Vec3{0.0, 1.0, 0.0}
	projection   mgl32.Mat4
)

const (
	camerSpeed = 0.05
)

type Game struct {
	MWindow *glfw.Window
	VAO     uint32
	VBO     uint32
	Program uint32
}

var triangleVertices = []float32{
	// Positions     // Colors
	0.0, 0.5, -1.0, 1.0, 0.0, 0.0, // Top vertex (red)
	-0.5, -0.5, 0.0, 0.0, 1.0, 0.0, // Bottom left vertex (green)
	0.5, -0.5, 0.0, 0.0, 0.0, 1.0, // Bottom right vertex (blue)
}

func (g *Game) Initialize() bool {
	fmt.Println(glfw.GetVersionString())

	var err error

	if err = glfw.Init(); err != nil {
		fmt.Printf("Failed to initialize glfw: %v", err)
		return false
	}

	glfw.WindowHint(glfw.RedBits, 8)
	glfw.WindowHint(glfw.GreenBits, 8)
	glfw.WindowHint(glfw.BlueBits, 8)
	glfw.WindowHint(glfw.AlphaBits, 8)
	glfw.WindowHint(glfw.DepthBits, 24)
	glfw.WindowHint(glfw.DoubleBuffer, 1)

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True) // FOR MAC

	if g.MWindow, err = glfw.CreateWindow(WindowWidth, WindowHeight, "GOUI", nil, nil); err != nil {
		fmt.Printf("Failed to create window: %v", err)
		return false
	}

	cwd := os.Getenv("PWD")

	if ok, _ := PathExists(cwd + "/icon/goui.ico"); ok {
		if icon, err := os.Open(cwd + "/icon/goui.ico"); err == nil {
			defer icon.Close()
			if img, err := ico.Decode(icon); err == nil {
				g.MWindow.SetIcon([]image.Image{img})
			}
		}
	}

	g.MWindow.SetCloseCallback(CloseCallback)
	g.MWindow.SetSizeCallback(SizeCallback)
	g.MWindow.SetKeyCallback(KeyCallback)
	g.MWindow.SetMouseButtonCallback(MouseButtonCallback)
	g.MWindow.SetCursorPosCallback(CursorPosCallback)
	g.MWindow.SetCursorEnterCallback(CursorEnterCallback)
	g.MWindow.SetScrollCallback(ScrollCallback)

	g.MWindow.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		fmt.Printf("Failed to initialize OpenGL: %v", err)
		return false
	}
	projection = mgl32.Perspective(mgl32.DegToRad(45.0), float32(WindowWidth)/float32(WindowHeight), 0.1, 100.0)

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)

	g.initializeOpenGL()

	IsRunning = true
	return true
}

func (g *Game) initializeOpenGL() {
	// Compile shaders and link them into a program
	vertexShaderSource := `
		#version 330 core
		layout(location = 0) in vec3 position;
		layout(location = 1) in vec3 color;
		out vec3 fragColor;

		uniform mat4 model;
		uniform mat4 view;
		uniform mat4 projection;

		void main() {
			gl_Position = projection * view * model * vec4(position, 1.0);
			fragColor = color;
		}
	` + "\x00"

	fragmentShaderSource := `
		#version 330 core
		in vec3 fragColor;
		out vec4 color;
		void main() {
			color = vec4(fragColor, 1.0);
		}
	` + "\x00"

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)

	// Convert Go strings to C strings and ensure null-termination
	vertexShaderSourceCString := gl.Str(vertexShaderSource)
	fragmentShaderSourceCString := gl.Str(fragmentShaderSource)

	// Compile vertex shader
	gl.ShaderSource(vertexShader, 1, &vertexShaderSourceCString, nil)
	gl.CompileShader(vertexShader)
	if !checkShaderCompileErrors(vertexShader, "VERTEX") {
		return
	}

	// Compile fragment shader
	gl.ShaderSource(fragmentShader, 1, &fragmentShaderSourceCString, nil)
	gl.CompileShader(fragmentShader)
	if !checkShaderCompileErrors(fragmentShader, "FRAGMENT") {
		return
	}

	// Link shaders into a program
	g.Program = gl.CreateProgram()
	gl.AttachShader(g.Program, vertexShader)
	gl.AttachShader(g.Program, fragmentShader)
	gl.LinkProgram(g.Program)
	if !checkProgramLinkingErrors(g.Program) {
		return
	}

	// Cleanup shaders
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Create VAO and VBO
	gl.GenVertexArrays(1, &g.VAO)
	gl.GenBuffers(1, &g.VBO)

	gl.BindVertexArray(g.VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, g.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangleVertices)*4, gl.Ptr(triangleVertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.Ptr(nil))
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.Ptr(uintptr(3*4)))
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

// func readShaderSource(path string) string {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer file.Close()

// 	stat, err := file.Stat()
// 	if err != nil {
// 		panic(err)
// 	}

// 	data := make([]byte, stat.Size())
// 	_, err = file.Read(data)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return string(data) + string('\x00')
// }

func checkShaderCompileErrors(shader uint32, shaderType string) bool {
	var success int32
	var infoLog [512]byte
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		gl.GetShaderInfoLog(shader, 512, nil, &infoLog[0])
		fmt.Printf("Error compiling %s shader: %s\n", shaderType, string(infoLog[:]))
		return false
	}
	return true
}

func checkProgramLinkingErrors(program uint32) bool {
	var success int32
	var infoLog [512]byte
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		gl.GetProgramInfoLog(program, 512, nil, &infoLog[0])
		fmt.Printf("Error linking program: %s\n", string(infoLog[:]))
		return false
	}
	return true
}

func (g *Game) RunLoop() {
	for IsRunning {
		// clear the screen
		gl.ClearColor(0.25, 0.25, 0.25, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		g.ProcessInput()
		g.Update()
		g.GenerateOutput()

		g.MWindow.SwapBuffers()
	}
}

func (g *Game) Shutdown() {
	g.MWindow.Destroy()
	glfw.Terminate()
}

func (g *Game) ProcessInput() {
	glfw.PollEvents()
}

func (g *Game) Update() {
	// Calculate the view matrix
	view := mgl32.LookAtV(cameraPos, cameraPos.Add(cameraFront), cameraUp)

	// Set the model, view, and projection uniforms in the shader
	model := mgl32.Ident4()
	viewUniform := gl.GetUniformLocation(g.Program, gl.Str("view\x00"))
	gl.UniformMatrix4fv(viewUniform, 1, false, &view[0])

	modelUniform := gl.GetUniformLocation(g.Program, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

	projectionUniform := gl.GetUniformLocation(g.Program, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

	// Render the scene
	g.GenerateOutput()
}

func (g *Game) GenerateOutput() {
	gl.UseProgram(g.Program)
	gl.BindVertexArray(g.VAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
	gl.BindVertexArray(0)
}

func CloseCallback(w *glfw.Window) {
	IsRunning = false
}

func SizeCallback(w *glfw.Window, width int, height int) {
	fmt.Printf("Size: %v %v\n", width, height)
}

func KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		if action == glfw.Press {
			w.SetShouldClose(true)
		}
	case glfw.KeyF11:
		if action == glfw.Press {
			IsFullscreen = !IsFullscreen
			if IsFullscreen {
				w.SetMonitor(
					glfw.GetPrimaryMonitor(),
					0,
					0,
					glfw.GetPrimaryMonitor().GetVideoMode().Width,
					glfw.GetPrimaryMonitor().GetVideoMode().Height,
					60,
				)
			} else {
				posX := (glfw.GetPrimaryMonitor().GetVideoMode().Width - WindowWidth) / 2
				posY := (glfw.GetPrimaryMonitor().GetVideoMode().Height - WindowHeight) / 2
				w.SetMonitor(nil, posX, posY, WindowWidth, WindowHeight, 60)
			}
		}
	case glfw.KeyW:
		cameraPos = cameraPos.Add(cameraFront.Mul(camerSpeed))
	case glfw.KeyS:
		cameraPos = cameraPos.Sub(cameraFront.Mul(camerSpeed))
	case glfw.KeyA:
		cameraPos = cameraPos.Sub(cameraFront.Cross(cameraUp).Normalize().Mul(camerSpeed))
	case glfw.KeyD:
		cameraPos = cameraPos.Add(cameraFront.Cross(cameraUp).Normalize().Mul(camerSpeed))
	}
}

func MouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	fmt.Printf("Mouse button: %v %v\n", button, action)
}

func CursorPosCallback(w *glfw.Window, xpos float64, ypos float64) {
	fmt.Printf("Cursor position: %v %v\n", xpos, ypos)
}

func CursorEnterCallback(w *glfw.Window, entered bool) {
	fmt.Printf("Cursor entered: %v\n", entered)
}

func ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	fmt.Printf("Scroll: %v %v\n", xoff, yoff)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
