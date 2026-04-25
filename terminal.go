package main

import (
	"os"

	"golang.org/x/term"
)

type Terminal struct {
	oldState  *term.State
	width     int
	height    int
	out       *os.File
	in        *os.File
	altScreen bool
}

func NewTerminal() *Terminal {
	return &Terminal{
		out: os.Stdout,
		in:  os.Stdin,
	}
}

func (t *Terminal) Init() error {
	var err error
	t.oldState, err = term.MakeRaw(int(t.in.Fd()))
	if err != nil {
		return err
	}
	t.altScreen = false
	t.EnterAltScreen()
	t.HideCursor()
	t.GetSize()
	return nil
}

func (t *Terminal) Restore() {
	t.ShowCursor()
	t.LeaveAltScreen()
	if t.oldState != nil {
		term.Restore(int(t.in.Fd()), t.oldState)
	}
}

func (t *Terminal) EnterAltScreen() {
	if !t.altScreen {
		t.Write("\x1b[?1049h")
		t.altScreen = true
	}
}

func (t *Terminal) LeaveAltScreen() {
	if t.altScreen {
		t.Write("\x1b[?1049l")
		t.altScreen = false
	}
}

func (t *Terminal) GetSize() {
	w, h, err := term.GetSize(int(t.out.Fd()))
	if err != nil {
		t.width = 80
		t.height = 24
		return
	}
	t.width = w
	t.height = h
}

func (t *Terminal) Width() int  { return t.width }
func (t *Terminal) Height() int { return t.height }

func (t *Terminal) Write(s string) {
	t.out.WriteString(s)
}

func (t *Terminal) Clear() {
	t.Write("\x1b[2J\x1b[H")
}

func (t *Terminal) MoveCursor(row, col int) {
	t.Write("\x1b[" + itoa(row+1) + ";" + itoa(col+1) + "H")
}

func (t *Terminal) HideCursor() {
	t.Write("\x1b[?25l")
}

func (t *Terminal) ShowCursor() {
	t.Write("\x1b[?25h")
}

func (t *Terminal) ReadKey() Key {
	buf := make([]byte, 16)
	n, err := t.in.Read(buf)
	if err != nil || n == 0 {
		return Key{Type: KeyError}
	}
	return parseKey(buf[:n])
}

type KeyType int

const (
	KeyError KeyType = iota
	KeyRune
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyDelete
	KeyTab
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyCtrlC
	KeyCtrlD
	KeyCtrlA
	KeyCtrlE
	KeyCtrlU
	KeyCtrlK
	KeyCtrlW
	KeyCtrlL
	KeyCtrlR
	KeyCtrlS
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

type Key struct {
	Type KeyType
	Rune rune
}

func parseKey(buf []byte) Key {
	if len(buf) == 0 {
		return Key{Type: KeyError}
	}

	if buf[0] == 0x03 {
		return Key{Type: KeyCtrlC}
	}
	if buf[0] == 0x04 {
		return Key{Type: KeyCtrlD}
	}
	if buf[0] == 0x01 {
		return Key{Type: KeyCtrlA}
	}
	if buf[0] == 0x05 {
		return Key{Type: KeyCtrlE}
	}
	if buf[0] == 0x15 {
		return Key{Type: KeyCtrlU}
	}
	if buf[0] == 0x0b {
		return Key{Type: KeyCtrlK}
	}
	if buf[0] == 0x17 {
		return Key{Type: KeyCtrlW}
	}
	if buf[0] == 0x0c {
		return Key{Type: KeyCtrlL}
	}
	if buf[0] == 0x12 {
		return Key{Type: KeyCtrlR}
	}
	if buf[0] == 0x13 {
		return Key{Type: KeyCtrlS}
	}

	if buf[0] == 0x1b {
		if len(buf) == 1 {
			return Key{Type: KeyEscape}
		}
		if buf[1] == '[' {
			if len(buf) == 2 {
				return Key{Type: KeyEscape}
			}
			switch buf[2] {
			case 'A':
				return Key{Type: KeyUp}
			case 'B':
				return Key{Type: KeyDown}
			case 'C':
				return Key{Type: KeyRight}
			case 'D':
				return Key{Type: KeyLeft}
			case 'H':
				return Key{Type: KeyHome}
			case 'F':
				return Key{Type: KeyEnd}
			case '1':
				if len(buf) > 3 && buf[3] == '~' {
					return Key{Type: KeyHome}
				}
				if len(buf) > 3 && buf[3] == ';' && buf[4] == '5' {
					return Key{Type: KeyCtrlA}
				}
			case '3':
				if len(buf) > 3 && buf[3] == '~' {
					return Key{Type: KeyDelete}
				}
			case '4':
				if len(buf) > 3 && buf[3] == '~' {
					return Key{Type: KeyEnd}
				}
			case '5':
				if len(buf) > 3 && buf[3] == '~' {
					return Key{Type: KeyPageUp}
				}
			case '6':
				if len(buf) > 3 && buf[3] == '~' {
					return Key{Type: KeyPageDown}
				}
			case 'O':
				if len(buf) > 3 {
					switch buf[3] {
					case 'P':
						return Key{Type: KeyF1}
					case 'Q':
						return Key{Type: KeyF2}
					case 'R':
						return Key{Type: KeyF3}
					case 'S':
						return Key{Type: KeyF4}
					}
				}
			}
			if len(buf) > 3 && buf[2] == '1' && buf[3] >= '5' && buf[3] <= '9' && buf[4] == '~' {
				switch buf[3] {
				case '5':
					return Key{Type: KeyF5}
				case '7':
					return Key{Type: KeyF6}
				case '8':
					return Key{Type: KeyF7}
				case '9':
					return Key{Type: KeyF8}
				}
			}
			if len(buf) > 4 && buf[2] == '1' && buf[3] == '9' && buf[4] == '~' {
				return Key{Type: KeyF9}
			}
			if len(buf) > 4 && buf[2] == '2' && buf[3] == '0' && buf[4] == '~' {
				return Key{Type: KeyF10}
			}
			if len(buf) > 4 && buf[2] == '2' && buf[3] == '1' && buf[4] == '~' {
				return Key{Type: KeyF11}
			}
			if len(buf) > 4 && buf[2] == '2' && buf[3] == '3' && buf[4] == '~' {
				return Key{Type: KeyF12}
			}
		}
		return Key{Type: KeyEscape}
	}

	if buf[0] == 0x0d {
		return Key{Type: KeyEnter}
	}
	if buf[0] == 0x7f || buf[0] == 0x08 {
		return Key{Type: KeyBackspace}
	}
	if buf[0] == 0x09 {
		return Key{Type: KeyTab}
	}

	r := rune(buf[0])
	if r >= 0x20 && r <= 0x7e {
		return Key{Type: KeyRune, Rune: r}
	}
	if r >= 0xc0 && len(buf) > 1 {
		var str []byte
		str = append(str, buf[0])
		i := 1
		for i < len(buf) && i < 4 && buf[i]&0xc0 == 0x80 {
			str = append(str, buf[i])
			i++
		}
		runes := []rune(string(str))
		if len(runes) > 0 {
			return Key{Type: KeyRune, Rune: runes[0]}
		}
	}

	return Key{Type: KeyError}
}

func EnableANSI() {
	// Linux/macOS 默认支持 ANSI，无需额外设置
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func (t *Terminal) Flush() {
	t.out.Sync()
}

func (t *Terminal) FillLine(row, col, width int, ch string) {
	t.MoveCursor(row, col)
	s := ""
	for i := 0; i < width; i++ {
		s += ch
	}
	t.Write(s)
}
