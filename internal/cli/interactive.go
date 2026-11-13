package cli

import (
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/iandesj/blerg/internal/rank"
	"golang.org/x/term"
)

type browseKey int

const (
	keyIgnore browseKey = iota
	keyNext
	keyPrev
	keyOpen
	keyQuit
)

func runInteractive(in io.Reader, out io.Writer, open func(string) error, matches []rank.MatchResult) error {
	if len(matches) == 0 {
		return fmt.Errorf("find: no matches to browse")
	}
	restore, raw, err := enterRaw(in)
	if err != nil {
		return err
	}
	defer restore()
	return browseLoop(bufio.NewReader(in), out, raw, open, matches)
}

func shuffleMatches(matches []rank.MatchResult, pick func(n int) int) {
	if pick == nil {
		pick = rand.IntN
	}
	for i := len(matches) - 1; i > 0; i-- {
		j := pick(i + 1)
		matches[i], matches[j] = matches[j], matches[i]
	}
}

func colorEnabled(raw bool) bool {
	return raw && os.Getenv("NO_COLOR") == ""
}

func browseLoop(in *bufio.Reader, out io.Writer, raw bool, open func(string) error, matches []rank.MatchResult) error {
	index := 0
	note := ""
	color := colorEnabled(raw)
	for {
		frame := renderFrame(matches[index], index, len(matches), note, color)
		if err := writeFrame(out, raw, frame); err != nil {
			return err
		}
		key, err := readBrowseKey(in)
		if err != nil {
			return ignoreEOF(err)
		}
		var quit bool
		index, note, quit = stepBrowse(index, matches, key, open)
		if quit {
			return nil
		}
	}
}

func stepBrowse(index int, matches []rank.MatchResult, key browseKey, open func(string) error) (int, string, bool) {
	switch key {
	case keyQuit:
		return index, "", true
	case keyNext:
		if index < len(matches)-1 {
			index++
		}
		return index, "", false
	case keyPrev:
		if index > 0 {
			index--
		}
		return index, "", false
	case keyOpen:
		if err := open(matches[index].URL); err != nil {
			return index, err.Error(), false
		}
		opened := time.Now().UTC()
		matches[index].OpenedAt = &opened
		return index, openNote(nil), false
	default:
		return index, "", false
	}
}

func openNote(err error) string {
	if err != nil {
		return err.Error()
	}
	return "opened in browser"
}

func ignoreEOF(err error) error {
	if err == io.EOF {
		return nil
	}
	return err
}

func renderFrame(match rank.MatchResult, index, total int, note string, color bool) string {
	var b strings.Builder
	pos := fmt.Sprintf("%d/%d", index+1, total)
	hint := "j/k next/prev   enter open   q quit"
	_, _ = fmt.Fprintf(&b, "%s   %s\n", paint(color, "\x1b[1;36m", pos), paint(color, "\x1b[2m", hint))
	writeField(&b, color, "Title", match.Title, "\x1b[1m")
	writeField(&b, color, "URL", match.URL, "\x1b[36m")
	writeField(&b, color, "Source", match.SourceName, "")
	published := "unknown"
	if match.PublishedAt != nil {
		published = match.PublishedAt.UTC().Format("2006-01-02")
	}
	writeField(&b, color, "Published", published, "")
	opened := "not yet"
	if match.OpenedAt != nil {
		opened = match.OpenedAt.UTC().Format("2006-01-02 15:04")
	}
	writeField(&b, color, "Opened", opened, "")
	if match.Explanation != "" {
		writeField(&b, color, "Match", match.Explanation, "")
	}
	if note != "" {
		code := "\x1b[32m"
		if note != "opened in browser" {
			code = "\x1b[31m"
		}
		_, _ = fmt.Fprintf(&b, "%s\n", paint(color, code, note))
	}
	return b.String()
}

func writeField(b *strings.Builder, color bool, label, value, valueCode string) {
	_, _ = fmt.Fprintf(b, "%s %s\n", paint(color, "\x1b[2m", label+":"), paint(color, valueCode, value))
}

func paint(enabled bool, code, text string) string {
	if !enabled || code == "" || text == "" {
		return text
	}
	return code + text + "\x1b[0m"
}

func writeFrame(out io.Writer, raw bool, frame string) error {
	if raw {
		frame = strings.ReplaceAll(frame, "\n", "\r\n")
		_, err := fmt.Fprintf(out, "\x1b[H\x1b[2J%s", frame)
		return err
	}
	_, err := io.WriteString(out, frame)
	return err
}

func readBrowseKey(r *bufio.Reader) (browseKey, error) {
	b, err := r.ReadByte()
	if err != nil {
		return keyIgnore, err
	}
	switch b {
	case 'j', 'J':
		return keyNext, nil
	case 'k', 'K':
		return keyPrev, nil
	case '\r', '\n':
		return keyOpen, nil
	case 'q', 'Q', 3:
		return keyQuit, nil
	case 0x1b:
		return readArrow(r), nil
	default:
		return keyIgnore, nil
	}
}

func readArrow(r *bufio.Reader) browseKey {
	if r.Buffered() < 2 {
		return keyQuit
	}
	bracket, err := r.ReadByte()
	if err != nil || bracket != '[' {
		return keyIgnore
	}
	dir, err := r.ReadByte()
	if err != nil {
		return keyIgnore
	}
	switch dir {
	case 'A':
		return keyPrev
	case 'B':
		return keyNext
	default:
		return keyIgnore
	}
}

func enterRaw(in io.Reader) (func(), bool, error) {
	f, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return func() {}, false, nil
	}
	state, err := term.MakeRaw(int(f.Fd()))
	if err != nil {
		return nil, false, err
	}
	return func() { _ = term.Restore(int(f.Fd()), state) }, true, nil
}
