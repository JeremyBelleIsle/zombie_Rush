package diamond

import (
	"math"
	"slices"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
)

type Diamond struct {
	x, y, r                float64
	DetectedInPickupRadius bool
	img                    *ebiten.Image
}

func DragToPlayer(diamonds []Diamond, PWX, PWY float64, playerDiamonds *int, bossCooldown int) ([]Diamond, bool) {

	if bossCooldown <= 0 {
		return diamonds, false
	}

	for i := len(diamonds) - 1; i >= 0; i-- {
		d := &diamonds[i]

		if d.DetectedInPickupRadius {
			f32x, f32y := gameutil.DirigePointToPoint(30, float32(d.x), float32(d.y), float32(PWX), float32(PWY))

			d.x, d.y = float64(f32x), float64(f32y)

			if math.Abs(d.x-PWX) < 31 && math.Abs(d.y-PWY) < 31 {

				diamonds = slices.Delete(diamonds, i, i+1)

				*playerDiamonds++

				return diamonds, true
			}
		}
	}

	return diamonds, false
}

func PickupRadius(px, py, pr float64, mapX, mapY float64, diamonds []Diamond, bossCooldown int) (bool, int) {

	if bossCooldown <= 0 {
		return false, 0
	}

	for i, d := range diamonds {
		if gameutil.CircleCollision(px-mapX, py-mapY, pr, d.x, d.y, d.r) {
			return true, i
		}
	}

	return false, 0
}

func Spawn(diamonds *[]Diamond, x, y, diamondStartedRadius float64, diamondImg *ebiten.Image, bossCooldown int) {

	if bossCooldown <= 0 {
		return
	}

	*diamonds = append(*diamonds, Diamond{
		x:   x,
		y:   y,
		r:   diamondStartedRadius,
		img: diamondImg,
	})

}

func (d Diamond) Draw(screen *ebiten.Image, mapX, mapY float64) {
	op := &ebiten.DrawImageOptions{}

	op.GeoM.Scale(0.3, 0.3)

	op.GeoM.Translate((d.x+mapX)-70, (d.y+mapY)-70)

	if d.img == nil {
		panic("diamond image == nil")
	}

	screen.DrawImage(d.img, op)
}
