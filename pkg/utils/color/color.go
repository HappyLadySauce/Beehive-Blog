package color

import (
	"regexp"
	"strings"
)

var hexColor = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// ValidHex 校验是否为 #RGB 或 #RRGGBB。
func ValidHex(s string) bool {
	return hexColor.MatchString(s)
}

var bareHex = regexp.MustCompile(`^(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// NormalizeToHex 将输入规范为 #RRGGBB：支持 #RGB/#RRGGBB、无 # 前缀的十六进制，以及常见 CSS 颜色名（忽略大小写与部分空格）。
// 无法识别时返回空字符串。
func NormalizeToHex(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if ValidHex(s) {
		return expandShortHex(strings.ToUpper(s))
	}
	if bareHex.MatchString(s) {
		return expandShortHex("#" + strings.ToUpper(s))
	}
	key := strings.ToLower(strings.ReplaceAll(s, " ", ""))
	if h, ok := namedColors[key]; ok {
		return h
	}
	return ""
}

func expandShortHex(s string) string {
	if len(s) != 4 { // not #RGB
		return s
	}
	r, g, b := s[1], s[2], s[3]
	return "#" + string([]byte{r, r, g, g, b, b})
}

// namedColors 常见颜色名 -> #RRGGBB（与前端主题展示兼容）。
var namedColors = map[string]string{
	"black":                "#000000",
	"white":                "#FFFFFF",
	"red":                  "#FF0000",
	"green":                "#008000",
	"blue":                 "#0000FF",
	"yellow":               "#FFFF00",
	"cyan":                 "#00FFFF",
	"magenta":              "#FF00FF",
	"gray":                 "#808080",
	"grey":                 "#808080",
	"orange":               "#FFA500",
	"purple":               "#800080",
	"violet":               "#EE82EE",
	"pink":                 "#FFC0CB",
	"brown":                "#A52A2A",
	"navy":                 "#000080",
	"teal":                 "#008080",
	"lime":                 "#00FF00",
	"olive":                "#808000",
	"maroon":               "#800000",
	"silver":               "#C0C0C0",
	"gold":                 "#FFD700",
	"skyblue":              "#87CEEB",
	"lightblue":            "#ADD8E6",
	"lightgreen":           "#90EE90",
	"lightgrey":            "#D3D3D3",
	"lightgray":            "#D3D3D3",
	"darkblue":             "#00008B",
	"darkgreen":            "#006400",
	"darkred":              "#8B0000",
	"darkgray":             "#A9A9A9",
	"darkgrey":             "#A9A9A9",
	"cornflowerblue":       "#6495ED",
	"dodgerblue":           "#1E90FF",
	"deepskyblue":          "#00BFFF",
	"royalblue":            "#4169E1",
	"steelblue":            "#4682B4",
	"mediumslateblue":      "#7B68EE",
	"slateblue":            "#6A5ACD",
	"mediumseagreen":       "#3CB371",
	"seagreen":             "#2E8B57",
	"forestgreen":          "#228B22",
	"springgreen":          "#00FF7F",
	"tomato":               "#FF6347",
	"coral":                "#FF7F50",
	"salmon":               "#FA8072",
	"crimson":              "#DC143C",
	"indigo":               "#4B0082",
	"turquoise":            "#40E0D0",
	"khaki":                "#F0E68C",
}
