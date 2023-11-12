## How to use custom shape

[Source Code](../../example/with-custom-shape/main.go)

first step, you must define your own shape to QRCode, which consists of two part:
* normal cell (of course, there are many types, separator, timing, alignment patter, data, format and version etc)
* finder cell (to help recognizer to locate the matrix's position)

<img src="../../assets/qrcode_structure.png" align="center" width="50%" />

```go
type IShape interface {
	// Draw to fill the IShape of qrcode.
	Draw(ctx *DrawContext)

	// DrawFinder to fill the finder pattern of QRCode, what's finder? google it for more information.
	DrawFinder(ctx *DrawContext)
}
```

> Notice:
> 
> if you must be careful to design finder's shape, otherwise qrcode could not be recognized.
> 


Now, if you're define your shape like this:

```go
func newShape(radiusPercent float64) qrcode.IShape {
	return &smallerCircle{smallerPercent: radiusPercent}
}

// smallerCircle use smaller circle to qrcode.  
type smallerCircle struct {
	smallerPercent float64
}

func (sc *smallerCircle) DrawFinder(ctx *qrcode.DrawContext) {
	// use normal radius to draw finder for that qrcode image can be recognized. 
	backup := sc.smallerPercent
	sc.smallerPercent = 1.0
	sc.Draw(ctx)
	sc.smallerPercent = backup
}

func (sc *smallerCircle) Draw(ctx *qrcode.DrawContext) {
	w, h := ctx.Edge()
	upperLeft := ctx.UpperLeft()
	color := ctx.Color()

	// choose a proper radius values
	radius := w / 2
	r2 := h / 2
	if r2 <= radius {
		radius = r2
	}

	// 80 percent smaller
	radius = int(float64(radius) * sc.smallerPercent)

	cx, cy := upperLeft.X+w/2, upperLeft.Y+h/2 // get center point
	ctx.DrawCircle(float64(cx), float64(cy), float64(radius))
	ctx.SetColor(color)
	ctx.Fill()

}
```

Finally, you can use your shape.

```go
func main() {
	shape := newShape(0.7)
	qrc, err := qrcode.New("with-custom-shape", qrcode.WithCustomShape(shape))
	if err != nil {
		panic(err)
	}

	err = qrc.Save("./smaller.png")
	if err != nil {
		panic(err)
	}
}
```