package main

import (
	"math"
	"slices"

	"github.com/JeremyBelleIsle/gameutil"
)

func DragDiamondsToPlayer(diamonds []diamond, PWX, PWY float64, playerDiamonds *int) []diamond {
	for i := len(diamonds) - 1; i >= 0; i-- {
		d := &diamonds[i]

		if d.detectedInPickupRadius {
			f32x, f32y := gameutil.DirigePointToPoint(30, float32(d.x), float32(d.y), float32(PWX), float32(PWY))

			d.x, d.y = float64(f32x), float64(f32y)

			// Maintenant, la suppression ne fera plus planter la boucle

			if math.Abs(d.x-PWX) < 31 && math.Abs(d.y-PWY) < 31 {

				diamonds = slices.Delete(diamonds, i, i+1)

				*playerDiamonds++

			}
		}
	}

	return diamonds
}

func DiamondXpickupRadius(px, py, pr float64, mapX, mapY float64, diamonds []diamond) (bool, int) {

	for i, d := range diamonds {
		if gameutil.CircleCollision(px-mapX, py-mapY, pr, d.x, d.y, d.r) {
			return true, i
		}
	}

	return false, 0
}

func SpawnDiamond(diamonds *[]diamond, x, y, diamondStartedRadius float64) {

	*diamonds = append(*diamonds, diamond{
		x:   x,
		y:   y,
		r:   diamondStartedRadius,
		img: diamondImg,
	})

}
