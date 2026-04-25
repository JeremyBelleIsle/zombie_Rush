package arrow

import (
	"math"
	"slices"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
)

type Arrow struct {
	img        *ebiten.Image
	x, y, w, h float64
	angle      float64
	vx, vy     float64
}

func Create(pWorldX, pWorldY float64, spawnX, spawnY float64, arrowImg *ebiten.Image, arrows *[]Arrow) {

	const bulletSpeed = 15.0

	dx := pWorldX - spawnX
	dy := pWorldY - spawnY
	dist := math.Sqrt(dx*dx + dy*dy)

	*arrows = append(*arrows, Arrow{
		x:     spawnX,
		y:     spawnY,
		w:     16,
		h:     16,
		angle: math.Atan2(dy, dx),
		img:   arrowImg,
		vx:    (dx / dist) * bulletSpeed,
		vy:    (dy / dist) * bulletSpeed,
	})
}

func Move(pWorldX, pWorldY float64, arrows []Arrow) []Arrow {
	for i := len(arrows) - 1; i >= 0; i-- {
		a := &arrows[i]

		a.x += a.vx
		a.y += a.vy

		// Si la flèche s'éloigne à plus de 1500 pixels, on la supprime
		dx := (pWorldX - a.x)
		dy := (pWorldY - a.y)
		if math.Sqrt(dx*dx+dy*dy) > 1000 {
			arrows = slices.Delete(arrows, i, i+1)
		}
	}

	return arrows
}

func ArrowsVsPlayerColl(arrows []Arrow, playerHealth *int, px, py, pr float64) []Arrow {

	for i := len(arrows) - 1; i >= 0; i-- {
		a := &arrows[i]

		if gameutil.CircleRectCollision(px, py, pr, a.x, a.y, a.w, a.h) {
			*playerHealth -= 5

			arrows = slices.Delete(arrows, i, i+1)
		}
	}

	return arrows
}

func (a Arrow) Draw(screen *ebiten.Image, mapX, mapY float64) {
	if a.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}

	w := a.img.Bounds().Dx()
	h := a.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Rotate(a.angle)

	op.GeoM.Scale(.14, .14)

	op.GeoM.Translate(a.x+mapX, a.y+mapY)

	screen.DrawImage(a.img, op)
}
