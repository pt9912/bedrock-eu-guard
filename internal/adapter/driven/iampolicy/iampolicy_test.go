package iampolicy_test

import (
	"reflect"
	"testing"

	"github.com/pt9912/bedrock-eu-guard/internal/adapter/driven/iampolicy"
	"github.com/pt9912/bedrock-eu-guard/internal/hexagon/core/model"
)

func TestStatements(t *testing.T) {
	doc := `{"Version":"2012-10-17","Statement":[
	  {"Effect":"Allow","Action":["bedrock:InvokeModel","bedrock:Converse"],"Resource":"*",
	   "Condition":{"StringEquals":{"aws:RequestedRegion":["eu-central-1","eu-west-1"]}}},
	  {"Effect":"Allow","Action":"bedrock:*","Resource":["arn:aws:bedrock:eu-central-1::foundation-model/*"],
	   "Condition":{"StringLike":{"AWS:requestedRegion":"eu-*"}}},
	  {"Effect":"Deny","NotAction":["s3:*"],"Resource":"*"},
	  {"Effect":"Allow","Action":"s3:Get*","NotResource":"arn:aws:s3:::x"},
	  {"Effect":"Allow","Action":"sts:AssumeRole","Resource":"*","Condition":{"StringEquals":{"sts:ExternalId":"abc"}}}
	]}`
	got, ok := iampolicy.Statements([]byte(doc), 7)
	if !ok || len(got) != 5 {
		t.Fatalf("ok=%v, %d Statements: %+v", ok, len(got), got)
	}
	want0 := model.IAMStatement{Effect: "Allow", Actions: []string{"bedrock:InvokeModel", "bedrock:Converse"}, Resources: []string{"*"},
		HasRegionCondition: true, ConditionRegions: []string{"eu-central-1", "eu-west-1"}, Line: 7}
	if !reflect.DeepEqual(got[0], want0) {
		t.Errorf("s0: %+v", got[0])
	}
	if !got[1].HasRegionCondition || !reflect.DeepEqual(got[1].ConditionRegions, []string{"eu-*"}) || got[1].Resources[0] != "arn:aws:bedrock:eu-central-1::foundation-model/*" {
		t.Errorf("s1: %+v", got[1])
	}
	if !got[2].Dynamic || got[2].Effect != "Deny" || len(got[2].Actions) != 0 {
		t.Errorf("s2 (NotAction): %+v", got[2])
	}
	if !got[3].Dynamic {
		t.Errorf("s3 (NotResource): %+v", got[3])
	}
	if got[4].HasRegionCondition || got[4].Dynamic {
		t.Errorf("s4 (andere Condition): %+v", got[4])
	}
}

func TestSingleStatementObject(t *testing.T) {
	got, ok := iampolicy.Statements([]byte(`{"Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}`), 0)
	if !ok || len(got) != 1 || got[0].Actions[0] != "*" {
		t.Errorf("ok=%v, %+v", ok, got)
	}
}

func TestUnreadable(t *testing.T) {
	for _, in := range []string{"", "kein json", "{}", `{"Statement":[]}`, `[1]`} {
		got, ok := iampolicy.Statements([]byte(in), 0)
		if in == `{"Statement":[]}` {
			if !ok || len(got) != 0 {
				t.Errorf("%q: leere Liste ist gültig, got %+v ok=%v", in, got, ok)
			}
			continue
		}
		if ok {
			t.Errorf("%q: ok=true, %+v", in, got)
		}
	}
	// kaputtes einzelnes Statement → Dynamic
	got, ok := iampolicy.Statements([]byte(`{"Statement":[{"Effect":1}]}`), 3)
	if !ok || len(got) != 1 || !got[0].Dynamic || got[0].Line != 3 {
		t.Errorf("got %+v ok=%v", got, ok)
	}
}
