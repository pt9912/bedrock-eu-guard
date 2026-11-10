package secretkey_test

import (
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/secretkey"
)

func TestIsSecret(t *testing.T) {
	for _, k := range []string{
		"ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"AWS_SESSION_TOKEN", "AWS_BEARER_TOKEN_BEDROCK", "DB_PASSWORD", "GCP_CREDENTIALS",
		"some_private_key", "MY_KEY", "apiKeyHelper",
	} {
		if !secretkey.IsSecret(k) {
			t.Errorf("%s soll Secret sein", k)
		}
	}
	for _, k := range []string{"AWS_REGION", "AWS_DEFAULT_REGION", "CLAUDE_CODE_USE_BEDROCK", "ANTHROPIC_BASE_URL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "HOME"} {
		if secretkey.IsSecret(k) {
			t.Errorf("%s soll kein Secret sein", k)
		}
	}
}
