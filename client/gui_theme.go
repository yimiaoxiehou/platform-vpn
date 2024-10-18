package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type fghGuiTheme struct {
}

func (f *fghGuiTheme) Font(s fyne.TextStyle) fyne.Resource {
	font, err := assetsFs.ReadFile("assets/" + GuiFontName)
	if err != nil {
		return theme.DefaultTheme().Font(s)
	}
	return fyne.NewStaticResource("fgh-font", font)
}

func (*fghGuiTheme) Color(c fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch c {
	case theme.ColorNamePrimary, theme.ColorNameButton:
		// 主要颜色和按钮颜色
		// 使用深青色 (#008080)，比原来的颜色稍深，以在亮色背景上保持足够的对比度
		return color.RGBA{R: 0x00, G: 0x80, B: 0x80, A: 0xff}
	case theme.ColorNameBackground:
		// 背景颜色
		// 使用非常浅的灰色 (#F5F5F5)，接近白色但不刺眼
		return color.RGBA{R: 0xF5, G: 0xF5, B: 0xF5, A: 0xff}
	case theme.ColorNameMenuBackground, theme.ColorNameInputBackground, theme.ColorNameOverlayBackground:
		// 菜单背景、输入框背景和覆盖层背景颜色
		// 使用白色 (#FFFFFF)，与主背景形成微妙区别
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xff}
	case theme.ColorNameDisabledButton:
		// 禁用状态的按钮颜色
		// 使用浅灰色 (#CCCCCC)
		return color.RGBA{R: 0xCC, G: 0xCC, B: 0xCC, A: 0xff}
	case theme.ColorNameDisabled:
		// 禁用状态的一般元素颜色
		// 使用中等灰色 (#999999)
		return color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xff}
	case theme.ColorNameForeground:
		// 前景色（主要用于文本）
		// 使用深灰色 (#333333)，在浅色背景上提供良好的可读性
		return color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff}
	case theme.ColorNameHover:
		// 悬停状态颜色
		// 使用浅青色 (#E0F0F0)，与主题色调和
		return color.RGBA{R: 0xE0, G: 0xF0, B: 0xF0, A: 0xff}
	case theme.ColorNameSelection:
		// 选中状态颜色
		// 使用浅青绿色 (#B0E0E0)
		return color.RGBA{R: 0xB0, G: 0xE0, B: 0xE0, A: 0xff}
	default:
		// 对于未特别指定的颜色，使用默认亮色主题的颜色
		return theme.DefaultTheme().Color(c, theme.VariantLight)
	}
}

func (*fghGuiTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (*fghGuiTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}
