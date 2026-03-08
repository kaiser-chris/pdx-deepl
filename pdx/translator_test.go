package pdx

import (
	"testing"
)

func TestEscaping(t *testing.T) {
	base := `@bur![Nbsp][concept_bureaucracy] Cost @innovation! $another_loc_key$ of\n\t $yet_another_loc_key$#bold;purple [GetInstitutionType('institution_gate_monster_hunting').GetName]#!`
	escaped := escape(base)
	expected := `<ignore>@bur!</ignore><ignore>[Nbsp]</ignore><ignore>[concept_bureaucracy]</ignore> Cost <ignore>@innovation!</ignore> <ignore>$another_loc_key$</ignore> of<ignore>\n</ignore><ignore>\t</ignore> <ignore>$yet_another_loc_key$</ignore><ignore>#bold;purple </ignore><ignore>[GetInstitutionType('institution_gate_monster_hunting').GetName]</ignore><ignore>#!</ignore>`
	if escaped != expected {
		t.Logf("Escaping did not work: %s != %s", expected, escaped)
		t.Fail()
	}
}

func TestEscapingBug1(t *testing.T) {
	base := `#header $Name$ progress#!\n[AddTextIf(NotEqualTo_CFixedPoint('(CFixedPoint)$TOTAL_VALUE$','(CFixedPoint)0'), '$TOTAL_VALUE|-=1$ average monthly progress currently\n$TOOLTIP_DELIMITER$\nThis is due to:' )][AddTextIf(Not(StringIsEmpty('$WEEKLY_DESC$')), '\n#title $WEEKLY_VALUE|-=2$ weekly change from:#! #indent_newline:3 \n$WEEKLY_DESC$#!\n')][AddTextIf(Not(StringIsEmpty('$MONTHLY_DESC$')), '\n#title $MONTHLY_VALUE|-=1$ monthly change from:#! #indent_newline:3 \n$MONTHLY_DESC$#!\n')][AddTextIf(Not(StringIsEmpty('$YEARLY_DESC$')), '\n#title $YEARLY_VALUE|-=1$ yearly change from:#! #indent_newline:3 \n$YEARLY_DESC$#!\n')]`
	escaped := escape(base)
	expected := `<ignore>#header </ignore><ignore>$Name$</ignore> progress<ignore>#!</ignore><ignore>\n</ignore><ignore>[AddTextIf(NotEqualTo_CFixedPoint('(CFixedPoint)<ignore>$TOTAL_VALUE$</ignore>','(CFixedPoint)0'), '<ignore>$TOTAL_VALUE|-=1$</ignore> average monthly progress currently<ignore>\n</ignore><ignore>$TOOLTIP_DELIMITER$</ignore><ignore>\n</ignore>This is due to:' )]</ignore><ignore>[AddTextIf(Not(StringIsEmpty('<ignore>$WEEKLY_DESC$</ignore>')), '<ignore>\n</ignore><ignore>#title </ignore><ignore>$WEEKLY_VALUE|-=2$</ignore> weekly change from:<ignore>#!</ignore> #indent_newline:3 <ignore>\n</ignore><ignore>$WEEKLY_DESC$</ignore><ignore>#!</ignore><ignore>\n</ignore>')]</ignore><ignore>[AddTextIf(Not(StringIsEmpty('<ignore>$MONTHLY_DESC$</ignore>')), '<ignore>\n</ignore><ignore>#title </ignore><ignore>$MONTHLY_VALUE|-=1$</ignore> monthly change from:<ignore>#!</ignore> #indent_newline:3 <ignore>\n</ignore><ignore>$MONTHLY_DESC$</ignore><ignore>#!</ignore><ignore>\n</ignore>')]</ignore><ignore>[AddTextIf(Not(StringIsEmpty('<ignore>$YEARLY_DESC$</ignore>')), '<ignore>\n</ignore><ignore>#title </ignore><ignore>$YEARLY_VALUE|-=1$</ignore> yearly change from:<ignore>#!</ignore> #indent_newline:3 <ignore>\n</ignore><ignore>$YEARLY_DESC$</ignore><ignore>#!</ignore><ignore>\n</ignore>')]</ignore>`
	if escaped != expected {
		t.Logf("Escaping did not work: %s != %s", expected, escaped)
		t.Fail()
	}
}
