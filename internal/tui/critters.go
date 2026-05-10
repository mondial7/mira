package tui

import "strings"

// Critters — a cat and a dog that live in the bottom-right corner of the
// view. Their idle animation cycles slowly (a sleepy blink every few
// ticks); their eyes track the direction of the user's last horizontal
// cursor move. Frames are deliberately the same width so the bounding
// box never reflows.

// CritterHeight is how many vertical lines the critter strip occupies.
const CritterHeight = 3

// critterWidth is the rendered width of a single animal (display columns,
// not bytes). All frame strings must be exactly this many display columns
// wide on every line.
const critterWidth = 8

// Cat frames — 8 columns × 3 rows.
var (
	catCenter = [3]string{
		" /\\_/\\  ",
		"( o.o ) ",
		" > ^ <  ",
	}
	catBlink = [3]string{
		" /\\_/\\  ",
		"( -.- ) ",
		" > ^ <  ",
	}
	catLeft = [3]string{
		" /\\_/\\  ",
		"(<.<  ) ",
		" > ^ <  ",
	}
	catRight = [3]string{
		" /\\_/\\  ",
		"(  >.>) ",
		" > ^ <  ",
	}
	catPurr = [3]string{
		" /\\_/\\  ",
		"( ^.^ ) ",
		" > ~ <  ",
	}
)

// Dog frames — 8 columns × 3 rows.
var (
	dogCenter = [3]string{
		" /^.^\\  ",
		"( o.o ) ",
		" v=-=v  ",
	}
	dogBlink = [3]string{
		" /^.^\\  ",
		"( -.- ) ",
		" v=-=v  ",
	}
	dogLeft = [3]string{
		" /^.^\\  ",
		"(<.<  ) ",
		" v=-=v  ",
	}
	dogRight = [3]string{
		" /^.^\\  ",
		"(  >.>) ",
		" v=-=v  ",
	}
	dogWag = [3]string{
		" /^.^\\  ",
		"( o.o ) ",
		" v=~=v  ",
	}
)

// pickFrames returns the cat + dog frames to render this tick. The
// ambient blink/wag cycle takes priority for a few specific tick slots
// (so the critters always feel alive); on every other tick they look in
// the direction lookDir indicates.
func pickFrames(animFrame, lookDir int) (cat, dog [3]string) {
	switch animFrame % 12 {
	case 3:
		return catBlink, dogCenter
	case 4:
		return catCenter, dogBlink
	case 9:
		return catPurr, dogWag
	}
	switch {
	case lookDir < 0:
		return catLeft, dogLeft
	case lookDir > 0:
		return catRight, dogRight
	default:
		return catCenter, dogCenter
	}
}

// renderCritters builds the 3-line block that holds the critters, padded
// with spaces on the left so the block right-aligns to `width` columns.
// Returns three styled lines joined by newlines.
func renderCritters(width, animFrame, lookDir int) string {
	cat, dog := pickFrames(animFrame, lookDir)

	const gap = "  " // two spaces between critters
	totalWidth := critterWidth*2 + len(gap)
	leftPad := width - totalWidth
	if leftPad < 0 {
		leftPad = 0
	}
	pad := strings.Repeat(" ", leftPad)

	var lines [CritterHeight]string
	for i := 0; i < CritterHeight; i++ {
		row := critterStyle.Render(cat[i] + gap + dog[i])
		lines[i] = pad + row
	}
	return strings.Join(lines[:], "\n")
}
