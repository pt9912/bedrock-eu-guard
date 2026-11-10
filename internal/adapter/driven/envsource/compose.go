package envsource

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

var yamlLineRE = regexp.MustCompile(`line (\d+)`)

// parseCompose liest die environment-Blöcke aller Services (Map und Liste
// "KEY=VAL"). env_file wird nicht gelesen. Anker/Aliase und Merge-Keys
// werden aufgelöst.
func parseCompose(path string, data []byte) (*model.EnvFile, *model.FileError) {
	f := &model.EnvFile{Path: path, Kind: model.EnvCompose}
	var root yaml.Node
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&root); err != nil {
		if errors.Is(err, io.EOF) {
			return f, nil // leere Datei: keine Variablen
		}
		return nil, &model.FileError{Path: path, Reason: model.ReasonFileInvalid, Line: yamlErrorLine(err)}
	}
	doc := resolve(&root)
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		doc = resolve(doc.Content[0])
	}
	if doc.Kind != yaml.MappingNode {
		return f, nil
	}
	services := mapGet(doc, "services")
	if services == nil || services.Kind != yaml.MappingNode {
		return f, nil
	}
	for i := 0; i+1 < len(services.Content); i += 2 {
		name := services.Content[i].Value
		svc := resolve(services.Content[i+1])
		if svc.Kind != yaml.MappingNode {
			continue
		}
		env := mapGet(svc, "environment")
		if env == nil {
			continue
		}
		f.Env = append(f.Env, composeEnv(resolve(env), name)...)
	}
	f.Env = dropInterpolatedRegions(f.Env)
	return f, nil
}

func composeEnv(n *yaml.Node, scope string) []model.EnvEntry {
	var out []model.EnvEntry
	add := func(key, val string, line int) {
		if key == "" {
			return
		}
		e := model.EnvEntry{Key: key, Line: line, Scope: scope}
		if secretkey.IsSecret(key) {
			e.Secret = true
		} else {
			e.Value = val
		}
		out = append(out, e)
	}
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i], resolve(n.Content[i+1])
			if k.Value == "<<" { // Merge-Key: Inhalt des Ankers übernehmen
				out = append(out, composeEnv(v, scope)...)
				continue
			}
			add(k.Value, scalar(v), k.Line)
		}
	case yaml.SequenceNode:
		for _, item := range n.Content {
			item = resolve(item)
			if item.Kind != yaml.ScalarNode {
				continue
			}
			key, val, _ := strings.Cut(item.Value, "=")
			add(key, val, item.Line)
		}
	}
	return out
}

// resolve folgt Aliasen (*anker).
func resolve(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

func mapGet(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// scalar liefert den Wert eines Skalars; Null und Nicht-Skalare ergeben "".
func scalar(n *yaml.Node) string {
	if n == nil || n.Kind != yaml.ScalarNode || n.Tag == "!!null" {
		return ""
	}
	return n.Value
}

// yamlErrorLine liest nur die Zeilennummer aus der yaml.v3-Meldung (die
// Meldung selbst kann Werte zitieren und wird nicht weitergegeben).
func yamlErrorLine(err error) int {
	if m := yamlLineRE.FindStringSubmatch(err.Error()); m != nil {
		if n, convErr := strconv.Atoi(m[1]); convErr == nil {
			return n
		}
	}
	return 0
}
