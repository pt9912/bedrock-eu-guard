package envsource

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// stripJSONC ersetzt Kommentare und Trailing Commas durch Leerzeichen. Die
// Länge und alle Zeilenumbrüche bleiben erhalten, damit Offsets und Zeilen
// des Originals gültig bleiben. Zeichenketten bleiben unverändert.
func stripJSONC(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	blank := func(i int) {
		if out[i] != '\n' && out[i] != '\r' {
			out[i] = ' '
		}
	}
	for i := 0; i < len(in); i++ {
		switch c := in[i]; {
		case c == '"':
			for i++; i < len(in) && in[i] != '"'; i++ {
				if in[i] == '\\' {
					i++
				}
			}
		case c == '/' && i+1 < len(in) && in[i+1] == '/':
			for ; i < len(in) && in[i] != '\n'; i++ {
				blank(i)
			}
			i--
		case c == '/' && i+1 < len(in) && in[i+1] == '*':
			blank(i)
			blank(i + 1)
			for i += 2; i < len(in); i++ {
				if in[i] == '*' && i+1 < len(in) && in[i+1] == '/' {
					blank(i)
					blank(i + 1)
					i++
					break
				}
				blank(i)
			}
		}
	}
	// Trailing Commas: ',' gefolgt (nur Whitespace dazwischen) von '}' oder ']'.
	inString := false
	for i := 0; i < len(out); i++ {
		switch c := out[i]; {
		case c == '"':
			inString = !inString
		case inString && c == '\\':
			i++
		case !inString && c == ',':
			j := i + 1
			for j < len(out) && (out[j] == ' ' || out[j] == '\t' || out[j] == '\n' || out[j] == '\r') {
				j++
			}
			if j < len(out) && (out[j] == '}' || out[j] == ']') {
				out[i] = ' '
			}
		}
	}
	return out
}

// parseDevcontainer liest containerEnv und remoteEnv einer devcontainer.json.
func parseDevcontainer(path string, data []byte) (*model.EnvFile, *model.FileError) {
	clean := stripJSONC(data)
	dec := json.NewDecoder(bytes.NewReader(clean))
	dec.UseNumber()
	f := &model.EnvFile{Path: path, Kind: model.EnvDevcontainer}
	invalid := func(err error) *model.FileError {
		off := dec.InputOffset()
		var se *json.SyntaxError
		if errors.As(err, &se) && se.Offset > off {
			off = se.Offset
		}
		return &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: lineAt(data, off)}
	}

	if d, err := dec.Token(); err != nil || d != json.Delim('{') {
		return nil, invalid(err)
	}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return nil, invalid(err)
		}
		key, _ := kt.(string)
		if key == "containerEnv" || key == "remoteEnv" {
			if err := parseEnvObject(dec, data, f); err != nil {
				return nil, invalid(err)
			}
			continue
		}
		var skip json.RawMessage
		if err := dec.Decode(&skip); err != nil {
			return nil, invalid(err)
		}
	}
	if _, err := dec.Token(); err != nil {
		return nil, invalid(err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: lineAt(data, dec.InputOffset())}
	}
	f.Env = dropInterpolatedRegions(f.Env)
	return f, nil
}

func parseEnvObject(dec *json.Decoder, data []byte, f *model.EnvFile) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	d, ok := tok.(json.Delim)
	if !ok || (d != '{' && d != '[') {
		return nil // null o. ä.: kein Env-Block
	}
	if d == '[' {
		for dec.More() {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return err
			}
		}
		_, err := dec.Token()
		return err
	}
	for dec.More() {
		kt, err := dec.Token()
		if err != nil {
			return err
		}
		key, _ := kt.(string)
		line := lineAt(data, dec.InputOffset())
		var v any
		if err := dec.Decode(&v); err != nil {
			return err
		}
		e := model.EnvEntry{Key: key, Line: line}
		if secretkey.IsSecret(key) {
			e.Secret = true
		} else {
			e.Value = stringify(v)
		}
		f.Env = append(f.Env, e)
	}
	_, err = dec.Token()
	return err
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "true"
		}
		return "false"
	}
	return ""
}

func lineAt(data []byte, off int64) int {
	if off > int64(len(data)) {
		off = int64(len(data))
	}
	if off < 0 {
		off = 0
	}
	return 1 + bytes.Count(data[:off], []byte("\n"))
}
