package mustacher

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

// DrawMustaches generates a new image with graphical mustaches
// drawn at every DatabaseMatch specified.
func DrawMustaches(img image.Image, info []*DatabaseMatch) image.Image {
	newImage := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	ctx := draw2dimg.NewGraphicContext(newImage)
	ctx.DrawImage(img)
	for _, mustacheInfo := range info {
		ctx.Save()
		ctx.Translate(mustacheInfo.Center.X, mustacheInfo.Center.Y)
		ctx.Rotate(mustacheInfo.Rotation * math.Pi / 180)
		drawMustache(ctx, mustacheInfo.MouthWidth*2)
		ctx.Restore()
	}
	return newImage
}

func drawMustache(ctx draw2d.GraphicContext, width float64) {
	// We do not use ctx.Scale() for scaling because scaling up Bezier curves
	// makes their vertices visible.
	scale := width / 100

	ctx.Save()
	ctx.Translate(-50*scale, -15*scale)

	ctx.SetFillColor(color.Black)
	ctx.BeginPath()
	ctx.MoveTo(14*scale, 4*scale)

	// "Bottom" section of mustache.
	ctx.CubicCurveTo(12*scale, 1*scale, 0*scale, 1*scale, 0*scale, 13*scale)
	ctx.CubicCurveTo(0*scale, 33*scale, 30*scale, 33*scale, 50*scale, 17*scale)
	ctx.CubicCurveTo(70*scale, 33*scale, 100*scale, 33*scale, 100*scale, 13*scale)
	ctx.CubicCurveTo(100*scale, 1*scale, 88*scale, 1*scale, 86*scale, 4*scale)

	// "Top" section of mustache.
	ctx.CubicCurveTo(96*scale, 4*scale, 96*scale, 19*scale, 88*scale, 19*scale)
	ctx.CubicCurveTo(83*scale, 19*scale, 68*scale, 1*scale, 60*scale, 1*scale)
	ctx.CubicCurveTo(57*scale, 1*scale, 53*scale, 2*scale, 50*scale, 6*scale)
	ctx.CubicCurveTo(47*scale, 2*scale, 43*scale, 1*scale, 40*scale, 1*scale)
	ctx.CubicCurveTo(32*scale, 1*scale, 17*scale, 19*scale, 12*scale, 19*scale)
	ctx.CubicCurveTo(4*scale, 19*scale, 4*scale, 4*scale, 14*scale, 4*scale)

	ctx.Close()
	ctx.Fill()
	ctx.Restore()
}
