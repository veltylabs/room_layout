package tests

import (
	"testing"

	roomlayout "github.com/veltylabs/room_layout"
)

func TestArea_RectangleIsOneEntry(t *testing.T) {
	var cells []roomlayout.Cell
	for row := 1; row <= 23; row++ {
		for col := 7; col <= 17; col++ {
			cells = append(cells, roomlayout.Cell{Col: col, Row: row})
		}
	}
	if got := roomlayout.EncodeArea(cells); got != "7,1,11,23" {
		t.Fatalf("EncodeArea = %q, want %q", got, "7,1,11,23")
	}
}

func TestArea_LShapeRoundTrips(t *testing.T) {
	// ##.
	// ###
	in := []roomlayout.Cell{{Col: 1, Row: 1}, {Col: 2, Row: 1}, {Col: 1, Row: 2}, {Col: 2, Row: 2}, {Col: 3, Row: 2}}
	enc := roomlayout.EncodeArea(in)
	if enc != "1,1,2,2;3,2,1,1" {
		t.Fatalf("EncodeArea = %q", enc)
	}
	out, err := roomlayout.DecodeArea(enc)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(in) || roomlayout.EncodeArea(out) != enc {
		t.Fatalf("DecodeArea = %v", out)
	}
}

func TestArea_EmptyAndInvalid(t *testing.T) {
	if roomlayout.EncodeArea(nil) != "" {
		t.Fatal("empty area must encode as empty text")
	}
	if cells, err := roomlayout.DecodeArea(""); err != nil || cells != nil {
		t.Fatalf("DecodeArea(\"\") = %v, %v", cells, err)
	}
	for _, bad := range []string{"A1,B1", "1,1,2", "0,1,1,1", "1,1,1,1;"} {
		if _, err := roomlayout.DecodeArea(bad); err != roomlayout.ErrInvalidArea {
			t.Fatalf("DecodeArea(%q) err = %v, want ErrInvalidArea", bad, err)
		}
	}
}

func TestCell_StringParse(t *testing.T) {
	c := roomlayout.Cell{Col: 12, Row: 4}
	if c.String() != "12,4" {
		t.Fatalf("String = %q", c.String())
	}
	if got, ok := roomlayout.ParseCell("12,4"); !ok || got != c {
		t.Fatalf("ParseCell = %v, %v", got, ok)
	}
	if _, ok := roomlayout.ParseCell("L4"); ok {
		t.Fatal("letters are not a cell")
	}
}

func TestRects_LShapeIsTwo(t *testing.T) {
	in := []roomlayout.Cell{{Col: 1, Row: 1}, {Col: 2, Row: 1}, {Col: 1, Row: 2}, {Col: 2, Row: 2}, {Col: 3, Row: 2}}
	got := roomlayout.Rects(in)
	want := []roomlayout.Rect{{Col: 1, Row: 1, W: 2, H: 2}, {Col: 3, Row: 2, W: 1, H: 1}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Rects = %v, want %v", got, want)
	}
}
