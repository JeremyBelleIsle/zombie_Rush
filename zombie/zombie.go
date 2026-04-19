package zombie

import (
	"image/color"
	"math"
	"math/rand/v2"
	"slices"
	"zombie_rush/ice"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Zombie struct {
	img        *ebiten.Image
	X, Y, R, s float64
	speed      float64
	angle      float64
	Health     int
	freeze     int

	Boss bool
}

func inTheIce(iceX, iceY, iceR, zombieX, zombieY, zombieR float64) bool {
	return gameutil.CircleCollision(iceX, iceY, iceR, zombieX, zombieY, zombieR)
}

func Movement(zombies []Zombie, px, py float64, ice ice.Ice) []Zombie {

	for i := range zombies {
		z := &zombies[i]

		if inTheIce(ice.X, ice.Y, ice.R, z.X, z.Y, z.R) {
			continue
		}

		x, y := gameutil.DirigePointToPoint(float32(z.speed), float32(z.X), float32(z.Y), float32(px), float32(py))

		z.angle = math.Atan2(py-z.Y, px-z.X)

		z.X, z.Y = float64(x), float64(y)
	}

	return zombies
}

func Spawn(zombies []Zombie, addZombieCooldown *float64, setSpawnZombieCadence float64, zombieImg *ebiten.Image, screenWidht, screenHeight int, mapX, mapY float64) []Zombie {

	if *addZombieCooldown > 0 {
		*addZombieCooldown--
	} else {
		*addZombieCooldown = setSpawnZombieCadence

		zombies = append(zombies, Zombie{
			X:      float64(rand.IntN(screenWidht+1000)+-500) - mapX,
			Y:      float64(rand.IntN(screenHeight+1000)+-500) - mapY,
			R:      40,
			s:      .2,
			Health: 1,
			speed:  float64(rand.IntN(2)+5) + rand.Float64(),
			img:    zombieImg,
		})
	}

	return zombies
}

func Attack(px, py, pr float64, zombies *[]Zombie, lifes *int, impactSnd *audio.Player) {
	deleteElems := []int{}

	for i, z := range *zombies {
		if gameutil.CircleCollision(px, py, pr, z.X, z.Y, z.R) {

			if !z.Boss {
				deleteElems = append(deleteElems, i)
				*lifes -= 8

				impactSnd.Rewind()
				impactSnd.Play()
			} else {
				*lifes--
			}
		}
	}

	for i := len(deleteElems) - 1; i >= 0; i-- {
		*zombies = slices.Delete(*zombies, deleteElems[i], deleteElems[i]+1)
	}
}

func UpdateBossPhase(
	zombies *[]Zombie, // Pointeur vers la slice
	bossCooldown *int, // Pointeur vers l'entier
	playerAngle *float64, // Pointeur vers l'angle
	playerX, playerY float64,
	playerSpeed float64,
	bossImg *ebiten.Image,
	mapX, mapY *float64,
	alarmSnd *audio.Player,
) {
	// 1. Filtrage des zombies : on modifie la slice pointée
	var remaining []Zombie
	for _, z := range *zombies {
		if z.Boss {
			remaining = append(remaining, z)
		}
	}

	*zombies = remaining

	if *bossCooldown == 0 {

		alarmSnd.Rewind()
		alarmSnd.Play()

		*zombies = append(*zombies, Zombie{
			X:      playerX + 1000,
			Y:      playerY,
			R:      200,
			s:      2,
			img:    bossImg,
			speed:  playerSpeed - 3,
			Boss:   true,
			Health: 10,
		})

		*mapX = 0
		*mapY = 0
	}

	// 4. Mise à jour de l'angle du joueur
	if len(*zombies) > 0 {
		target := (*zombies)[0]
		*playerAngle = math.Atan2(-(target.Y - (playerY - *mapY)), -(target.X - (playerX - *mapX)))
	}

	// 5. Décrémentation du cooldown
	*bossCooldown--
}

func (z Zombie) Draw(screen *ebiten.Image, mapX, mapY float64) {
	op := &ebiten.DrawImageOptions{}

	w := z.img.Bounds().Dx()
	h := z.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Rotate(z.angle + math.Pi)

	op.GeoM.Scale(z.s, z.s)

	op.GeoM.Translate(z.X+mapX, z.Y+mapY)

	screen.DrawImage(z.img, op)

	if z.Boss {
		// jauge de vie du boss
		vector.StrokeRect(screen, (float32(z.X)-100)+float32(mapX), (float32(z.Y)-300)+float32(mapY), 260, 60, 10, color.RGBA{255, 255, 255, 255}, true)

		vector.DrawFilledRect(screen, ((float32(z.X)-100)+5)+float32(mapX), ((float32(z.Y)-300)+5)+float32(mapY), float32(z.Health)*(250/10), 54, color.RGBA{0, 255, 0, 255}, true)
	}
}
