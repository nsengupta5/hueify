package main

import "github.com/EdlinOrg/prominentcolor"
import "testing"

func TestSimilarColorWithExactOpposite(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem
	color1.Color.R = 0
	color1.Color.G = 0
	color1.Color.B = 0
	color2.Color.R = 255
	color2.Color.G = 255
	color2.Color.B = 255
	actual := betterSimilarColor(color1, color2)
	var expected int = 100

	if int(actual) != expected {
		t.Fail()
	}
}

func TestSimilarColorWithExactSame(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem
	color1.Color.R = 0
	color1.Color.G = 0
	color1.Color.B = 0
	color2.Color.R = 0
	color2.Color.G = 0
	color2.Color.B = 0
	actual := betterSimilarColor(color1, color2)
	var expected int = 0

	if int(actual) != expected {
		t.Fail()
	}
}
