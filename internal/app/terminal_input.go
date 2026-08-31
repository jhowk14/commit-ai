package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"golang.org/x/term"
)

const commitMessagePrompt = "📝 Commit message: "

type keyKind uint8

const (
	keyRune keyKind = iota
	keyEnter
	keyCancel
	keyBackspace
	keyLeft
	keyRight
	keyHome
	keyEnd
	keyDelete
	keyClear
)

type terminalKey struct {
	kind keyKind
	rune rune
}

func terminalFiles(in io.Reader, out io.Writer) (*os.File, *os.File, bool) {
	input, inputOK := in.(*os.File)
	output, outputOK := out.(*os.File)
	if !inputOK || !outputOK || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) {
		return nil, nil, false
	}
	return input, output, true
}

// editPreFilledLine mirrors the old readline workflow: the generated message
// is already in the input line, so Enter accepts it and typing edits it.
func editPreFilledLine(input, output *os.File, suggested string) (string, error) {
	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(input.Fd()), state)

	value, cursor := []rune(suggested), len([]rune(suggested))
	reader := bufio.NewReader(input)
	fmt.Fprintln(output)
	for {
		renderEditableLine(output, value, cursor)
		key, err := readTerminalKey(reader)
		if err != nil {
			return "", err
		}
		switch key.kind {
		case keyEnter:
			fmt.Fprint(output, "\r\n")
			if strings.TrimSpace(string(value)) == "" {
				return suggested, nil
			}
			return string(value), nil
		case keyCancel:
			fmt.Fprint(output, "\r\n")
			return "", errors.New("commit cancelado")
		case keyLeft:
			if cursor > 0 {
				cursor--
			}
		case keyRight:
			if cursor < len(value) {
				cursor++
			}
		case keyHome:
			cursor = 0
		case keyEnd:
			cursor = len(value)
		case keyBackspace:
			if cursor > 0 {
				value = append(value[:cursor-1], value[cursor:]...)
				cursor--
			}
		case keyDelete:
			if cursor < len(value) {
				value = append(value[:cursor], value[cursor+1:]...)
			}
		case keyClear:
			value, cursor = nil, 0
		case keyRune:
			if !unicode.IsControl(key.rune) {
				value = append(value, 0)
				copy(value[cursor+1:], value[cursor:])
				value[cursor] = key.rune
				cursor++
			}
		}
	}
}

func readTerminalKey(reader *bufio.Reader) (terminalKey, error) {
	r, _, err := reader.ReadRune()
	if err != nil {
		return terminalKey{}, err
	}
	switch r {
	case '\r', '\n':
		return terminalKey{kind: keyEnter}, nil
	case 3:
		return terminalKey{kind: keyCancel}, nil
	case 8, 127:
		return terminalKey{kind: keyBackspace}, nil
	case 21:
		return terminalKey{kind: keyClear}, nil
	case 1:
		return terminalKey{kind: keyHome}, nil
	case 5:
		return terminalKey{kind: keyEnd}, nil
	case 27:
		return readEscapeSequence(reader)
	default:
		return terminalKey{kind: keyRune, rune: r}, nil
	}
}

func readEscapeSequence(reader *bufio.Reader) (terminalKey, error) {
	next, err := reader.ReadByte()
	if err != nil {
		return terminalKey{}, err
	}
	if next != '[' {
		return terminalKey{kind: keyRune}, nil
	}
	code, err := reader.ReadByte()
	if err != nil {
		return terminalKey{}, err
	}
	switch code {
	case 'D':
		return terminalKey{kind: keyLeft}, nil
	case 'C':
		return terminalKey{kind: keyRight}, nil
	case 'H':
		return terminalKey{kind: keyHome}, nil
	case 'F':
		return terminalKey{kind: keyEnd}, nil
	case '3':
		if _, err := reader.ReadByte(); err != nil {
			return terminalKey{}, err
		}
		return terminalKey{kind: keyDelete}, nil
	default:
		return terminalKey{kind: keyRune}, nil
	}
}

func renderEditableLine(output io.Writer, value []rune, cursor int) {
	fmt.Fprintf(output, "\r\033[2K%s%s", commitMessagePrompt, string(value))
	if remaining := displayWidth(value[cursor:]); remaining > 0 {
		fmt.Fprintf(output, "\033[%dD", remaining)
	}
}

func displayWidth(value []rune) int {
	width := 0
	for _, r := range value {
		switch {
		case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r), unicode.Is(unicode.Cf, r):
		case r >= 0x1100 && (r <= 0x115f || r >= 0x2e80):
			width += 2
		default:
			width++
		}
	}
	return width
}
