// Package iampolicy flacht IAM-Policy-Dokumente (JSON) zu model.IAMStatement
// ab. Es wird von den Adaptern awscli und terraform gemeinsam genutzt, damit
// beide dieselbe Auslegung liefern. Der Kern sieht nur diese Fakten.
package iampolicy

import (
	"encoding/json"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// RegionConditionKey ist der Condition-Schlüssel für die Regionseinschränkung.
const RegionConditionKey = "aws:requestedregion"

// Statements liest ein Policy-Dokument. Dokumente ohne lesbare Statements
// liefern ok=false. NotAction und NotResource sowie nicht lesbare Statements
// ergeben Dynamic=true (nicht auswertbar). line wird jedem Statement
// mitgegeben (0 = unbekannt).
func Statements(doc []byte, line int) (out []model.IAMStatement, ok bool) {
	var d struct {
		Statement json.RawMessage `json:"Statement"`
	}
	if err := json.Unmarshal(doc, &d); err != nil || len(d.Statement) == 0 {
		return nil, false
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(d.Statement, &raws); err != nil {
		raws = []json.RawMessage{d.Statement} // einzelnes Statement als Objekt
	}
	for _, raw := range raws {
		var st struct {
			Effect      string                    `json:"Effect"`
			Action      json.RawMessage           `json:"Action"`
			NotAction   json.RawMessage           `json:"NotAction"`
			Resource    json.RawMessage           `json:"Resource"`
			NotResource json.RawMessage           `json:"NotResource"`
			Condition   map[string]map[string]any `json:"Condition"`
		}
		if err := json.Unmarshal(raw, &st); err != nil {
			out = append(out, model.IAMStatement{Dynamic: true, Line: line})
			continue
		}
		s := model.IAMStatement{
			Effect:    st.Effect,
			Actions:   strs(st.Action),
			Resources: strs(st.Resource),
			Dynamic:   len(st.NotAction) > 0 || len(st.NotResource) > 0,
			Line:      line,
		}
		for _, keys := range st.Condition {
			for k, v := range keys {
				if strings.EqualFold(k, RegionConditionKey) {
					s.HasRegionCondition = true
					s.ConditionRegions = append(s.ConditionRegions, anyStrings(v)...)
				}
			}
		}
		out = append(out, s)
	}
	return out, true
}

// strs liest einen String oder eine Liste von Strings.
func strs(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var one string
	if err := json.Unmarshal(raw, &one); err == nil {
		return []string{one}
	}
	var many []string
	if err := json.Unmarshal(raw, &many); err == nil {
		return many
	}
	return nil
}

func anyStrings(v any) []string {
	switch x := v.(type) {
	case string:
		return []string{x}
	case []any:
		var out []string
		for _, e := range x {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
