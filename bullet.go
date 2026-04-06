package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/JeremyBelleIsle/gameutil"
)

func BulletHitZombie(zombies []zombie, bullets []bullet, WX, WY float64) (bool, int, int) {

	for i, z := range zombies {

		for j, b := range bullets {

			if gameutil.CircleRectCollision(z.x, z.y, z.r, b.x-WX, b.y-WY, b.w, b.h) {
				return true, i, j
			}

		}
	}

	return false, 0, 0

}

func CreateBullet(px, py, pWorldX, pWorldY float64, playerAngle *float64, shootRange float64, cadence float64, zombies []zombie, bullets []bullet, cooldown *int) []bullet {
	// 1. Gestion du délai de tir
	if *cooldown > 0 {
		*cooldown--
		return bullets
	}

	// 2. Recherche du zombie le plus proche
	var closestZombie *zombie
	minDist := float64(shootRange * shootRange) // Distance max au carré (ex: 200 pixels)

	for i := range zombies {
		z := &zombies[i]

		// On compare avec la position du joueur dans le monde
		dx := z.x - pWorldX
		dy := z.y - pWorldY
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

		dx := closestZombie.x - pWorldX
		dy := closestZombie.y - pWorldY
		dist := math.Sqrt(dx*dx + dy*dy)

		bulletSpeed := 15.0

		// On ajoute la balle aux coordonnées de l'écran (px, py) comme tu l'as précisé
		bullets = append(bullets, bullet{
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

func MoveBullets(bullets []bullet, px, py float64) []bullet {
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

func BulletHitZombieReaction(zombieIndex int, bulletIndex int, zombies []zombie, bullets []bullet, upgrades map[string]int, lifes *int, bossCooldown *int) ([]zombie, []bullet) {

	zombies[zombieIndex].health--

	if zombies[zombieIndex].health <= 0 {

		if zombies[zombieIndex].boss {
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
