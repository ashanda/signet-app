package migrate

import (
	"fmt"
	"strings"
)

// SplitStatements splits a migration file into individual statements on
// top-level semicolons. It understands '...', "..." and `...` quoting
// (including backslash and doubled-quote escapes), and strips `-- `, `#`
// and /* */ comments. MySQL executable comments (/*! ... */) are code and
// are kept. DELIMITER is a mysql-client command, not SQL, so stored
// routines/triggers that need it are rejected rather than mis-split.
func SplitStatements(src string) ([]string, error) {
	var (
		stmts   []string
		b       strings.Builder
		hasCode bool
	)
	flush := func() {
		if hasCode {
			stmts = append(stmts, strings.TrimSpace(b.String()))
		}
		b.Reset()
		hasCode = false
	}

	n := len(src)
	for i := 0; i < n; {
		c := src[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			j := i + 1
			for ; j < n; j++ {
				if src[j] == '\\' && c != '`' {
					j++
					continue
				}
				if src[j] == c {
					if j+1 < n && src[j+1] == c {
						j++
						continue
					}
					break
				}
			}
			if j >= n {
				return nil, fmt.Errorf("unterminated %c quote starting at byte %d", c, i)
			}
			b.WriteString(src[i : j+1])
			hasCode = true
			i = j + 1

		case c == '-' && i+1 < n && src[i+1] == '-' && (i+2 >= n || isSpace(src[i+2])),
			c == '#':
			for i < n && src[i] != '\n' {
				i++
			}

		case c == '/' && i+1 < n && src[i+1] == '*':
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				return nil, fmt.Errorf("unterminated /* comment starting at byte %d", i)
			}
			stop := i + 2 + end + 2
			if i+2 < n && src[i+2] == '!' {
				b.WriteString(src[i:stop])
				hasCode = true
			} else {
				b.WriteByte(' ')
			}
			i = stop

		case c == ';':
			flush()
			i++

		default:
			b.WriteByte(c)
			if !isSpace(c) {
				hasCode = true
			}
			i++
		}
	}
	flush()

	for _, s := range stmts {
		if f := strings.Fields(s); len(f) > 0 && strings.EqualFold(f[0], "DELIMITER") {
			return nil, fmt.Errorf("DELIMITER is not supported (it is a mysql-client command, not SQL)")
		}
	}
	return stmts, nil
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
