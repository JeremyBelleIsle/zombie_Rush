package player

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	img           *ebiten.Image
	Angle         float64
	X, Y, R       float64
	PickupRadius  float64
	Speed         float64
	ShootRange    float64
	ShootCooldown int
	Cadence       float64
	Lifes         int
	Diamond       int
	DiamondQuota  int
	clr           color.RGBA
}

func IsInTheArena(pwx, pwy, pr float64, fenceX, fenceY, fenceRadius float64) bool {
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

func (p *Player) Initialization(playerImg *ebiten.Image, screenWidth, screenHeight float64) {
	p.img = playerImg
	p.X = screenWidth / 2
	p.Y = screenHeight / 2
	p.R = 70
	p.PickupRadius = 80
	p.Cadence = 60
	p.ShootRange = 700
	p.ShootCooldown = 60
	p.DiamondQuota = 7
	p.Speed = 10
	p.Lifes = 100
	p.clr = color.RGBA{255, 255, 0, 255}
}

func Move(px, py, pr float64, mapX, mapY *float64, speed float64, bossCooldown int, fenceImg *ebiten.Image, screenWidth, screenHeight float64) {

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

		if !IsInTheArena(px-futureMapX, py-futureMapY, pr, screenWidth/2, screenHeight/2, fenceRadius) {
			fmt.Println("player want to go outside of the arena")
			return
		}

	}

	*mapX = futureMapX
	*mapY = futureMapY

}

func (p Player) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	w := p.img.Bounds().Dx()
	h := p.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Rotate(p.Angle)

	op.GeoM.Scale(.5, .5)

	op.GeoM.Translate(p.X, p.Y)

	screen.DrawImage(p.img, op)
}
