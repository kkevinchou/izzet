package contentbrowser

import "github.com/inkyblackness/imgui-go/v4"

type ContentBrowser struct {
	Items []ContentItem
}

type ContentItem struct {
	Texture imgui.TextureID
	Name    string
}
