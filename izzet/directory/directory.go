package directory

import (
	"sync"
	"time"

	"github.com/kkevinchou/izzet/izzet/managers/item"
	"github.com/kkevinchou/izzet/izzet/managers/path"
	"github.com/kkevinchou/izzet/izzet/managers/player"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/font"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/shaders"
	"github.com/kkevinchou/kitolib/textures"
)

type IAssetManager interface {
	GetTexture(name string) *textures.Texture
	GetFont(name string) font.Font
	GetModel(name string) *modelspec.ModelSpecification
}

type IRenderSystem interface {
	Render(time.Duration)
}

type IShaderManager interface {
	CompileShaderProgram(name, vertexShader, fragmentShader string) error
	GetShaderProgram(name string) *shaders.ShaderProgram
}

type IPlayerManager interface {
	RegisterPlayer(playerID int, client types.NetworkClient)
	GetPlayer(id int) *player.Player
	GetPlayers() []*player.Player
}

type Directory struct {
	renderSystem  IRenderSystem
	assetManager  IAssetManager
	itemManager   *item.Manager
	pathManager   *path.Manager
	shaderManager IShaderManager
	playerManager IPlayerManager
}

var instance *Directory
var once sync.Once

func GetDirectory() *Directory {
	once.Do(func() {
		instance = &Directory{}
	})
	return instance
}

func (d *Directory) RegisterRenderSystem(system IRenderSystem) {
	d.renderSystem = system
}

func (d *Directory) RenderSystem() IRenderSystem {
	return d.renderSystem
}

func (d *Directory) RegisterAssetManager(manager IAssetManager) {
	d.assetManager = manager
}

func (d *Directory) AssetManager() IAssetManager {
	return d.assetManager
}

func (d *Directory) RegisterItemManager(manager *item.Manager) {
	d.itemManager = manager
}

func (d *Directory) ItemManager() *item.Manager {
	return d.itemManager
}

func (d *Directory) RegisterPathManager(manager *path.Manager) {
	d.pathManager = manager
}

func (d *Directory) PathManager() *path.Manager {
	return d.pathManager
}

func (d *Directory) RegisterShaderManager(manager IShaderManager) {
	d.shaderManager = manager
}

func (d *Directory) ShaderManager() IShaderManager {
	return d.shaderManager
}

func (d *Directory) RegisterPlayerManager(manager IPlayerManager) {
	d.playerManager = manager
}

func (d *Directory) PlayerManager() IPlayerManager {
	return d.playerManager
}
