package hypervisor

import (
	"fmt"
	"github.com/corpusc/viscript/app"
	"github.com/corpusc/viscript/cGfx"
	"github.com/corpusc/viscript/gfx"
	"github.com/corpusc/viscript/hypervisor/input/mouse"
	"github.com/corpusc/viscript/tree"
	"github.com/corpusc/viscript/ui"
	"github.com/go-gl/gl/v2.1/gl"
	//"math"
)

type Terminal struct {
	FractionOfStrip float32 // fraction of the parent PanelStrip (in 1 dimension)
	CursX           int     // current cursor/insert position (in character grid cells/units)
	CursY           int
	MouseX          int // current mouse position in character grid space (units/cells)
	MouseY          int
	IsEditable      bool // editing is hardwired to TextBodies[0], but we probably never want
	// to edit text unless the whole panel is dedicated to just one TextBody (& no graphical trees)
	Whole      *app.Rectangle    // the whole panel, including chrome (title bar & scroll bars)
	Content    *app.PicRectangle // viewport into virtual space, subset of the Whole rect
	Selection  *ui.SelectionRange
	BarHori    *ui.ScrollBar // horizontal
	BarVert    *ui.ScrollBar // vertical
	TextBodies [][]string
	TextColors []*cGfx.ColorSpot
	Trees      []*tree.Tree
}

func (t *Terminal) Init() {
	fmt.Printf("Terminal.Init()\n")

	t.TextBodies = append(t.TextBodies, []string{})

	t.Selection = &ui.SelectionRange{}
	t.Selection.Init()

	// scrollbars
	t.BarHori = &ui.ScrollBar{IsHorizontal: true}
	t.BarVert = &ui.ScrollBar{}
	t.BarHori.Rect = &app.PicRectangle{0, 0, 0, cGfx.Pic_GradientBorder, &app.Rectangle{}}
	t.BarVert.Rect = &app.PicRectangle{0, 0, 0, cGfx.Pic_GradientBorder, &app.Rectangle{}}

	t.SetSize()
}

func (t *Terminal) SetSize() {
	fmt.Printf("Terminal.SetSize()\n")

	t.Whole = &app.Rectangle{
		cGfx.CanvasExtents.Y - cGfx.CharHei,
		cGfx.CanvasExtents.X,
		-cGfx.CanvasExtents.Y,
		-cGfx.CanvasExtents.X}

	if t.FractionOfStrip == runOutputTerminalFrac { // FIXME: this is hardwired for one use case for now
		t.Whole.Top = t.Whole.Bottom + t.Whole.Height()*t.FractionOfStrip
	} else {
		t.Whole.Bottom = t.Whole.Bottom + t.Whole.Height()*runOutputTerminalFrac
	}

	t.Content = &app.PicRectangle{0, 0, 0, cGfx.Pic_GradientBorder, &app.Rectangle{}}
	t.Content.Rect.Top = t.Whole.Top
	t.Content.Rect.Right = t.Whole.Right - ui.ScrollBarThickness
	t.Content.Rect.Bottom = t.Whole.Bottom + ui.ScrollBarThickness
	t.Content.Rect.Left = t.Whole.Left

	// set scrollbars' upper left corners
	t.BarHori.Rect.Rect.Left = t.Whole.Left
	t.BarHori.Rect.Rect.Top = t.Content.Rect.Bottom
	t.BarVert.Rect.Rect.Left = t.Content.Rect.Right
	t.BarVert.Rect.Rect.Top = t.Whole.Top
}

func (t *Terminal) RespondToMouseClick() {
	Focused = t

	// diffs/deltas from home position of panel (top left corner)
	glDeltaFromHome := app.Vec2F{
		mouse.GlX - t.Whole.Left,
		mouse.GlY - t.Whole.Top}

	t.MouseX = int((glDeltaFromHome.X + t.BarHori.ScrollDelta) / cGfx.CharWid)
	t.MouseY = int(-(glDeltaFromHome.Y + t.BarVert.ScrollDelta) / cGfx.CharHei)

	if t.MouseY < 0 {
		t.MouseY = 0
	}

	if t.MouseY >= len(t.TextBodies[0]) {
		t.MouseY = len(t.TextBodies[0]) - 1
	}
}

func (t *Terminal) GoToTopEdge() {
	cGfx.CurrY = t.Whole.Top - t.BarVert.ScrollDelta
}
func (t *Terminal) GoToLeftEdge() float32 {
	cGfx.CurrX = t.Whole.Left - t.BarHori.ScrollDelta
	return cGfx.CurrX
}
func (t *Terminal) GoToTopLeftCorner() {
	t.GoToTopEdge()
	t.GoToLeftEdge()
}

func (t *Terminal) Draw() {
	t.GoToTopLeftCorner()
	t.DrawBackground(t.Content)
	t.DrawText()
	//gfx.SetColor(cGfx..GrayDark)
	t.DrawScrollbarChrome(10, 11, t.Whole.Right-ui.ScrollBarThickness, t.Whole.Top)                          // vertical bar background
	t.DrawScrollbarChrome(13, 12, t.Whole.Left, t.Whole.Bottom+ui.ScrollBarThickness)                        // horizontal bar background
	t.DrawScrollbarChrome(12, 11, t.Whole.Right-ui.ScrollBarThickness, t.Whole.Bottom+ui.ScrollBarThickness) // corner elbow piece
	//gfx.SetColor(cGfx..Gray)
	t.BarHori.SetSize(t.Whole, t.TextBodies[0], cGfx.CharWid, cGfx.CharHei) // FIXME? (to consider multiple bodies & multiple trees)
	t.BarVert.SetSize(t.Whole, t.TextBodies[0], cGfx.CharWid, cGfx.CharHei)
	gfx.Update9SlicedRect(t.BarHori.Rect)
	gfx.Update9SlicedRect(t.BarVert.Rect)
	//gfx.SetColor(cGfx..White)
	t.DrawTree()
}

func (t *Terminal) DrawText() {
	cX := cGfx.CurrX // current drawing position
	cY := cGfx.CurrY
	cW := cGfx.CharWid
	cH := cGfx.CharHei
	b := t.BarHori.Rect.Rect.Top // bottom of text area

	// setup for colored text
	ncId := 0              // next color
	var nc *cGfx.ColorSpot // ^
	if /* colors exist */ len(t.TextColors) > 0 {
		nc = t.TextColors[ncId]
	}

	// iterate over lines
	for y, line := range t.TextBodies[0] {
		lineVisible := cY <= t.Whole.Top+cH && cY >= b

		if lineVisible {
			r := &app.PicRectangle{0, 0, 0, cGfx.Pic_GradientBorder, &app.Rectangle{cY, cX + cW, cY - cH, cX}} // t, r, b, l

			// if line needs vertical adjustment
			if cY > t.Whole.Top {
				r.Rect.Top = t.Whole.Top
			}
			if cY-cH < b {
				r.Rect.Bottom = b
			}

			// iterate over runes
			//cGfx.SetColor(cGfx.Gray)
			for x, c := range line {
				ncId, nc = t.changeColorIfCodeAt(x, y, ncId, nc)

				// drawing
				if /* char visible */ cX >= t.Whole.Left-cW && cX < t.BarVert.Rect.Rect.Left {
					app.ClampLeftAndRightOf(r.Rect, t.Whole.Left, t.BarVert.Rect.Rect.Left)
					gfx.DrawCharAtRect(c, r.Rect)

					if t.IsEditable { //&& Curs.Visible == true {
						if x == t.CursX && y == t.CursY {
							//gfx.SetColor(gfx.White)
							gfx.Update9SlicedRect(cGfx.Curs.GetAnimationModifiedRect(*r))
							//gfx.SetColor(gfx.PrevColor)
						}
					}
				}

				cX += cW
				r.Rect.Left = cX
				r.Rect.Right = cX + cW
			}

			// draw cursor at the end of line if needed
			if cX < t.BarVert.Rect.Rect.Left && y == t.CursY && t.CursX == len(line) {
				if t.IsEditable { //&& Curs.Visible == true {
					//gfx.SetColor(gfx.White)
					app.ClampLeftAndRightOf(r.Rect, t.Whole.Left, t.BarVert.Rect.Rect.Left)
					gfx.Update9SlicedRect(cGfx.Curs.GetAnimationModifiedRect(*r))
				}
			}

			cX = t.GoToLeftEdge()
		} else { // line not visible
			for x := range line {
				ncId, nc = t.changeColorIfCodeAt(x, y, ncId, nc)
			}
		}

		cY -= cH // go down a line height
	}
}

func (t *Terminal) changeColorIfCodeAt(x, y, ncId int, nc *cGfx.ColorSpot) (int, *cGfx.ColorSpot) {
	if /* colors exist */ len(t.TextColors) > 0 {
		if x == nc.Pos.X &&
			y == nc.Pos.Y {
			//gfx.SetColor(nc.Color)
			//fmt.Println("-------- nc-------, then 3rd():", nc, t.TextColors[2])
			ncId++

			if ncId < len(t.TextColors) {
				nc = t.TextColors[ncId]
			}
		}
	}

	return ncId, nc
}

// ATM the only different between the 2 funcs below is the top left corner (involving 3 vertices)
func (t *Terminal) DrawScrollbarChrome(atlasCellX, atlasCellY, l, top float32) { // l = left
	span := cGfx.UvSpan
	u := float32(atlasCellX) * span
	v := float32(atlasCellY) * span

	gl.Normal3f(0, 0, 1)

	// bottom left   0, 1
	gl.TexCoord2f(u, v+span)
	gl.Vertex3f(l, t.Whole.Bottom, 0)

	// bottom right   1, 1
	gl.TexCoord2f(u+span, v+span)
	gl.Vertex3f(t.Whole.Right, t.Whole.Bottom, 0)

	// top right   1, 0
	gl.TexCoord2f(u+span, v)
	gl.Vertex3f(t.Whole.Right, top, 0)

	// top left   0, 0
	gl.TexCoord2f(u, v)
	gl.Vertex3f(l, top, 0)
}

func (t *Terminal) DrawBackground(r *app.PicRectangle) {
	//gfx.SetColor(gfx.GrayDark)
	gfx.Update9SlicedRect(r)
}

func (t *Terminal) ScrollIfMouseOver(mousePixelDeltaX, mousePixelDeltaY float32) {
	if t.ContainsMouseCursor() {
		// position increments in gl space
		xInc := mousePixelDeltaX * cGfx.PixelSize.X
		yInc := mousePixelDeltaY * cGfx.PixelSize.Y
		t.BarHori.Scroll(xInc)
		t.BarVert.Scroll(yInc)
	}
}

func (t *Terminal) ContainsMouseCursor() bool {
	return mouse.CursorIsInside(t.Whole)
}

func (t *Terminal) RemoveCharacter(fromUnderCursor bool) {
	txt := t.TextBodies[0]

	if fromUnderCursor {
		if len(txt[t.CursY]) > t.CursX {
			txt[t.CursY] = txt[t.CursY][:t.CursX] + txt[t.CursY][t.CursX+1:len(txt[t.CursY])]
		}
	} else {
		if t.CursX > 0 {
			txt[t.CursY] = txt[t.CursY][:t.CursX-1] + txt[t.CursY][t.CursX:len(txt[t.CursY])]
			t.CursX--
		}
	}
}

func (t *Terminal) DrawTree() {
	if len(t.Trees) > 0 {
		// setup main rect
		span := float32(1.3)
		x := -span / 2
		y := t.Whole.Top - 0.1
		r := &app.Rectangle{y, x + span, y - span, x}

		t.drawNodeAndDescendants(r, 0)
	}
}

func (t *Terminal) drawNodeAndDescendants(r *app.Rectangle, nodeId int) {
	/*
		//fmt.Println("drawNode(r *app.Rectangle)")
		nameBar := &app.Rectangle{r.Top, r.Right, r.Top - 0.2*r.Height(), r.Left}
		cGfx.Update9SlicedRect(Pic_GradientBorder, r)
		SetColor(Blue)
		cGfx.Update9SlicedRect(Pic_GradientBorder, nameBar)
		DrawTextInRect(t.Trees[0].Nodes[nodeId].Text, nameBar)
		SetColor(White)

		cX := r.CenterX()
		rSp := r.Width() // rect span (height & width are the same)
		top := r.Bottom - rSp*0.5
		b := r.Bottom - rSp*1.5 // bottom

		node := t.Trees[0].Nodes[nodeId] // FIXME? .....
		// find t.Trees[0].Nodes[i].....
		// ......(if we ever use multiple trees per panel)
		// ......(also update DrawTree to use range)

		if node.ChildIdL != math.MaxInt32 {
			// (left child exists)
			x := cX - rSp*1.5
			t.drawArrowAndChild(r, &app.Rectangle{top, x + rSp, b, x}, node.ChildIdL)
		}

		if node.ChildIdR != math.MaxInt32 {
			// (right child exists)
			x := cX + rSp*0.5
			t.drawArrowAndChild(r, &app.Rectangle{top, x + rSp, b, x}, node.ChildIdR)
		}
	*/
}

func (t *Terminal) drawArrowAndChild(parent, child *app.Rectangle, childId int) {
	/*
		latExt := child.Width() * 0.15 // lateral extent of arrow's triangle top
		DrawTriangle(9, 1,
			app.Vec2F{parent.CenterX() - latExt, parent.Bottom},
			app.Vec2F{parent.CenterX() + latExt, parent.Bottom},
			app.Vec2F{child.CenterX(), child.Top})
		t.drawNodeAndDescendants(child, childId)
	*/
}

func (t *Terminal) SetupDemoProgram() {
	txt := []string{}

	txt = append(txt, "// ------- variable declarations ------- -------")
	//txt = append(txt, "var myVar int32")
	txt = append(txt, "var a int32 = 42 // end-of-line comment")
	txt = append(txt, "var b int32 = 58")
	txt = append(txt, "")
	txt = append(txt, "// ------- builtin function calls ------- ------- ------- ------- ------- ------- ------- end")
	txt = append(txt, "//    sub32(7, 9)")
	//txt = append(txt, "sub32(4,8)")
	//txt = append(txt, "mult32(7, 7)")
	//txt = append(txt, "mult32(3,5)")
	//txt = append(txt, "div32(8,2)")
	//txt = append(txt, "div32(15,  3)")
	//txt = append(txt, "add32(2,3)")
	//txt = append(txt, "add32(a, b)")
	txt = append(txt, "")
	txt = append(txt, "// ------- user function calls -------")
	txt = append(txt, "myFunc(a, b)")
	txt = append(txt, "")
	txt = append(txt, "// ------- function declarations -------")
	txt = append(txt, "func myFunc(a int32, b int32){")
	txt = append(txt, "")
	txt = append(txt, "        div32(6, 2)")
	txt = append(txt, "        innerFunc(a,b)")
	txt = append(txt, "}")
	txt = append(txt, "")
	txt = append(txt, "func innerFunc (a, b int32) {")
	txt = append(txt, "        var locA int32 = 71")
	txt = append(txt, "        var locB int32 = 29")
	txt = append(txt, "        sub32(locA, locB)")
	txt = append(txt, "}")

	/*
		for i := 0; i < 22; i++ {
			txt = append(txt, fmt.Sprintf("%d: put lots of text on screen", i))
		}
	*/

	t.TextBodies[0] = txt
}
