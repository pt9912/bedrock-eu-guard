package awscli

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/iampolicy"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

const (
	opSimulate = "iam:SimulatePrincipalPolicy"
	opPolicies = "iam:GetPrincipalPolicies"
	opGetRole  = "iam:GetRole"
)

// principal ist ein auflösbarer IAM-Principal (Benutzer oder Rolle).
type principal struct {
	kind string // "user" | "role"
	name string // Benutzer- bzw. Rollenname (ohne Pfad)
	arn  string // IAM-ARN (bei Rollen inklusive Pfad)
}

// resolve wandelt den von STS gemeldeten ARN in einen IAM-Principal. Für
// `assumed-role` wird der IAM-Rollen-ARN (inklusive Pfad) über `iam get-role`
// ermittelt. Nicht auflösbare Principals (Root, Federated User) ergeben einen
// AWS_ACCESS_DENIED-Fehler: Die Rechte sind dann nicht ermittelbar.
func (r Reader) resolve(op, arn string) (principal, error) {
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) != 6 || !strings.HasPrefix(parts[0], "arn") {
		return principal{}, &model.AWSError{Op: op, Reason: model.AWSAccessDenied}
	}
	service, resource := parts[2], parts[5]
	segs := strings.Split(resource, "/")
	last := segs[len(segs)-1]
	switch {
	case service == "iam" && strings.HasPrefix(resource, "user/") && last != "":
		return principal{kind: "user", name: last, arn: arn}, nil
	case service == "iam" && strings.HasPrefix(resource, "role/") && last != "":
		return principal{kind: "role", name: last, arn: arn}, nil
	case service == "sts" && strings.HasPrefix(resource, "assumed-role/") && len(segs) >= 2 && segs[1] != "":
		var out struct {
			Role struct {
				Arn string `json:"Arn"`
			} `json:"Role"`
		}
		args := []string{"iam", "get-role", "--role-name", segs[1], "--output", "json"}
		if err := r.call(opGetRole, args, &out); err != nil {
			return principal{}, err
		}
		if out.Role.Arn == "" {
			return principal{}, &model.AWSError{Op: opGetRole, Reason: model.AWSUnreachable}
		}
		return principal{kind: "role", name: segs[1], arn: out.Role.Arn}, nil
	}
	return principal{}, &model.AWSError{Op: op, Reason: model.AWSAccessDenied}
}

// SimulateBedrock ruft `aws iam simulate-principal-policy` für den Principal
// auf. Aktionen und ARN sind einzelne Argumente (keine Shell).
func (r Reader) SimulateBedrock(principalARN string, actions []string) ([]model.ActionDecision, error) {
	p, err := r.resolve(opSimulate, principalARN)
	if err != nil {
		return nil, err
	}
	args := []string{"iam", "simulate-principal-policy", "--policy-source-arn", p.arn, "--action-names"}
	args = append(args, actions...)
	args = append(args, "--output", "json")
	var out struct {
		Results []struct {
			Action   string `json:"EvalActionName"`
			Decision string `json:"EvalDecision"`
		} `json:"EvaluationResults"`
	}
	if err := r.call(opSimulate, args, &out); err != nil {
		return nil, err
	}
	res := make([]model.ActionDecision, 0, len(out.Results))
	for _, e := range out.Results {
		res = append(res, model.ActionDecision{Action: e.Action, Allowed: strings.EqualFold(e.Decision, "allowed")})
	}
	return res, nil
}

// PrincipalPolicies liefert die Statements aller angehängten und Inline-
// Policies des Principals (bei Benutzern zusätzlich seiner Gruppen).
func (r Reader) PrincipalPolicies(principalARN string) ([]model.IAMStatement, error) {
	p, err := r.resolve(opPolicies, principalARN)
	if err != nil {
		return nil, err
	}
	var out []model.IAMStatement
	add := func(sts []model.IAMStatement, err error) error {
		out = append(out, sts...)
		return err
	}
	if err := add(r.policiesOf(p.kind, p.name)); err != nil {
		return nil, err
	}
	if p.kind == "user" {
		var groups struct {
			Groups []struct {
				Name string `json:"GroupName"`
			} `json:"Groups"`
		}
		args := []string{"iam", "list-groups-for-user", "--user-name", p.name, "--output", "json"}
		if err := r.call(opPolicies, args, &groups); err != nil {
			return nil, err
		}
		for _, g := range groups.Groups {
			if err := add(r.policiesOf("group", g.Name)); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

// policiesOf liest angehängte und Inline-Policies einer Identität (kind =
// user | role | group).
func (r Reader) policiesOf(kind, name string) ([]model.IAMStatement, error) {
	flag := "--" + kind + "-name"
	var out []model.IAMStatement

	var attached struct {
		Policies []struct {
			ARN string `json:"PolicyArn"`
		} `json:"AttachedPolicies"`
	}
	if err := r.call(opPolicies, []string{"iam", "list-attached-" + kind + "-policies", flag, name, "--output", "json"}, &attached); err != nil {
		return nil, err
	}
	for _, a := range attached.Policies {
		var pol struct {
			Policy struct {
				Version string `json:"DefaultVersionId"`
			} `json:"Policy"`
		}
		if err := r.call(opPolicies, []string{"iam", "get-policy", "--policy-arn", a.ARN, "--output", "json"}, &pol); err != nil {
			return nil, err
		}
		var ver struct {
			Version struct {
				Document json.RawMessage `json:"Document"`
			} `json:"PolicyVersion"`
		}
		args := []string{"iam", "get-policy-version", "--policy-arn", a.ARN, "--version-id", pol.Policy.Version, "--output", "json"}
		if err := r.call(opPolicies, args, &ver); err != nil {
			return nil, err
		}
		out = append(out, statementsOf(ver.Version.Document)...)
	}

	var inline struct {
		Names []string `json:"PolicyNames"`
	}
	if err := r.call(opPolicies, []string{"iam", "list-" + kind + "-policies", flag, name, "--output", "json"}, &inline); err != nil {
		return nil, err
	}
	for _, n := range inline.Names {
		var doc struct {
			Document json.RawMessage `json:"PolicyDocument"`
		}
		args := []string{"iam", "get-" + kind + "-policy", flag, name, "--policy-name", n, "--output", "json"}
		if err := r.call(opPolicies, args, &doc); err != nil {
			return nil, err
		}
		out = append(out, statementsOf(doc.Document)...)
	}
	return out, nil
}

// statementsOf liest ein Policy-Dokument, das als JSON-Objekt oder (wie von
// der IAM-API) als URL-codierter String vorliegen kann. Nicht lesbare
// Dokumente ergeben ein Statement mit Dynamic=true.
func statementsOf(raw json.RawMessage) []model.IAMStatement {
	doc := []byte(raw)
	if len(doc) > 0 && doc[0] == '"' {
		var s string
		if err := json.Unmarshal(doc, &s); err != nil {
			return []model.IAMStatement{{Dynamic: true}}
		}
		if dec, err := url.QueryUnescape(s); err == nil {
			s = dec
		}
		doc = []byte(s)
	}
	sts, ok := iampolicy.Statements(doc, 0)
	if !ok {
		return []model.IAMStatement{{Dynamic: true}}
	}
	return sts
}
