package main

import (
	"math"
	"testing"
	"github.com/EdlinOrg/prominentcolor"
)

func TestRGBToX(t *testing.T) {
	var color1 prominentcolor.ColorItem

	color1.Color.R = 30
	color1.Color.G = 60
	color1.Color.B = 90

	x,_,_ := RGBtoXYZ(color1)

	actual := math.Floor(x*100)/100
	expected := 3.99

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}


func TestRGBToY(t *testing.T) {
	var color1 prominentcolor.ColorItem

	color1.Color.R = 30
	color1.Color.G = 60
	color1.Color.B = 90

	_,y,_ := RGBtoXYZ(color1)

	actual := math.Floor(y*100)/100
	expected := 4.24

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}


func TestRGBToZ(t *testing.T) {
	var color1 prominentcolor.ColorItem

	color1.Color.R = 30
	color1.Color.G = 60
	color1.Color.B = 90

	_,_,z := RGBtoXYZ(color1)

	actual := math.Floor(z*100)/100
	expected := 10.28

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}

func TestXYZToL(t *testing.T) {
	x := 0.4
	y := 0.5
	z := 0.6

	l,_,_ := XYZToLAB(x,y,z)

	actual := math.Floor(l*100)/100
	expected := 4.51

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}

func TestXYZToA(t *testing.T) {
	x := 0.4
	y := 0.5
	z := 0.6

	_,a,_ := XYZToLAB(x,y,z)

	actual := math.Floor(a*100)/100
	expected := -3.09

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}

func TestXYZToB(t *testing.T) {
	x := 0.4
	y := 0.5
	z := 0.6

	_,_,b := XYZToLAB(x,y,z)

	actual := math.Floor(b*100)/100
	expected := -0.8

	if actual != expected {
		t.Errorf("got %f, wanted %f", actual, expected)
	}
}

func TestDeltaEZero(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem

	color1.Color.R = 0
	color1.Color.G = 0
	color1.Color.B = 0

	color2.Color.R = 0
	color2.Color.G = 0
	color2.Color.B = 0

	actual := betterSimilarColor(color1, color2)
	expected := 0

	if int(actual) != expected {
		t.Errorf("got %f, wanted %d", actual, expected)
	}
}

func TestDeltaEFull(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem

	color1.Color.R = 255
	color1.Color.G = 255
	color1.Color.B = 255

	color2.Color.R = 0
	color2.Color.G = 0
	color2.Color.B = 0

	actual := betterSimilarColor(color1, color2)
	expected := 100

	if int(actual) != expected {
		t.Errorf("got %f, wanted %d", actual, expected)
	} }


func TestDeltaEDifferent(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem

	color1.Color.R = 120
	color1.Color.G = 240
	color1.Color.B = 60

	color2.Color.R = 120
	color2.Color.G = 20
	color2.Color.B = 60

	actual := betterSimilarColor(color1, color2)
	expected := 76

	if int(actual) != expected {
		t.Errorf("got %f, wanted %d", actual, expected)
	}
}

func TestDeltaESimilar(t *testing.T) {
	var color1 prominentcolor.ColorItem
	var color2 prominentcolor.ColorItem

	color1.Color.R = 30
	color1.Color.G = 60
	color1.Color.B = 90

	color2.Color.R = 10
	color2.Color.G = 20
	color2.Color.B = 50

	actual := betterSimilarColor(color1, color2)
	expected := 18

	if int(actual) != expected {
		t.Errorf("got %f, wanted %d", actual, expected)
	}
}
