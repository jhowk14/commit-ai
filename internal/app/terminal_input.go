package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/jhowk14/commit-ai/v2/internal/i18n"
	"golang.org/x/term"
)

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
func editPreFilledLine(input, output *os.File, language i18n.Language, suggested string) (string, error) {
	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(input.Fd()), state)

	value, cursor := []rune(suggested), len([]rune(suggested))
	reader := bufio.NewReader(input)
	prompt := i18n.T(language, "editor_prompt")
	lastRender := renderedLine{}
	fmt.Fprintln(output)
	for {
		lastRender = renderEditableLine(output, value, cursor, prompt, lastRender)
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
			return "", errors.New(i18n.T(language, "commit_canceled"))
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

type renderedLine struct {
	rows      int
	cursorRow int
}

// renderEditableLine clears every visual row used by the previous render.
// Clearing only the current row leaves duplicated text when a long generated
// message wraps in the terminal, which made editing appear to append text.
func renderEditableLine(output *os.File, value []rune, cursor int, prompt string, previous renderedLine) renderedLine {
	clearRenderedLine(output, previous)
	columns, _, err := term.GetSize(int(output.Fd()))
	if err != nil || columns < 20 {
		columns = 80
	}
	promptWidth := displayWidth([]rune(prompt))
	valueWidth := displayWidth(value)
	cursorWidth := promptWidth + displayWidth(value[:cursor])
	rows := max(1, (promptWidth+valueWidth+columns-1)/columns)
	cursorRow := min(rows-1, cursorWidth/columns)
	cursorColumn := cursorWidth % columns

	fmt.Fprintf(output, "%s%s", prompt, string(value))
	endRow := rows - 1
	fmt.Fprint(output, "\r")
	if endRow > 0 {
		fmt.Fprintf(output, "\033[%dA", endRow)
	}
	if cursorRow > 0 {
		fmt.Fprintf(output, "\033[%dB", cursorRow)
	}
	if cursorColumn > 0 {
		fmt.Fprintf(output, "\033[%dC", cursorColumn)
	}
	return renderedLine{rows: rows, cursorRow: cursorRow}
}

func clearRenderedLine(output io.Writer, previous renderedLine) {
	if previous.rows == 0 {
		return
	}
	fmt.Fprint(output, "\r")
	if previous.cursorRow > 0 {
		fmt.Fprintf(output, "\033[%dA", previous.cursorRow)
	}
	for row := 0; row < previous.rows; row++ {
		fmt.Fprint(output, "\033[2K")
		if row+1 < previous.rows {
			fmt.Fprint(output, "\033[1B\r")
		}
	}
	if previous.rows > 1 {
		fmt.Fprintf(output, "\033[%dA\r", previous.rows-1)
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
