package main

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func PlayerIsInTheArena(pwx, pwy, pr float64, fenceX, fenceY, fenceRadius float64) bool {
	// 1. Calcul de la distance entre le centre du joueur et le centre de la clôture
	dx := pwx - fenceX
	dy := pwy - fenceY
	distance := math.Sqrt(dx*dx + dy*dy)

	// 2. Vérification : est-ce que le bord du joueur dépasse le bord de la clôture ?
	if distance+pr <= fenceRadius {
		return true // Le joueur est bien à l'intérieur
	}

	return false // Le joueur a touché ou dépassé la limite
}

func MovePlayer(mapX, mapY *float64, speed float64, bossCooldown int, px, py, pr float64) {

	futureMapX := *mapX
	futureMapY := *mapY

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		futureMapX += speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		futureMapX -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		futureMapY -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		futureMapY += speed
	}

	if bossCooldown <= 0 {

		w := fenceImg.Bounds().Dx()
		// (largeur / 2) * scale
		fenceRadius := (float64(w) / 2.0) * 7.0

		fenceRadius -= 1200

		if !PlayerIsInTheArena(px-futureMapX, py-futureMapY, pr, screenWidth/2, screenHeight/2, fenceRadius) {
			fmt.Println("player want to go outside of the arena")
			return
		}

	}

	*mapX = futureMapX
	*mapY = futureMapY

}
