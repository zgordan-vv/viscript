package main

import (
	//"fmt"
	"github.com/go-gl/gl/v2.1/gl"
)

/*
mouse position updates use pixels, so the smallest drag motions will be
a jump of at least 1 pixel height.
the ratio of that height / LenOfVoid (bar representing the page size),
compared to the void/offscreen length of the text body,
gives us the jump size in scrolling through the text body
*/

type ScrollBar struct {
	PosX            float32
	PosY            float32
	LenOfBar        float32
	LenOfVoid       float32 // length of the negative space representing the length of entire document
	LenOfOffscreenY float32
	ScrollDistY     float32 // distance/offset from top of document (negative number cuz Y goes down screen)
}

func (bar *ScrollBar) UpdateSize(tp *TextPanel) {
	hei := textRend.CharHei * float32(tp.NumCharsY) /* height of panel */

	if /* content smaller than screen */ len(tp.Body) <= tp.NumCharsY {
		// NO BAR
		bar.LenOfBar = 0
		bar.LenOfVoid = hei
		bar.LenOfOffscreenY = 0
	} else {
		bar.LenOfBar = float32(tp.NumCharsY) / float32(len(tp.Body)) * hei
		bar.LenOfVoid = hei - bar.LenOfBar
		bar.LenOfOffscreenY = float32(len(tp.Body)-tp.NumCharsY) * textRend.CharHei
	}
}

func (bar *ScrollBar) DragHandleContainsMouseCursor() bool {
	if curs.MouseGlY <= bar.PosY && curs.MouseGlY >= bar.PosY-bar.LenOfBar {
		if curs.MouseGlX <= bar.PosX+textRend.CharWid && curs.MouseGlX >= bar.PosX {
			return true
		}
	}

	return false
}

func (bar *ScrollBar) ScrollThisMuch(tp *TextPanel, incrementY float32) {
	bar.PosY -= incrementY

	if bar.PosY < tp.Bottom+bar.LenOfBar {
		bar.PosY = tp.Bottom + bar.LenOfBar
	}
	if bar.PosY > tp.Top {
		bar.PosY = tp.Top
	}

	bar.ScrollDistY -= incrementY / bar.LenOfVoid * bar.LenOfOffscreenY

	if bar.ScrollDistY > 0 {
		bar.ScrollDistY = 0
	}

	if bar.ScrollDistY < -bar.LenOfOffscreenY {
		bar.ScrollDistY = -bar.LenOfOffscreenY
	}
}

func (bar *ScrollBar) DrawVertical(atlasX, atlasY float32) {
	rad := textRend.ScreenRad
	sp := textRend.UvSpan
	u := float32(atlasX) * sp
	v := float32(atlasY) * sp

	top := bar.PosY                 //rad - 1
	bott := bar.PosY - bar.LenOfBar //-rad + 1

	gl.Normal3f(0, 0, 1)

	// bottom left   0, 1
	gl.TexCoord2f(u, v+sp)
	gl.Vertex3f(rad-textRend.CharWid, bott, 0)

	// bottom right   1, 1
	gl.TexCoord2f(u+sp, v+sp)
	gl.Vertex3f(rad, bott, 0)

	// top right   1, 0
	gl.TexCoord2f(u+sp, v)
	gl.Vertex3f(rad, top, 0)

	// top left   0, 0
	gl.TexCoord2f(u, v)
	gl.Vertex3f(rad-textRend.CharWid, top, 0)
}
