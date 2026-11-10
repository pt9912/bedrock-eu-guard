package envsource

import (
	"strconv"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// instruction ist eine logische Dockerfile-Anweisung (Fortsetzungszeilen
// zusammengefasst) mit der Zeile ihres Beginns.
type instruction struct {
	line int
	word string // Großbuchstaben, z. B. "ENV"
	rest string
}

// parseDockerfile liest die ENV-Anweisungen (alte und neue Form, mehrere
// Paare je Zeile) samt Stage. ARG und andere Anweisungen werden ignoriert.
func parseDockerfile(path string, data []byte) (*model.EnvFile, *model.FileError) {
	f := &model.EnvFile{Path: path, Kind: model.EnvDockerfile}
	stage, stageNo := "", 0
	for _, in := range splitInstructions(string(data)) {
		switch in.word {
		case "FROM":
			stageNo++
			stage = strconv.Itoa(stageNo - 1)
			fields := strings.Fields(in.rest)
			for i, w := range fields {
				if strings.EqualFold(w, "AS") && i+1 < len(fields) {
					stage = fields[i+1]
				}
			}
		case "ENV":
			for _, kv := range parseEnvPairs(in.rest) {
				e := model.EnvEntry{Key: kv[0], Line: in.line, Scope: stage}
				if secretkey.IsSecret(kv[0]) {
					e.Secret = true
				} else {
					e.Value = kv[1]
				}
				f.Env = append(f.Env, e)
			}
		}
	}
	f.Env = dropInterpolatedRegions(f.Env)
	return f, nil
}

// splitInstructions fasst Zeilenfortsetzungen (Backslash am Zeilenende)
// zusammen; Kommentar- und Leerzeilen entfallen auch innerhalb einer
// Fortsetzung.
func splitInstructions(src string) []instruction {
	var out []instruction
	var buf strings.Builder
	start := 0
	flush := func() {
		text := strings.TrimSpace(buf.String())
		buf.Reset()
		if text == "" {
			return
		}
		word, rest := text, ""
		if i := strings.IndexAny(text, " \t"); i >= 0 {
			word, rest = text[:i], strings.TrimSpace(text[i:])
		}
		out = append(out, instruction{line: start, word: strings.ToUpper(word), rest: rest})
	}
	continued := false
	for i, raw := range strings.Split(src, "\n") {
		ln := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "#") || (trimmed == "" && continued) {
			continue
		}
		if !continued {
			if trimmed == "" {
				continue
			}
			start = i + 1
		}
		if strings.HasSuffix(strings.TrimRight(ln, " \t"), "\\") {
			buf.WriteString(strings.TrimSuffix(strings.TrimRight(ln, " \t"), "\\"))
			buf.WriteString(" ")
			continued = true
			continue
		}
		buf.WriteString(ln)
		continued = false
		flush()
	}
	if continued {
		flush()
	}
	return out
}

// parseEnvPairs zerlegt den Rest einer ENV-Anweisung. Neue Form:
// KEY=VALUE [KEY2="VAL 2" ...]; alte Form: KEY VALUE mit Rest als Wert.
func parseEnvPairs(rest string) [][2]string {
	first := rest
	if i := strings.IndexAny(rest, " \t"); i >= 0 {
		first = rest[:i]
	}
	if !strings.Contains(first, "=") {
		key, val, _ := strings.Cut(rest, " ")
		if key == "" {
			return nil
		}
		return [][2]string{{key, unquote(strings.TrimSpace(val))}}
	}
	var pairs [][2]string
	for _, tok := range splitQuoted(rest) {
		k, v, ok := strings.Cut(tok, "=")
		if !ok || k == "" {
			continue
		}
		pairs = append(pairs, [2]string{k, unquote(v)})
	}
	return pairs
}

// splitQuoted trennt an Whitespace außerhalb von Anführungszeichen; die
// Anführungszeichen bleiben im Token (unquote entfernt sie später).
func splitQuoted(s string) []string {
	var toks []string
	var cur strings.Builder
	var quote byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			cur.WriteByte(c)
			if c == '\\' && quote == '"' && i+1 < len(s) {
				i++
				cur.WriteByte(s[i])
			} else if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
			cur.WriteByte(c)
		case c == ' ' || c == '\t':
			if cur.Len() > 0 {
				toks = append(toks, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		toks = append(toks, cur.String())
	}
	return toks
}

// unquote entfernt umschließende Anführungszeichen (auch mitten im Token,
// wie KEY="a b").
func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}
