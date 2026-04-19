package card

import (
	"fmt"
	"image/color"
	"math/rand/v2"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Card struct {
	x, y, w, h  float64
	description string
	name        string
	clr         color.RGBA
}

func Create(cards *[]Card, diamondCount *int, diamondQuota *int, screenHeight float64, bossCooldown int, special bool) {

	// s'en aller immédiatement si c'est le boss
	if bossCooldown <= 0 {
		return
	}

	// modifier 1 à genre 7
	if *diamondCount >= *diamondQuota || special {
		if !special {
			*diamondQuota += 2
			*diamondCount = 0 // ← ici, une seule fois
		}

		type card2 struct {
			name        string
			description string
		}

		cards2 := []card2{}

		if !special {
			cards2 = []card2{
				{"pierce", "the bullets pierce the enemies"},
				{"machine gun", "the cadence is accelerated"},
				{"vampire", "when you kill a zombie you regenerate"},
				{"treasure hunter", "attracts diamonds from a greater distance"},
				{"player speed", "you go faster"},
				{"sniper", "you can shoot from further away"},
				{"fridge", "It freezes enemies around the bullet if it hits its target."},
			}
		} else {
			cards2 = []card2{
				{"guided missile", "your bullet always hits its target"},
				{"care kit", "you regenerate half of your maximum HP"},
				{"tank", "you got more max HP"},
			}
		}

		usedIndices := map[int]bool{}
		position := 0

		for position < 3 {
			var nameInt int
			retry := 0
			for {
				nameInt = rand.IntN(len(cards2))
				if !usedIndices[nameInt] {
					break
				}
				retry++
				if retry > 100 { // Prevent infinite loop
					break
				}
			}
			usedIndices[nameInt] = true
			def := cards2[nameInt]
			*cards = append(*cards, Card{
				x:           float64(position*775 + 160),
				y:           30,
				w:           775,
				h:           screenHeight - 60,
				description: def.description,
				name:        def.name,
				clr:         color.RGBA{0, 200, 0, 255},
			})
			position++
		}
	}
}

func DetectClick(cards *[]Card, upgrades map[string]int, cadence *float64, playerSpeed *float64, shootRange *float64, clicPrecedent *bool, pickupRadius *float64, maxHealth *int, health *int) map[string]int {

	xC, yC := ebiten.CursorPosition()
	x, y := float64(xC), float64(yC)

	clicActuel := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	for _, c := range *cards {
		if clicActuel && !*clicPrecedent {
			if gameutil.Within(x, y, c.x, c.y, c.w, c.h) {
				// debug
				fmt.Println("=====================")
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
					*pickupRadius += 75
				case "player speed":
					*playerSpeed += 3
				case "sniper":
					*shootRange += 200
				case "fridge":
					upgrades["fridge"] += 8
				case "guided missile":
					upgrades["guided missile"] += 40
				case "care kit":
					*health += *maxHealth / 2
				case "tank":
					*maxHealth += 30
				}

				*cards = []Card{}
			}
		}
	}

	*clicPrecedent = clicActuel

	return upgrades
}

func (c Card) Draw(screen *ebiten.Image, mplusSource *text.GoTextFaceSource) {

	vector.StrokeRect(screen, float32(c.x), float32(c.y), float32(c.w), float32(c.h), 30, c.clr, true)

	gameutil.DrawText(c.name, 100, int(c.x+c.w), c.x+30, c.y+100, 0, screen, color.RGBA{255, 255, 255, 255}, mplusSource)

	gameutil.DrawText(c.description, 40, int(c.x+c.w), c.x+30, c.y+400, 0, screen, color.RGBA{255, 255, 255, 255}, mplusSource)
}
