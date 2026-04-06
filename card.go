package main

import (
	"fmt"
	"image/color"
	"math/rand/v2"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
)

func CreateCards(cards *[]card, diamondCount *int, diamondQuota *int) {
	// modifier 1 à genre 7
	if *diamondCount >= *diamondQuota {
		*diamondQuota += 2
		*diamondCount = 0 // ← ici, une seule fois

		type card2 struct {
			name        string
			description string
		}
		cards2 := []card2{
			{"pierce", "the bullets pierce the enemies"},
			{"machine gun", "the cadence is accelerated"},
			{"vampire", "when you kill a zombie you regenerate"},
			{"treasure hunter", "attracts diamonds from a greater distance"},
			{"player speed", "you go faster"},
			{"sniper", "you can shoot from further away"},
		}

		usedIndices := map[int]bool{}

		for i := 0; i < 3; i++ {
			nameInt := rand.IntN(len(cards2))
			for usedIndices[nameInt] {
				nameInt = rand.IntN(len(cards2))
			}
			usedIndices[nameInt] = true

			if nameInt == 1 {
				if rand.IntN(2) == 1 {
					i--
					continue
				}
			}

			def := cards2[nameInt]
			*cards = append(*cards, card{
				x:           float64(i*775 + 160),
				y:           30,
				w:           775,
				h:           screenHeight - 60,
				description: def.description,
				name:        def.name,
				clr:         color.RGBA{0, 200, 0, 255},
			})
		}
	}
}

func DetectClickOnCard(cards *[]card, upgrades map[string]int, cadence *float64, playerSpeed *float64, shootRange *float64, clicPrecedent *bool, pickupRadius *float64) map[string]int {

	xC, yC := ebiten.CursorPosition()
	x, y := float64(xC), float64(yC)

	clicActuel := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	for _, c := range *cards {
		if clicActuel && !*clicPrecedent {
			if gameutil.Within(x, y, c.x, c.y, c.w, c.h) {
				// debug
				fmt.Println(lineSeparation)
				fmt.Print("upgrade: ")
				fmt.Println(c.name)

				switch c.name {
				case "machine gun":
					*cadence -= 5
				case "pierce":
					upgrades["pierce"] += 6
				case "vampire":
					upgrades["vampire"] += 3
				case "treasure hunter":
					*pickupRadius += 75 // <-- MODIFIÉ : Augmente le rayon de ramassage du joueur
				case "player speed":
					*playerSpeed += 3
				case "sniper":
					*shootRange += 200
				}

				*cards = []card{}
			}
		}
	}

	*clicPrecedent = clicActuel

	return upgrades
}
