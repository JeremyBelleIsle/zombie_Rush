package bullet

import (
	"image/color"
	"math"
	"math/rand/v2"
	"slices"
	"zombie_rush/zombie"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
)

type Bullet struct {
	img        *ebiten.Image
	x, y, w, h float64
	angle      float64
	vx, vy     float64
	clr        color.RGBA
}

func HitZombie(zombies []zombie.Zombie, bullets []Bullet, WX, WY float64) (bool, int, int) {

	for i, z := range zombies {

		for j, b := range bullets {

			if gameutil.CircleRectCollision(z.X, z.Y, z.R, b.x-WX, b.y-WY, b.w, b.h) {
				return true, i, j
			}

		}
	}

	return false, 0, 0

}

func Create(px, py, pWorldX, pWorldY float64, playerAngle *float64, shootRange float64, cadence float64, zombies []zombie.Zombie, bullets []Bullet, cooldown *int, bulletImg *ebiten.Image) []Bullet {
	// 1. Gestion du délai de tir
	if *cooldown > 0 {
		*cooldown--
		return bullets
	}

	// 2. Recherche du zombie le plus proche
	var closestZombie *zombie.Zombie
	minDist := float64(shootRange * shootRange) // Distance max au carré (ex: 200 pixels)

	for i := range zombies {
		z := &zombies[i]

		// On compare avec la position du joueur dans le monde
		dx := z.X - pWorldX
		dy := z.Y - pWorldY
		distSq := dx*dx + dy*dy

		if distSq < minDist {
			minDist = distSq
			closestZombie = z
			*playerAngle = math.Atan2(-dy, -dx)
		}
	}

	// 3. Création de la balle si on a une cible
	if closestZombie != nil {
		*cooldown = int(cadence) // On réinitialise le délai

		dx := closestZombie.X - pWorldX
		dy := closestZombie.Y - pWorldY
		dist := math.Sqrt(dx*dx + dy*dy)

		bulletSpeed := 15.0

		// On ajoute la balle aux coordonnées de l'écran (px, py) comme tu l'as précisé
		bullets = append(bullets, Bullet{
			x:     px,
			y:     py,
			w:     16,
			h:     16,
			angle: math.Atan2(dy, dx),
			img:   bulletImg,
			vx:    (dx / dist) * bulletSpeed,
			vy:    (dy / dist) * bulletSpeed,
			clr:   color.RGBA{0, 255, 0, 255},
		})
	}

	return bullets
}

func Move(bullets []Bullet, px, py float64) []Bullet {
	for i := len(bullets) - 1; i >= 0; i-- {
		bullets[i].x += bullets[i].vx
		bullets[i].y += bullets[i].vy

		// Si la balle s'éloigne à plus de 1500 pixels, on la supprime
		dx := bullets[i].x - px
		dy := bullets[i].y - py
		if math.Sqrt(dx*dx+dy*dy) > 1500 {
			bullets = slices.Delete(bullets, i, i+1)
		}
	}
	return bullets
}

func HitZombieReaction(zombieIndex int, bulletIndex int, zombies []zombie.Zombie, bullets []Bullet, upgrades map[string]int, lifes *int, bossCooldown *int) ([]zombie.Zombie, []Bullet) {

	zombies[zombieIndex].Health--

	if zombies[zombieIndex].Health <= 0 {

		if zombies[zombieIndex].Boss {
			*bossCooldown = 1800
		}

		zombies = slices.Delete(zombies, zombieIndex, zombieIndex+1)
	}

	if upgrades["vampire"] > 0 && rand.IntN(100) < upgrades["vampire"] {
		*lifes++
	}

	if upgrades["pierce"] == 0 || rand.IntN(100) > upgrades["pierce"] {
		bullets = slices.Delete(bullets, bulletIndex, bulletIndex+1)
	}

	return zombies, bullets
}

func (b Bullet) Draw(screen *ebiten.Image) {
	if b.img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}

	w := b.img.Bounds().Dx()
	h := b.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Rotate(b.angle)

	op.GeoM.Scale(.14, .14)

	op.GeoM.Translate(b.x, b.y)

	screen.DrawImage(b.img, op)
}
