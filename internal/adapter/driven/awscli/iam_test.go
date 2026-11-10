package awscli

import (
	"errors"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/port/driven"
)

var _ driven.IAMReader = Reader{}

const (
	roleARN    = "arn:aws:iam::123456789012:role/service-role/dev-role"
	assumedARN = "arn:aws:sts::123456789012:assumed-role/dev-role/session-1"
	userARN    = "arn:aws:iam::123456789012:user/team/alice"
)

func readLog(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

func TestSimulateBedrockResolvesAssumedRole(t *testing.T) {
	bin, log := fakeAWS(t, `echo "$@" >> "$LOG"
case "$2" in
  get-role) echo '{"Role":{"Arn":"`+roleARN+`"}}' ;;
  simulate-principal-policy) echo '{"EvaluationResults":[{"EvalActionName":"bedrock:InvokeModel","EvalDecision":"allowed"},{"EvalActionName":"bedrock:Converse","EvalDecision":"implicitDeny"},{"EvalActionName":"bedrock:ListInferenceProfiles","EvalDecision":"explicitDeny"}]}' ;;
esac`)
	got, err := New(bin, "eu-central-1").SimulateBedrock(assumedARN, []string{"bedrock:InvokeModel", "bedrock:Converse", "bedrock:ListInferenceProfiles"})
	if err != nil {
		t.Fatal(err)
	}
	want := []model.ActionDecision{
		{Action: "bedrock:InvokeModel", Allowed: true},
		{Action: "bedrock:Converse", Allowed: false},
		{Action: "bedrock:ListInferenceProfiles", Allowed: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v", got)
	}
	lines := readLog(t, log)
	if len(lines) != 2 ||
		lines[0] != "iam get-role --role-name dev-role --output json" ||
		lines[1] != "iam simulate-principal-policy --policy-source-arn "+roleARN+
			" --action-names bedrock:InvokeModel bedrock:Converse bedrock:ListInferenceProfiles --output json" {
		t.Errorf("Aufrufe: %q", lines)
	}
}

func TestSimulateBedrockDirectARNsNeedNoLookup(t *testing.T) {
	for _, arn := range []string{roleARN, userARN} {
		bin, log := fakeAWS(t, `echo "$@" >> "$LOG"; echo '{"EvaluationResults":[]}'`)
		if _, err := New(bin, "").SimulateBedrock(arn, []string{"bedrock:InvokeModel"}); err != nil {
			t.Fatal(err)
		}
		lines := readLog(t, log)
		if len(lines) != 1 || !strings.Contains(lines[0], "--policy-source-arn "+arn+" ") {
			t.Errorf("%s: Aufrufe %q", arn, lines)
		}
	}
}

func TestUnresolvablePrincipalsAreAccessDenied(t *testing.T) {
	bin, log := fakeAWS(t, `echo "$@" >> "$LOG"`)
	for _, arn := range []string{"arn:aws:iam::123456789012:root", "arn:aws:sts::123456789012:federated-user/bob", "kein-arn", ""} {
		_, err := New(bin, "").SimulateBedrock(arn, []string{"bedrock:InvokeModel"})
		var ae *model.AWSError
		if !errors.As(err, &ae) || ae.Reason != model.AWSAccessDenied {
			t.Errorf("%q: %v", arn, err)
		}
		if _, err := New(bin, "").PrincipalPolicies(arn); err == nil {
			t.Errorf("%q: PrincipalPolicies ohne Fehler", arn)
		}
	}
	if _, err := os.Stat(log); err == nil {
		t.Error("für nicht auflösbare Principals darf die CLI nicht aufgerufen werden")
	}
}

func TestPrincipalPoliciesRole(t *testing.T) {
	// Verwaltete Policy: Dokument als Objekt; Inline-Policy: URL-codierter String.
	inline := url.QueryEscape(`{"Statement":[{"Effect":"Allow","Action":"bedrock:*","Resource":"*","Condition":{"StringEquals":{"aws:RequestedRegion":"eu-central-1"}}}]}`)
	bin, log := fakeAWS(t, `echo "$@" >> "$LOG"
case "$2" in
  list-attached-role-policies) echo '{"AttachedPolicies":[{"PolicyArn":"arn:aws:iam::123456789012:policy/p1"}]}' ;;
  get-policy) echo '{"Policy":{"DefaultVersionId":"v3"}}' ;;
  get-policy-version) echo '{"PolicyVersion":{"Document":{"Statement":[{"Effect":"Allow","Action":["bedrock:InvokeModel"],"Resource":"arn:aws:bedrock:eu-central-1::foundation-model/*"}]}}}' ;;
  list-role-policies) echo '{"PolicyNames":["inline1"]}' ;;
  get-role-policy) echo '{"PolicyDocument":"`+inline+`"}' ;;
esac`)
	got, err := New(bin, "").PrincipalPolicies(roleARN)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("Statements: %+v", got)
	}
	if got[0].Effect != "Allow" || !reflect.DeepEqual(got[0].Actions, []string{"bedrock:InvokeModel"}) || got[0].HasRegionCondition {
		t.Errorf("verwaltet: %+v", got[0])
	}
	if !reflect.DeepEqual(got[1].Actions, []string{"bedrock:*"}) || !got[1].HasRegionCondition || !reflect.DeepEqual(got[1].ConditionRegions, []string{"eu-central-1"}) {
		t.Errorf("inline (URL-decodiert): %+v", got[1])
	}
	want := []string{
		"iam list-attached-role-policies --role-name dev-role --output json",
		"iam get-policy --policy-arn arn:aws:iam::123456789012:policy/p1 --output json",
		"iam get-policy-version --policy-arn arn:aws:iam::123456789012:policy/p1 --version-id v3 --output json",
		"iam list-role-policies --role-name dev-role --output json",
		"iam get-role-policy --role-name dev-role --policy-name inline1 --output json",
	}
	if lines := readLog(t, log); !reflect.DeepEqual(lines, want) {
		t.Errorf("Aufrufe:\n%q\nwant\n%q", lines, want)
	}
}

func TestPrincipalPoliciesUserIncludesGroups(t *testing.T) {
	bin, log := fakeAWS(t, `echo "$@" >> "$LOG"
case "$2" in
  list-groups-for-user) echo '{"Groups":[{"GroupName":"devs"}]}' ;;
  list-attached-user-policies|list-attached-group-policies) echo '{"AttachedPolicies":[]}' ;;
  list-user-policies) echo '{"PolicyNames":[]}' ;;
  list-group-policies) echo '{"PolicyNames":["g1"]}' ;;
  get-group-policy) echo '{"PolicyDocument":{"Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}}' ;;
esac`)
	got, err := New(bin, "").PrincipalPolicies(userARN)
	if err != nil || len(got) != 1 || !reflect.DeepEqual(got[0].Actions, []string{"*"}) {
		t.Fatalf("got %+v, %v", got, err)
	}
	joined := strings.Join(readLog(t, log), "\n")
	for _, want := range []string{
		"iam list-attached-user-policies --user-name alice --output json",
		"iam list-groups-for-user --user-name alice --output json",
		"iam list-group-policies --group-name devs --output json",
		"iam get-group-policy --group-name devs --policy-name g1 --output json",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("Aufruf fehlt: %s\n%s", want, joined)
		}
	}
}

func TestUnreadableDocumentsAreDynamic(t *testing.T) {
	bin, _ := fakeAWS(t, `case "$2" in
  list-attached-role-policies) echo '{"AttachedPolicies":[]}' ;;
  list-role-policies) echo '{"PolicyNames":["x"]}' ;;
  get-role-policy) echo '{"PolicyDocument":"kein%20json"}' ;;
esac`)
	got, err := New(bin, "").PrincipalPolicies(roleARN)
	if err != nil || len(got) != 1 || !got[0].Dynamic {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestIAMErrorsAreTypedAndLeakFree(t *testing.T) {
	bin, _ := fakeAWS(t, "echo 'An error occurred (AccessDenied) when calling the SimulatePrincipalPolicy operation "+leak+"' >&2; exit 255")
	_, serr := New(bin, "").SimulateBedrock(roleARN, []string{"bedrock:InvokeModel"})
	var ae *model.AWSError
	if !errors.As(serr, &ae) || ae.Op != "iam:SimulatePrincipalPolicy" || ae.Reason != model.AWSAccessDenied || strings.Contains(serr.Error(), leak) {
		t.Errorf("simulate: %v", serr)
	}
	_, perr := New(bin, "").PrincipalPolicies(roleARN)
	if !errors.As(perr, &ae) || ae.Op != "iam:GetPrincipalPolicies" || ae.Reason != model.AWSAccessDenied || strings.Contains(perr.Error(), leak) {
		t.Errorf("policies: %v", perr)
	}
	_, gerr := New(bin, "").SimulateBedrock(assumedARN, []string{"bedrock:InvokeModel"})
	if !errors.As(gerr, &ae) || ae.Op != "iam:GetRole" || strings.Contains(gerr.Error(), leak) {
		t.Errorf("get-role: %v", gerr)
	}
}
