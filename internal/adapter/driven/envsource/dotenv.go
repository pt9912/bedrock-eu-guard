package envsource

import (
	"regexp"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// dotenvKind ist die Art von .env-Dateien. Sie wird durch model.EnvDotenv ersetzt,
// sobald der Kern-Vertrag von slice-022 auf main liegt.
const dotenvKind = model.EnvSourceKind("dotenv")

var dotenvKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.-]*$`)

// parseDotenv liest eine .env-Datei in der gängigen Syntax: Kommentarzeilen,
// optionales "export", KEY=VALUE mit optionalen Leerzeichen um "=", unquotierte,
// einfach und doppelt quotierte Werte (auch über mehrere Zeilen), Inline-Kommentare
// nach Leerzeichen und "#". Ein Schlüssel ohne "=" hat keinen Wert und erzeugt keinen
// Eintrag. Eine Zeile, die sich nicht lesen lässt, ergibt FILE_INVALID mit der Zeile,
// nie mit Dateiinhalt. Secret-Werte werden geschwärzt (secretkey).
func parseDotenv(path string, data []byte) (*model.EnvFile, *model.FileError) {
	f := &model.EnvFile{Path: path, Kind: dotenvKind}
	text := strings.TrimPrefix(string(data), "\ufeff")
	lines := strings.Split(text, "\n")
	invalid := func(line int) (*model.EnvFile, *model.FileError) {
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: line}
	}
	for i := 0; i < len(lines); i++ {
		trim := strings.TrimSpace(strings.TrimRight(lines[i], "\r"))
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		if rest, ok := strings.CutPrefix(trim, "export"); ok && rest != "" && (rest[0] == ' ' || rest[0] == '\t') {
			trim = strings.TrimSpace(rest)
		}
		eq := strings.IndexByte(trim, '=')
		if eq < 0 {
			if dotenvKey.MatchString(trim) {
				continue // Schlüssel ohne Wert
			}
			return invalid(i + 1)
		}
		key := strings.TrimSpace(trim[:eq])
		if !dotenvKey.MatchString(key) {
			return invalid(i + 1)
		}
		value, used, ok := dotenvValue(strings.TrimLeft(trim[eq+1:], " \t"), lines, i)
		if !ok {
			return invalid(i + 1)
		}
		e := model.EnvEntry{Key: key, Line: i + 1}
		if secretkey.IsSecret(key) {
			e.Secret = true
		} else {
			e.Value = value
		}
		f.Env = append(f.Env, e)
		i += used
	}
	f.Env = dropInterpolatedRegions(f.Env)
	return f, nil
}

// dotenvValue liest den Wert ab rest in Zeile i. used ist die Zahl zusätzlich
// verbrauchter Zeilen (mehrzeilige Werte in Anführungszeichen); ok=false bei
// nicht geschlossenem Anführungszeichen.
func dotenvValue(rest string, lines []string, i int) (value string, used int, ok bool) {
	if rest == "" {
		return "", 0, true
	}
	q := rest[0]
	if q != '"' && q != '\'' {
		// unquotiert: Inline-Kommentar beginnt mit Leerzeichen und "#"
		for j := 0; j < len(rest); j++ {
			if rest[j] == '#' && j > 0 && (rest[j-1] == ' ' || rest[j-1] == '\t') {
				rest = rest[:j]
				break
			}
		}
		return strings.TrimSpace(rest), 0, true
	}
	buf := rest[1:]
	for {
		if end := closingQuote(buf, q); end >= 0 {
			return unescape(buf[:end], q), used, true
		}
		used++
		if i+used >= len(lines) {
			return "", 0, false
		}
		buf += "\n" + strings.TrimRight(lines[i+used], "\r")
	}
}

// closingQuote liefert den Index des schließenden Anführungszeichens (bei
// doppelten Anführungszeichen unter Beachtung von Backslash) oder -1.
func closingQuote(s string, q byte) int {
	for j := 0; j < len(s); j++ {
		switch {
		case q == '"' && s[j] == '\\':
			j++
		case s[j] == q:
			return j
		}
	}
	return -1
}

func unescape(s string, q byte) string {
	if q == '\'' {
		return s
	}
	r := strings.NewReplacer(`\"`, `"`, `\\`, `\`, `\n`, "\n")
	return r.Replace(s)
}
