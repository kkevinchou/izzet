package shaders

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	utils "github.com/kkevinchou/izzet/lib/libutils"
)

type ShaderProgram struct {
	ID uint32
}

const (
	vertexShaderExtension   = ".vs"
	fragmentShaderExtension = ".fs"
)

type ShaderManager struct {
	vertexShaders   map[string]uint32
	fragmentShaders map[string]uint32
	shaderPrograms  map[string]*ShaderProgram
}

func NewShaderManager(directory string) *ShaderManager {
	shaderManager := ShaderManager{
		vertexShaders:   loadShaders(directory, gl.VERTEX_SHADER, vertexShaderExtension),
		fragmentShaders: loadShaders(directory, gl.FRAGMENT_SHADER, fragmentShaderExtension),
		shaderPrograms:  map[string]*ShaderProgram{},
	}

	return &shaderManager
}

func (s *ShaderManager) CompileShaderProgram(name, vertexShader, fragmentShader string) error {
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, s.vertexShaders[vertexShader])
	gl.AttachShader(shaderProgram, s.fragmentShaders[fragmentShader])
	gl.LinkProgram(shaderProgram)

	var status int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		return fmt.Errorf("failed to link shader program:\n%s", log)
	}

	// gl.DeleteShader(vertexShader)
	// gl.DeleteShader(fragmentShader)

	s.shaderPrograms[name] = &ShaderProgram{ID: shaderProgram}

	return nil
}

func (s *ShaderManager) GetShaderProgram(name string) *ShaderProgram {
	return s.shaderPrograms[name]
}

func loadShaders(directory string, shaderType uint32, extension string) map[string]uint32 {
	extensions := map[string]any{
		extension: nil,
	}

	shaderMap := map[string]uint32{}
	fileMetaData := utils.GetFileMetaData(directory, nil, extensions)

	for _, metaData := range fileMetaData {
		shader, err := compileShader(metaData.Path, shaderType)
		if err != nil {
			panic(err)
		}
		shaderMap[metaData.Name] = shader
	}

	return shaderMap
}

func compileShader(shaderPath string, shaderType uint32) (uint32, error) {
	shaderSource, err := ioutil.ReadFile(shaderPath)
	if err != nil {
		return 0, err
	}

	shader := gl.CreateShader(shaderType)
	csource, free := gl.Strs(string(shaderSource) + "\x00")
	gl.ShaderSource(shader, 1, csource, nil)
	free()

	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile shader (%s):\n%s", shaderPath, log)
	}

	return shader, nil
}

func (s *ShaderProgram) SetUniformMat4(uniform string, value mgl32.Mat4) {
	uniformLocation := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
	gl.UniformMatrix4fv(uniformLocation, 1, false, &value[0])
}

func (s *ShaderProgram) SetUniformVec3(uniform string, value mgl32.Vec3) {
	floats := []float32{value.X(), value.Y(), value.Z()}
	uniformLocation := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
	gl.Uniform3fv(uniformLocation, 1, &floats[0])
}

func (s *ShaderProgram) SetUniformVec4(uniform string, value mgl32.Vec4) {
	floats := []float32{value.X(), value.Y(), value.Z(), value.W()}
	uniformLocation := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
	gl.Uniform4fv(uniformLocation, 1, &floats[0])
}

func (s *ShaderProgram) SetUniformInt(uniform string, value int32) {
	uniformLocation := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
	gl.Uniform1i(uniformLocation, value)
}

func (s *ShaderProgram) SetUniformFloat(uniform string, value float32) {
	uniformLocation := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
	gl.Uniform1f(uniformLocation, value)
}

func (s *ShaderProgram) Use() {
	gl.UseProgram(s.ID)
}
