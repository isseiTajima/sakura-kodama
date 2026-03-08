package persona

import (
	"devcompanion/internal/types"
)

// CharacterCore は DevCompanion の基本人格（後輩）。
type CharacterCore struct {
	Name string
	Tone string
}

// PersonaEngine は Core と Style を組み合わせて最終的な表現を決定する。
type PersonaEngine struct {
	Core  CharacterCore
	Style types.PersonaStyle
}

func NewPersonaEngine(style types.PersonaStyle) *PersonaEngine {
	return &PersonaEngine{
		Core: CharacterCore{
			Name: "サクラ",
			Tone: "フレンドリーな後輩",
		},
		Style: style,
	}
}

// GetPromptModifiers はスタイルに応じたプロンプト指示を返す。
func (p *PersonaEngine) GetPromptModifiers() string {
	switch p.Style {
	case types.StyleSoft:
		return "優しく見守る、落ち着いたトーンで話してください。"
	case types.StyleEnergetic:
		return "元気いっぱいに、リアクション多めで話してください。"
	case types.StyleStrict:
		return "少しプレッシャーをかけるような、直接的な言い方で話してください。"
	default:
		return "丁寧かつ明るい態度で接してください。"
	}
}
