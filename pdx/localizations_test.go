package pdx

import (
	"testing"
)

func TestMarkedCommentRemoval(t *testing.T) {
	base := "l_english:\r\n#! some comment\r\n   #! some other comment\r\n some_loc: \"loc content\"\n some_other_loc: \"loc content\""
	escaped := regexRemoveComment.ReplaceAllString(base, "")
	expected := "l_english:\n some_loc: \"loc content\"\n some_other_loc: \"loc content\""
	if escaped != expected {
		t.Log(expected)
		t.Log(escaped)
		t.Logf("Comment removal did not work: %s != %s", expected, escaped)
		t.Fail()
	}
}
