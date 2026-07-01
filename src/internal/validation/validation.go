package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
)

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Email ověří formát e-mailové adresy. Odmítne řetězce delší než 254 znaků
// nebo neodpovídající RFC 5322 (kontroluje net/mail.ParseAddress).
func Email(s string) error {
	if len(s) > 254 {
		return fmt.Errorf("e-mail nesmí přesáhnout 254 znaků")
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Errorf("neplatný formát e-mailové adresy")
	}
	return nil
}

// Length ověří délku řetězce (v Unicode code points) vůči zadaným mezím.
// min=0 přeskočí dolní kontrolu, max=0 přeskočí horní kontrolu.
func Length(s, field string, min, max int) error {
	n := len([]rune(s))
	if min > 0 && n < min {
		return fmt.Errorf("pole %s musí mít alespoň %d znaků", field, min)
	}
	if max > 0 && n > max {
		return fmt.Errorf("pole %s nesmí přesáhnout %d znaků", field, max)
	}
	return nil
}

// HexColor ověří, že řetězec je validní barva ve formátu #RRGGBB.
func HexColor(s string) error {
	if !hexColorRe.MatchString(s) {
		return fmt.Errorf("neplatný formát barvy, očekáváno #RRGGBB")
	}
	return nil
}

// Enum ověří, že hodnota je jednou z povolených možností.
func Enum(val, field string, allowed ...string) error {
	for _, a := range allowed {
		if val == a {
			return nil
		}
	}
	return fmt.Errorf("pole %s musí být jednou z hodnot: %s", field, strings.Join(allowed, ", "))
}
