package patch

import "testing"

func TestExtractOpsFile(t *testing.T) {
	text := "Here is the patch:\n```json\n{\n  \"reason\": \"repair\",\n  \"ops\": [{\"path\": \"a.txt\", \"old\": \"hi\", \"new\": \"hello\"}]\n}\n```"
	ops, err := ExtractOpsFile(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops.Ops) != 1 || ops.Ops[0].Path != "a.txt" {
		t.Fatalf("unexpected ops: %+v", ops)
	}
}
