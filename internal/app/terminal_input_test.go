package app

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadTerminalKeyRecognizesEditingKeys(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("a\x1b[D\x1b[C\x1b[3~\x7f\x15\r"))
	want := []keyKind{keyRune, keyLeft, keyRight, keyDelete, keyBackspace, keyClear, keyEnter}
	for _, expected := range want {
		key, err := readTerminalKey(reader)
		if err != nil || key.kind != expected {
			t.Fatalf("tecla=%#v err=%v; esperado=%v", key, err, expected)
		}
	}
}

func TestDisplayWidthHandlesTextAndEmoji(t *testing.T) {
	if got := displayWidth([]rune("fix: test")); got != 9 {
		t.Fatalf("largura ASCII: %d", got)
	}
	if got := displayWidth([]rune("🐛 fix")); got < 6 {
		t.Fatalf("largura com emoji: %d", got)
	}
}
