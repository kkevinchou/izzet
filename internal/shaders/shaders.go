package shaders

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/utils"
)

type ShaderProgram struct {
	ID uint32

	uniformCacheMat4  map[string]int32
	uniformCacheVec2  map[string]int32
	uniformCacheVec3  map[string]int32
	uniformCacheVec4  map[string]int32
	uniformCacheFloat map[string]int32
	uniformCacheInt   map[string]int32
	uniformCacheUInt  map[string]int32
}

const (
	vertexShaderExtension   = ".vs"
	fragmentShaderExtension = ".fs"
	geometryShaderExtension = ".gs"
)

type ShaderManager struct {
	vertexShaders   map[string]uint32
	fragmentShaders map[string]uint32
	geometryShaders map[string]uint32
	shaderPrograms  map[string]*ShaderProgram
}

func NewShaderManager(directory string) *ShaderManager {
	return &ShaderManager{
		vertexShaders:   loadShaders(directory, gl.VERTEX_SHADER, vertexShaderExtension),
		fragmentShaders: loadShaders(directory, gl.FRAGMENT_SHADER, fragmentShaderExtension),
		geometryShaders: loadShaders(directory, gl.GEOMETRY_SHADER, geometryShaderExtension),
		shaderPrograms:  map[string]*ShaderProgram{},
	}
}

func (s *ShaderManager) CompileShaderProgram(name, vertexShader, fragmentShader, geometryShader string) error {
	shaderProgram := gl.CreateProgram()

	if shader, ok := s.vertexShaders[vertexShader]; ok {
		gl.AttachShader(shaderProgram, shader)
	} else {
		return fmt.Errorf("vertex shader %s not found", vertexShader)
	}

	if shader, ok := s.fragmentShaders[fragmentShader]; ok {
		gl.AttachShader(shaderProgram, shader)
	} else {
		return fmt.Errorf("fragment shader %s not found", fragmentShader)
	}

	if geometryShader != "" {
		if shader, ok := s.geometryShaders[geometryShader]; ok {
			gl.AttachShader(shaderProgram, shader)
		} else {
			return fmt.Errorf("geometry shader %s not found", geometryShader)
		}
	}

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

	s.shaderPrograms[name] = &ShaderProgram{
		ID:                shaderProgram,
		uniformCacheMat4:  map[string]int32{},
		uniformCacheVec2:  map[string]int32{},
		uniformCacheVec3:  map[string]int32{},
		uniformCacheVec4:  map[string]int32{},
		uniformCacheFloat: map[string]int32{},
		uniformCacheInt:   map[string]int32{},
		uniformCacheUInt:  map[string]int32{},
	}

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
	shaderSource, err := os.ReadFile(shaderPath)
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
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheMat4[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheMat4[uniform] = u
		uniformLocation = u
	}

	gl.UniformMatrix4fv(uniformLocation, 1, false, &value[0])
}

func (s *ShaderProgram) SetUniformVec2(uniform string, value mgl32.Vec2) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheVec2[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheVec2[uniform] = u
		uniformLocation = u
	}
	gl.Uniform2fv(uniformLocation, 1, &value[0])
}

func (s *ShaderProgram) SetUniformVec3(uniform string, value mgl32.Vec3) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheVec3[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheVec3[uniform] = u
		uniformLocation = u
	}
	gl.Uniform3fv(uniformLocation, 1, &value[0])
}

func (s *ShaderProgram) SetUniformVec4(uniform string, value mgl32.Vec4) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheVec4[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheVec4[uniform] = u
		uniformLocation = u
	}
	gl.Uniform4fv(uniformLocation, 1, &value[0])
}

func (s *ShaderProgram) SetUniformInt(uniform string, value int32) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheInt[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheInt[uniform] = u
		uniformLocation = u
	}
	gl.Uniform1i(uniformLocation, value)
}

func (s *ShaderProgram) SetUniformUInt(uniform string, value uint32) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheUInt[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheUInt[uniform] = u
		uniformLocation = u
	}
	gl.Uniform1ui(uniformLocation, value)
}

func (s *ShaderProgram) SetUniformFloat(uniform string, value float32) {
	var ok bool
	var uniformLocation int32
	if uniformLocation, ok = s.uniformCacheFloat[uniform]; !ok {
		u := gl.GetUniformLocation(s.ID, gl.Str(fmt.Sprintf("%s\x00", uniform)))
		s.uniformCacheFloat[uniform] = u
		uniformLocation = u
	}
	gl.Uniform1f(uniformLocation, value)
}

func (s *ShaderProgram) Use() {
	gl.UseProgram(s.ID)
}
