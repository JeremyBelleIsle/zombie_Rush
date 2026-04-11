package ice

import "github.com/hajimehoshi/ebiten/v2"

type Ice struct {
	X, Y, R, s float64
	img        *ebiten.Image
	Life       int
}

func Spawn(x, y float64, iceCircleImg *ebiten.Image, ice *Ice) {
	*ice = Ice{X: x, Y: y, R: 150, s: 1, img: iceCircleImg, Life: 50}
}

func (i Ice) Draw(screen *ebiten.Image, mapX, mapY float64) {
	if i.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}

	w := i.img.Bounds().Dx()
	h := i.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Scale(i.s, i.s)

	op.GeoM.Translate(i.X+mapX, i.Y+mapY)

	screen.DrawImage(i.img, op)
}
