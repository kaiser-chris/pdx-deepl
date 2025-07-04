package pdx

import (
	"testing"
)

func TestEscaping(t *testing.T) {
	base := `@bur![Nbsp][concept_bureaucracy] Cost @innovation! $another_loc_key$ of\n\t $yet_another_loc_key$#bold [GetInstitutionType('institution_gate_monster_hunting').GetName]#!`
	escaped := escape(base)
	expected := `<ignore>@bur!</ignore><ignore>[Nbsp]</ignore><ignore>[concept_bureaucracy]</ignore> Cost <ignore>@innovation!</ignore> <ignore>$another_loc_key$</ignore> of<ignore>\n</ignore><ignore>\t</ignore> <ignore>$yet_another_loc_key$</ignore><ignore>#bold </ignore><ignore>[GetInstitutionType('institution_gate_monster_hunting').GetName]</ignore><ignore>#!</ignore>`
	if escaped != expected {
		t.Logf("Escaping did not work: %s != %s", expected, escaped)
		t.Fail()
	}
}
