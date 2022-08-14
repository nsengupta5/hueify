package generator

import (
	"github.com/fogleman/gg"
	"image/color"
)

func DrawPlayistCover(primaryR, primaryG, primaryB, secondaryR, secondaryG, secondaryB uint8) (err error) {
	dc := gg.NewContext(300, 300)

	grad := gg.NewRadialGradient(150, 150, 10, 150, 150, 250)
	grad.AddColorStop(1, color.RGBA{primaryR, primaryG, primaryB, 255})
	grad.AddColorStop(0, color.RGBA{secondaryR, secondaryG, secondaryB, 255})

	grad2 := gg.NewRadialGradient(150, 150, 1, 150, 150, 120)
	grad2.AddColorStop(1, color.RGBA{primaryR, primaryG, primaryB, 200})
	grad2.AddColorStop(0, color.RGBA{secondaryR, secondaryG, secondaryB, 200})

	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(dc.Width()), float64(dc.Height()))
	dc.Fill()

	dc.SetFillStyle(grad2)
	dc.DrawCircle(120, 120, 50)
	dc.Fill()

	dc.SetFillStyle(grad2)
	dc.DrawCircle(180, 180, 50)
	dc.Fill()

	err = dc.SavePNG("image.png")
	if err != nil {
		return err
	}

	return nil
}
