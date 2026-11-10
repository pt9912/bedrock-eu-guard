package rules_test

import (
	"strings"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/rules"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/coretest/fakeenv"
)

var ident = &model.AWSIdentity{Account: "123456789012", ARN: "arn:aws:sts::123456789012:assumed-role/dev/x", UserID: "AROA"}

func TestAWSIdentity(t *testing.T) {
	cfg := model.DefaultConfig()
	f := sole(t, rules.CheckAWSIdentity(rules.AWSScan{Env: fakeenv.Env{}, Identity: ident}, cfg))
	if f.Severity != model.PASS || !strings.Contains(f.Message, "123456789012") || !strings.Contains(f.Message, "assumed-role/dev/x") {
		t.Errorf("pass: %+v", f)
	}
	for reason, wantMsg := range map[string]string{
		model.AWSNoCredentials: "Credentials",
		model.AWSAccessDenied:  "verweigert",
		model.AWSUnreachable:   "erreichbar",
	} {
		f := sole(t, rules.CheckAWSIdentity(rules.AWSScan{Env: fakeenv.Env{}, IdentityErr: &model.AWSError{Op: "sts:GetCallerIdentity", Reason: reason}}, cfg))
		if f.Severity != model.FAIL || f.Reason != reason || !strings.Contains(f.Message, wantMsg) || f.RuleID != rules.IDAWSIdentity {
			t.Errorf("%s: %+v", reason, f)
		}
	}
}

func TestAWSProfile(t *testing.T) {
	cfg := model.DefaultConfig()
	cases := []struct {
		name string
		s    rules.AWSScan
		sev  model.Severity
		rsn  string
	}{
		{"gesetzt+verwendbar", rules.AWSScan{Env: fakeenv.Env{"AWS_PROFILE": "dev-eu"}, Identity: ident}, model.PASS, model.ReasonProfileSet},
		{"nicht gesetzt", rules.AWSScan{Env: fakeenv.Env{}, Identity: ident}, model.WARN, model.ReasonProfileNotSet},
		{"leer", rules.AWSScan{Env: fakeenv.Env{"AWS_PROFILE": ""}, Identity: ident}, model.WARN, model.ReasonProfileNotSet},
		{"unbrauchbar", rules.AWSScan{Env: fakeenv.Env{"AWS_PROFILE": "x"}, IdentityErr: &model.AWSError{Reason: model.AWSNoCredentials}}, model.WARN, model.ReasonProfileUnusable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := sole(t, rules.CheckAWSProfile(c.s, cfg))
			if f.Severity != c.sev || f.Reason != c.rsn {
				t.Fatalf("got %+v", f)
			}
		})
	}
}

func TestAWSAccount(t *testing.T) {
	with := func(accts ...string) model.Config { c := model.DefaultConfig(); c.AllowedAccounts = accts; return c }
	if fs := rules.CheckAWSAccount(rules.AWSScan{Env: fakeenv.Env{}, Identity: ident}, model.DefaultConfig()); len(fs) != 0 {
		t.Errorf("leere Liste: %+v", fs)
	}
	if f := sole(t, rules.CheckAWSAccount(rules.AWSScan{Env: fakeenv.Env{}, Identity: ident}, with("999999999999", "123456789012"))); f.Severity != model.PASS || f.Reason != model.ReasonAccountAllowed {
		t.Errorf("erlaubt: %+v", f)
	}
	if f := sole(t, rules.CheckAWSAccount(rules.AWSScan{Env: fakeenv.Env{}, Identity: ident}, with("999999999999"))); f.Severity != model.FAIL || f.Reason != model.ReasonAccountNotAllowed {
		t.Errorf("nicht erlaubt: %+v", f)
	}
	if f := sole(t, rules.CheckAWSAccount(rules.AWSScan{Env: fakeenv.Env{}, IdentityErr: &model.AWSError{Reason: model.AWSUnreachable}}, with("123456789012"))); f.Severity != model.WARN || f.Reason != model.ReasonAccountUnchecked {
		t.Errorf("nicht prüfbar: %+v", f)
	}
}

func TestAuthMethod(t *testing.T) {
	cfg := model.DefaultConfig()
	reasons := func(env fakeenv.Env) map[string]model.Severity {
		out := map[string]model.Severity{}
		for _, f := range rules.CheckAuthMethod(rules.AWSScan{Env: env}, cfg) {
			out[f.Reason] = f.Severity
		}
		return out
	}
	cases := []struct {
		name string
		env  fakeenv.Env
		want map[string]model.Severity
	}{
		{"profile", fakeenv.Env{"AWS_PROFILE": "dev"}, map[string]model.Severity{model.ReasonAuthProfile: model.PASS}},
		{"temporaer", fakeenv.Env{"AWS_ACCESS_KEY_ID": "x", "AWS_SESSION_TOKEN": "y"}, map[string]model.Severity{model.ReasonAuthTemporary: model.PASS}},
		{"statisch", fakeenv.Env{"AWS_ACCESS_KEY_ID": "x"}, map[string]model.Severity{model.ReasonAuthStaticKeys: model.WARN}},
		{"statisch schlaegt Profile", fakeenv.Env{"AWS_ACCESS_KEY_ID": "x", "AWS_PROFILE": "dev"}, map[string]model.Severity{model.ReasonAuthStaticKeys: model.WARN}},
		{"bedrock api key", fakeenv.Env{"AWS_BEARER_TOKEN_BEDROCK": "t"}, map[string]model.Severity{model.ReasonAuthBedrockAPIKey: model.WARN}},
		{"anthropic", fakeenv.Env{"ANTHROPIC_API_KEY": "sk-x", "AWS_PROFILE": "dev"}, map[string]model.Severity{model.ReasonAuthAnthropicKey: model.FAIL, model.ReasonAuthProfile: model.PASS}},
		{"nichts", fakeenv.Env{}, map[string]model.Severity{model.ReasonAuthNoneDetected: model.WARN}},
		{"leere Werte zaehlen nicht", fakeenv.Env{"AWS_ACCESS_KEY_ID": "", "ANTHROPIC_API_KEY": ""}, map[string]model.Severity{model.ReasonAuthNoneDetected: model.WARN}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := reasons(c.env)
			if len(got) != len(c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
			for k, v := range c.want {
				if got[k] != v {
					t.Fatalf("got %v want %v", got, c.want)
				}
			}
		})
	}
}

func TestAuthMethodNeverLeaksValues(t *testing.T) {
	env := fakeenv.Env{"ANTHROPIC_API_KEY": "sk-geheim", "AWS_ACCESS_KEY_ID": "AKIAGEHEIM", "AWS_BEARER_TOKEN_BEDROCK": "tok-geheim", "AWS_SESSION_TOKEN": "sess-geheim"}
	for _, f := range rules.CheckAuthMethod(rules.AWSScan{Env: env}, model.DefaultConfig()) {
		if strings.Contains(f.Message+f.Recommendation+f.Subject, "geheim") || strings.Contains(f.Message, "AKIA") {
			t.Errorf("leak: %+v", f)
		}
	}
}
