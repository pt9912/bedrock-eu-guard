package terraform

import (
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/iampolicy"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

// policyResources sind Ressourcentypen mit einem "policy"-Attribut.
var policyResources = map[string]bool{
	"aws_iam_policy":       true,
	"aws_iam_role_policy":  true,
	"aws_iam_user_policy":  true,
	"aws_iam_group_policy": true,
}

// providerFrom liest region und alias eines provider "aws"-Blocks.
func providerFrom(b *hclsyntax.Block) model.AWSProvider {
	p := model.AWSProvider{Line: b.DefRange().Start.Line}
	if attr, ok := b.Body.Attributes["alias"]; ok {
		p.Alias, _ = evalString(attr.Expr)
	}
	if attr, ok := b.Body.Attributes["region"]; ok {
		p.Line = attr.SrcRange.Start.Line
		if s, ok := evalString(attr.Expr); ok {
			p.Region = s
		} else {
			p.RegionDynamic = true
		}
	}
	return p
}

// documentStatements liest die statement-Blöcke eines
// data "aws_iam_policy_document".
func documentStatements(b *hclsyntax.Block) []model.IAMStatement {
	var out []model.IAMStatement
	for _, st := range b.Body.Blocks {
		if st.Type != "statement" {
			continue
		}
		s := model.IAMStatement{Effect: "Allow", Line: st.DefRange().Start.Line}
		if attr, ok := st.Body.Attributes["effect"]; ok {
			if e, ok := evalString(attr.Expr); ok {
				s.Effect = e
			} else {
				s.Effect, s.Dynamic = "", true
			}
		}
		if attr, ok := st.Body.Attributes["actions"]; ok {
			var dyn bool
			s.Actions, dyn = evalStringList(attr.Expr)
			s.Dynamic = s.Dynamic || dyn
		}
		if attr, ok := st.Body.Attributes["resources"]; ok {
			var dyn bool
			s.Resources, dyn = evalStringList(attr.Expr)
			s.Dynamic = s.Dynamic || dyn
		}
		for _, c := range st.Body.Blocks {
			if c.Type != "condition" {
				continue
			}
			if attr, ok := c.Body.Attributes["variable"]; ok {
				if v, ok := evalString(attr.Expr); ok && strings.EqualFold(v, "aws:RequestedRegion") {
					s.HasRegionCondition = true
					if vals, ok := c.Body.Attributes["values"]; ok {
						regions, _ := evalStringList(vals.Expr) // nicht auflösbar: Regionen bleiben unbekannt
						s.ConditionRegions = append(s.ConditionRegions, regions...)
					}
				}
			}
		}
		out = append(out, s)
	}
	return out
}

// resourcePolicy liest das policy-Attribut (jsonencode(...) oder JSON-String/
// Heredoc). Nicht auflösbare Ausdrücke ergeben ein Statement mit Dynamic=true.
func resourcePolicy(b *hclsyntax.Block) []model.IAMStatement {
	attr, ok := b.Body.Attributes["policy"]
	if !ok {
		return nil
	}
	line := attr.SrcRange.Start.Line
	text, ok := policyJSON(attr.Expr)
	if !ok {
		return []model.IAMStatement{{Dynamic: true, Line: line}}
	}
	stmts, ok := iampolicy.Statements([]byte(text), line)
	if !ok {
		return []model.IAMStatement{{Dynamic: true, Line: line}}
	}
	return stmts
}

// policyJSON liefert den Policy-Text aus jsonencode(...) oder einem
// auflösbaren String/Heredoc.
func policyJSON(e hclsyntax.Expression) (string, bool) {
	if fc, ok := e.(*hclsyntax.FunctionCallExpr); ok {
		if fc.Name != "jsonencode" || len(fc.Args) != 1 {
			return "", false
		}
		v, diags := fc.Args[0].Value(nil)
		if diags.HasErrors() || !v.IsWhollyKnown() {
			return "", false
		}
		b, err := ctyjson.Marshal(v, v.Type())
		if err != nil {
			return "", false
		}
		return string(b), true
	}
	return evalString(e)
}

// evalString wertet einen Ausdruck ohne Kontext aus; nur reine String-Literale
// (auch Heredocs ohne Interpolation) gelten als aufgelöst.
func evalString(e hclsyntax.Expression) (string, bool) {
	v, diags := e.Value(nil)
	if diags.HasErrors() || !v.IsKnown() || v.IsNull() || v.Type() != cty.String {
		return "", false
	}
	return v.AsString(), true
}

// evalStringList liefert die auflösbaren String-Literale einer Liste; dynamic
// ist wahr, wenn mindestens ein Eintrag (oder die ganze Liste) nicht
// auflösbar war.
func evalStringList(e hclsyntax.Expression) (vals []string, dynamic bool) {
	if tc, ok := e.(*hclsyntax.TupleConsExpr); ok {
		for _, el := range tc.Exprs {
			if s, ok := evalString(el); ok {
				vals = append(vals, s)
			} else {
				dynamic = true
			}
		}
		return vals, dynamic
	}
	if s, ok := evalString(e); ok {
		return []string{s}, false
	}
	return nil, true
}

var (
	// arnRE: Bedrock-ARN; idRE: Modell-/Profil-ID mit Bindestrich im Modellteil,
	// damit Prosa wie "Amazon.com" nicht trifft.
	arnRE = regexp.MustCompile(`arn:aws[a-z-]*:bedrock:[a-z0-9-]*:[0-9]*:[A-Za-z0-9/._:*-]+`)
	idRE  = regexp.MustCompile(`\b(?:(?:us|eu|apac|global|us-gov)\.)?(?:anthropic|amazon|meta|mistral|cohere|ai21|deepseek|writer|stability|openai|qwen|twelvelabs|luma)\.[a-z0-9][a-z0-9._]*-[a-z0-9._:-]*`)
)

// modelRefs sammelt Modell-IDs und Bedrock-ARNs aus allen String-Literalen.
// Zurückgegeben wird nur das gefundene Token, nie das ganze Literal.
func modelRefs(body *hclsyntax.Body) []model.ModelRef {
	var out []model.ModelRef
	seen := map[model.ModelRef]bool{}
	add := func(value string, line int) {
		value = strings.TrimRight(value, ".:-")
		r := model.ModelRef{Value: value, Line: line}
		if value != "" && !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	_ = hclsyntax.VisitAll(body, func(n hclsyntax.Node) hcl.Diagnostics {
		lit, ok := n.(*hclsyntax.LiteralValueExpr)
		if !ok || lit.Val.IsNull() || !lit.Val.IsKnown() || lit.Val.Type() != cty.String {
			return nil
		}
		s := lit.Val.AsString()
		start := lit.SrcRange.Start.Line
		lineOf := func(idx int) int { return start + strings.Count(s[:idx], "\n") }

		arns := arnRE.FindAllStringIndex(s, -1)
		for _, m := range arns {
			add(s[m[0]:m[1]], lineOf(m[0]))
		}
		for _, m := range idRE.FindAllStringIndex(s, -1) {
			inside := false
			for _, a := range arns {
				if m[0] >= a[0] && m[1] <= a[1] {
					inside = true
				}
			}
			if !inside {
				add(s[m[0]:m[1]], lineOf(m[0]))
			}
		}
		return nil
	})
	return out
}
