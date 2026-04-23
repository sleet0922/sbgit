package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Color string

const (
	Reset           Color = "\x1b[0m"
	Bold            Color = "\x1b[1m"
	Dim             Color = "\x1b[2m"
	Italic          Color = "\x1b[3m"
	Underline       Color = "\x1b[4m"
	Blink           Color = "\x1b[5m"
	Reverse         Color = "\x1b[7m"
	Black           Color = "\x1b[30m"
	Red             Color = "\x1b[31m"
	Green           Color = "\x1b[32m"
	Yellow          Color = "\x1b[33m"
	Blue            Color = "\x1b[34m"
	Magenta         Color = "\x1b[35m"
	Cyan            Color = "\x1b[36m"
	White           Color = "\x1b[37m"
	BrightBlack     Color = "\x1b[90m"
	BrightRed       Color = "\x1b[91m"
	BrightGreen     Color = "\x1b[92m"
	BrightYellow    Color = "\x1b[93m"
	BrightBlue      Color = "\x1b[94m"
	BrightMagenta   Color = "\x1b[95m"
	BrightCyan      Color = "\x1b[96m"
	BrightWhite     Color = "\x1b[97m"
	BgBlack         Color = "\x1b[40m"
	BgRed           Color = "\x1b[41m"
	BgGreen         Color = "\x1b[42m"
	BgYellow        Color = "\x1b[43m"
	BgBlue          Color = "\x1b[44m"
	BgMagenta       Color = "\x1b[45m"
	BgCyan          Color = "\x1b[46m"
	BgWhite         Color = "\x1b[47m"
	BgBrightBlack   Color = "\x1b[100m"
	BgBrightRed     Color = "\x1b[101m"
	BgBrightGreen   Color = "\x1b[102m"
	BgBrightYellow  Color = "\x1b[103m"
	BgBrightBlue    Color = "\x1b[104m"
	BgBrightMagenta Color = "\x1b[105m"
	BgBrightCyan    Color = "\x1b[106m"
	BgBrightWhite   Color = "\x1b[107m"
)

func RGB(r, g, b int) Color {
	return Color(fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b))
}

func BgRGB(r, g, b int) Color {
	return Color(fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b))
}

type Theme struct {
	Border      Color
	BorderTitle Color
	Text        Color
	TextDim     Color
	Accent      Color
	AccentBg    Color
	Selected    Color
	SelectedBg  Color
	Success     Color
	Warning     Color
	Error       Color
	HeaderBg    Color
	HeaderFg    Color
	StatusBarBg Color
	StatusBarFg Color
	MenuItem    Color
	MenuHotkey  Color
	InputBorder Color
	InputActive Color
	Scrollbar   Color
	TabActive   Color
	TabInactive Color
	TabActiveBg Color
}

var DarkTheme = Theme{
	Border:      BrightBlack,
	BorderTitle: BrightCyan,
	Text:        White,
	TextDim:     BrightBlack,
	Accent:      BrightCyan,
	AccentBg:    BgBrightBlack,
	Selected:    BrightWhite,
	SelectedBg:  BgBlue,
	Success:     BrightGreen,
	Warning:     BrightYellow,
	Error:       BrightRed,
	HeaderBg:    BgBlue,
	HeaderFg:    BrightWhite,
	StatusBarBg: BgBrightBlack,
	StatusBarFg: BrightWhite,
	MenuItem:    White,
	MenuHotkey:  BrightYellow,
	InputBorder: BrightCyan,
	InputActive: BrightGreen,
	Scrollbar:   BrightBlack,
	TabActive:   BrightWhite,
	TabInactive: BrightBlack,
	TabActiveBg: BgBlue,
}

var CatppuccinTheme = Theme{
	Border:      RGB(88, 91, 112),
	BorderTitle: RGB(137, 180, 250),
	Text:        RGB(205, 214, 244),
	TextDim:     RGB(88, 91, 112),
	Accent:      RGB(137, 180, 250),
	AccentBg:    BgRGB(49, 50, 68),
	Selected:    RGB(205, 214, 244),
	SelectedBg:  BgRGB(49, 50, 68),
	Success:     RGB(166, 227, 161),
	Warning:     RGB(249, 226, 175),
	Error:       RGB(243, 139, 168),
	HeaderBg:    BgRGB(49, 50, 68),
	HeaderFg:    RGB(137, 180, 250),
	StatusBarBg: BgRGB(30, 30, 46),
	StatusBarFg: RGB(205, 214, 244),
	MenuItem:    RGB(205, 214, 244),
	MenuHotkey:  RGB(249, 226, 175),
	InputBorder: RGB(137, 180, 250),
	InputActive: RGB(166, 227, 161),
	Scrollbar:   RGB(88, 91, 112),
	TabActive:   RGB(205, 214, 244),
	TabInactive: RGB(88, 91, 112),
	TabActiveBg: BgRGB(49, 50, 68),
}

type Rect struct {
	X int
	Y int
	W int
	H int
}

type UI struct {
	term  *Terminal
	theme Theme
}

func NewUI(t *Terminal, theme Theme) *UI {
	return &UI{term: t, theme: theme}
}

func (u *UI) Clear() {
	u.term.Clear()
}

func (u *UI) SetPixel(row, col int, content string) {
	if row < 0 || col < 0 {
		return
	}
	u.term.MoveCursor(row, col)
	u.term.Write(content)
}

func strWidth(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe10 && r <= 0xfe19) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff01 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func padRight(s string, width int) string {
	sw := strWidth(s)
	if sw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}

func padLeft(s string, width int) string {
	sw := strWidth(s)
	if sw >= width {
		return s
	}
	return strings.Repeat(" ", width-sw) + s
}

func truncate(s string, width int) string {
	if strWidth(s) <= width {
		return s
	}
	for i := range s {
		if strWidth(s[:i+1]) > width-1 {
			return s[:i] + "…"
		}
	}
	return s
}

func stripANSI(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && ((s[j] >= '0' && s[j] <= '9') || s[j] == ';' || s[j] == '?') {
				j++
			}
			if j < len(s) && (s[j] >= 'A' && s[j] <= 'Z' || s[j] >= 'a' && s[j] <= 'z') {
				i = j + 1
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}
	return result.String()
}

func (u *UI) DrawBox(r Rect, title string) {
	tl := "╭"
	tr := "╮"
	bl := "╰"
	br := "╯"
	hz := "─"
	vt := "│"

	u.SetPixel(r.Y, r.X, string(tl))
	u.SetPixel(r.Y, r.X+r.W-1, string(tr))
	u.SetPixel(r.Y+r.H-1, r.X, string(bl))
	u.SetPixel(r.Y+r.H-1, r.X+r.W-1, string(br))

	topLine := strings.Repeat(hz, r.W-2)
	u.SetPixel(r.Y, r.X+1, string(u.theme.Border)+topLine+string(Reset))

	botLine := strings.Repeat(hz, r.W-2)
	u.SetPixel(r.Y+r.H-1, r.X+1, string(u.theme.Border)+botLine+string(Reset))

	for row := r.Y + 1; row < r.Y+r.H-1; row++ {
		u.SetPixel(row, r.X, string(u.theme.Border)+vt+string(Reset))
		u.SetPixel(row, r.X+r.W-1, string(u.theme.Border)+vt+string(Reset))
	}

	if title != "" {
		titleStr := " " + title + " "
		tw := strWidth(title)
		col := r.X + 2
		u.SetPixel(r.Y, col, string(u.theme.BorderTitle)+string(Bold)+titleStr+string(Reset)+string(u.theme.Border))
		_ = tw
	}
}

func (u *UI) DrawFilledBox(r Rect, title string, bg Color) {
	u.DrawBox(r, title)
	for row := r.Y + 1; row < r.Y+r.H-1; row++ {
		line := string(bg) + strings.Repeat(" ", r.W-2) + string(Reset)
		u.SetPixel(row, r.X+1, line)
	}
}

func (u *UI) DrawHeader(r Rect, text string) {
	u.SetPixel(r.Y, r.X, string(u.theme.HeaderBg)+string(u.theme.HeaderFg)+string(Bold)+padRight(text, r.W)+string(Reset))
}

func (u *UI) DrawStatusBar(r Rect, sections []string) {
	fullText := strings.Join(sections, " │ ")
	u.SetPixel(r.Y, r.X, string(u.theme.StatusBarBg)+string(u.theme.StatusBarFg)+padRight(fullText, r.W)+string(Reset))
}

type MenuItem struct {
	Label    string
	Hotkey   string
	Disabled bool
}

func (u *UI) DrawMenu(r Rect, items []MenuItem, selected int, scrollOffset int) {
	maxVisible := r.H - 2
	if maxVisible < 1 {
		maxVisible = 1
	}

	for i := 0; i < maxVisible; i++ {
		itemIdx := i + scrollOffset
		row := r.Y + 1 + i
		if row >= r.Y+r.H-1 {
			break
		}

		bgLine := string(u.theme.AccentBg) + strings.Repeat(" ", r.W-2) + string(Reset)
		u.SetPixel(row, r.X+1, bgLine)

		if itemIdx >= len(items) {
			continue
		}

		item := items[itemIdx]
		isSelected := itemIdx == selected

		var line string
		if isSelected {
			cursor := string(u.theme.Accent) + "❯" + string(Reset)
			numStr := fmt.Sprintf("%2d.", itemIdx+1)
			hotkeyStr := ""
			if item.Hotkey != "" {
				hotkeyStr = fmt.Sprintf(" [%s]", string(u.theme.MenuHotkey)+item.Hotkey+string(Reset)+string(u.theme.SelectedBg))
			}
			label := string(u.theme.SelectedBg) + string(u.theme.Selected) + string(Bold) + item.Label + string(Reset) + string(u.theme.SelectedBg)
			line = cursor + " " + string(u.theme.SelectedBg) + numStr + " " + label + hotkeyStr
			remaining := r.W - 2 - strWidth(stripANSI(cursor+" "))
			padW := remaining - strWidth(stripANSI(numStr+" "+item.Label+hotkeyStr))
			if padW > 0 {
				line += strings.Repeat(" ", padW)
			}
			line += string(Reset)
		} else {
			numStr := fmt.Sprintf("%2d.", itemIdx+1)
			hotkeyStr := ""
			if item.Hotkey != "" {
				hotkeyStr = fmt.Sprintf(" [%s]", string(u.theme.MenuHotkey)+item.Hotkey+string(Reset))
			}
			textColor := string(u.theme.MenuItem)
			if item.Disabled {
				textColor = string(u.theme.TextDim)
			}
			label := textColor + item.Label + string(Reset)
			line = "  " + numStr + " " + label + hotkeyStr
		}

		u.SetPixel(row, r.X+1, line)
	}

	if len(items) > maxVisible {
		scrollbarH := r.H - 2
		thumbSize := max(1, scrollbarH*maxVisible/len(items))
		thumbPos := scrollOffset * scrollbarH / len(items)
		for i := 0; i < scrollbarH; i++ {
			col := r.X + r.W - 2
			row := r.Y + 1 + i
			if i >= thumbPos && i < thumbPos+thumbSize {
				u.SetPixel(row, col, string(u.theme.Scrollbar)+"█"+string(Reset))
			} else {
				u.SetPixel(row, col, string(u.theme.Scrollbar)+"░"+string(Reset))
			}
		}
	}
}

type ListItem struct {
	Columns  []string
	Tag      string
	TagColor Color
}

func (u *UI) DrawList(r Rect, items []ListItem, selected int, scrollOffset int, colWidths []int) {
	maxVisible := r.H - 2
	if maxVisible < 1 {
		maxVisible = 1
	}

	for i := 0; i < maxVisible; i++ {
		itemIdx := i + scrollOffset
		row := r.Y + 1 + i
		if row >= r.Y+r.H-1 {
			break
		}

		if itemIdx >= len(items) {
			bgLine := strings.Repeat(" ", r.W-2)
			u.SetPixel(row, r.X+1, bgLine)
			continue
		}

		item := items[itemIdx]
		isSelected := itemIdx == selected

		var line string
		if isSelected {
			line = string(u.theme.SelectedBg) + string(u.theme.Selected) + string(Bold) + " ❯ " + string(Reset) + string(u.theme.SelectedBg)
		} else {
			line = "   "
		}

		for c, col := range item.Columns {
			w := 20
			if c < len(colWidths) {
				w = colWidths[c]
			}
			if isSelected {
				line += string(u.theme.SelectedBg) + string(u.theme.Selected) + padRight(truncate(col, w), w) + string(Reset) + string(u.theme.SelectedBg)
			} else {
				line += string(u.theme.Text) + padRight(truncate(col, w), w) + string(Reset)
			}
		}

		if item.Tag != "" {
			tagStr := " " + item.Tag + " "
			if isSelected {
				line += string(item.TagColor) + string(u.theme.SelectedBg) + tagStr + string(Reset) + string(u.theme.SelectedBg)
			} else {
				line += string(item.TagColor) + tagStr + string(Reset)
			}
		}

		plainLine := stripANSI(line)
		padW := r.W - 2 - strWidth(plainLine)
		if padW > 0 {
			if isSelected {
				line += strings.Repeat(" ", padW) + string(Reset)
			} else {
				line += strings.Repeat(" ", padW)
			}
		}
		if isSelected {
			line += string(Reset)
		}

		u.SetPixel(row, r.X+1, line)
	}
}

func (u *UI) DrawText(r Rect, text string) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		row := r.Y + 1 + i
		if row >= r.Y+r.H-1 {
			break
		}
		u.SetPixel(row, r.X+1, string(u.theme.Text)+truncate(line, r.W-2)+string(Reset))
	}
}

func (u *UI) DrawDialog(r Rect, title, message string, buttons []string, selectedBtn int) {
	u.DrawFilledBox(r, title, BgRGB(30, 30, 50))

	lines := strings.Split(message, "\n")
	for i, line := range lines {
		row := r.Y + 2 + i
		if row >= r.Y+r.H-3 {
			break
		}
		u.SetPixel(row, r.X+2, string(u.theme.Text)+truncate(line, r.W-4)+string(Reset))
	}

	btnRow := r.Y + r.H - 2
	totalBtnWidth := 0
	for _, b := range buttons {
		totalBtnWidth += utf8.RuneCountInString(b) + 4
	}
	startCol := r.X + (r.W-totalBtnWidth)/2

	for i, btn := range buttons {
		btnText := " " + btn + " "
		if i == selectedBtn {
			btnText = string(u.theme.SelectedBg) + string(u.theme.Selected) + string(Bold) + btnText + string(Reset)
		} else {
			btnText = string(u.theme.AccentBg) + string(u.theme.Accent) + btnText + string(Reset)
		}
		u.SetPixel(btnRow, startCol, btnText)
		startCol += utf8.RuneCountInString(buttons[i]) + 4
	}
}

type InputState struct {
	Value  string
	Cursor int
	Active bool
	Prompt string
}

func (u *UI) DrawInput(r Rect, state InputState) {
	borderColor := u.theme.InputBorder
	if state.Active {
		borderColor = u.theme.InputActive
	}

	u.DrawBox(r, state.Prompt)

	displayStart := 0
	displayWidth := r.W - 4
	valRunes := []rune(state.Value)
	if state.Cursor > displayWidth-1 {
		displayStart = state.Cursor - displayWidth + 5
		if displayStart < 0 {
			displayStart = 0
		}
	}
	visibleRunes := valRunes
	if displayStart < len(valRunes) {
		visibleRunes = valRunes[displayStart:]
	}
	if len(visibleRunes) > displayWidth {
		visibleRunes = visibleRunes[:displayWidth]
	}

	displayText := string(visibleRunes)
	row := r.Y + 1
	u.SetPixel(row, r.X+1, string(borderColor)+"> "+string(Reset)+string(u.theme.Text)+displayText+string(Reset))

	cursorCol := r.X + 3 + state.Cursor - displayStart
	if state.Active {
		u.SetPixel(row, cursorCol, string(Reverse)+string(u.theme.Text)+" "+string(Reset))
	}
}

func (u *UI) DrawTabs(r Rect, tabs []string, active int) {
	col := r.X
	for i, tab := range tabs {
		tabText := " " + tab + " "
		if i == active {
			tabText = string(u.theme.TabActiveBg) + string(u.theme.TabActive) + string(Bold) + tabText + string(Reset)
		} else {
			tabText = string(u.theme.TabInactive) + tabText + string(Reset)
		}
		u.SetPixel(r.Y, col, tabText)
		col += strWidth(stripANSI(tabText))
	}
	remaining := r.W - (col - r.X)
	if remaining > 0 {
		u.SetPixel(r.Y, col, strings.Repeat(" ", remaining))
	}
}

func (u *UI) DrawProgressBar(r Rect, progress float64, label string) {
	row := r.Y
	filled := int(float64(r.W) * progress)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", r.W-filled)
	u.SetPixel(row, r.X, string(u.theme.Accent)+bar+string(Reset))
	if label != "" {
		labelCol := r.X + (r.W-strWidth(label))/2
		if labelCol < r.X {
			labelCol = r.X
		}
		u.SetPixel(row, labelCol, string(Bold)+string(u.theme.HeaderFg)+label+string(Reset))
	}
}

func (u *UI) DrawDiff(r Rect, lines []string, scrollOffset int) {
	maxVisible := r.H - 2
	for i := 0; i < maxVisible; i++ {
		lineIdx := i + scrollOffset
		row := r.Y + 1 + i
		if row >= r.Y+r.H-1 {
			break
		}
		if lineIdx >= len(lines) {
			u.SetPixel(row, r.X+1, strings.Repeat(" ", r.W-2))
			continue
		}

		line := lines[lineIdx]
		var color Color
		var bg Color
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			color = BrightWhite
			bg = ""
		} else if strings.HasPrefix(line, "+") {
			color = BrightGreen
			bg = BgRGB(20, 40, 20)
		} else if strings.HasPrefix(line, "-") {
			color = BrightRed
			bg = BgRGB(40, 20, 20)
		} else if strings.HasPrefix(line, "@@") {
			color = BrightCyan
			bg = BgRGB(20, 30, 40)
		} else {
			color = u.theme.Text
			bg = ""
		}

		display := truncate(line, r.W-2)
		padded := padRight(display, r.W-2)
		if bg != "" {
			u.SetPixel(row, r.X+1, string(bg)+string(color)+padded+string(Reset))
		} else {
			u.SetPixel(row, r.X+1, string(color)+padded+string(Reset))
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
