package main

import (
	"math"
	"math/rand/v2"
	"slices"

	"github.com/JeremyBelleIsle/gameutil"
)

func ZombiesMovement(zombies []zombie, px, py float64) []zombie {
	for i := range zombies {
		z := &zombies[i]

		x, y := gameutil.DirigePointToPoint(float32(z.speed), float32(z.x), float32(z.y), float32(px), float32(py))

		z.angle = math.Atan2(py-z.y, px-z.x)

		z.x, z.y = float64(x), float64(y)
	}

	return zombies
}

func ZombieSpawn(zombies []zombie, addZombieCooldown *float64) []zombie {

	if *addZombieCooldown > 0 {
		*addZombieCooldown--
	} else {
		*addZombieCooldown = 60

		zombies = append(zombies, zombie{
			x:      float64(rand.IntN(screenWidth+1000) + -500),
			y:      float64(rand.IntN(screenHeight+1000) + -500),
			r:      40,
			s:      .2,
			health: 1,
			speed:  float64(rand.IntN(2)+5) + rand.Float64(),
			img:    zombieImg,
		})
	}

	return zombies
}

func ZombieAttack(px, py, pr float64, zombies *[]zombie, lifes *int) {
	deleteElems := []int{}

	for i, z := range *zombies {
		if gameutil.CircleCollision(px, py, pr, z.x, z.y, z.r) {

			if !z.boss {
				deleteElems = append(deleteElems, i)
				*lifes -= 8
			} else {
				*lifes--
			}
		}
	}

	for i := len(deleteElems) - 1; i >= 0; i-- {
		*zombies = slices.Delete(*zombies, deleteElems[i], deleteElems[i]+1)
	}
}
