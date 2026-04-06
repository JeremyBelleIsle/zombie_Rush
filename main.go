package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand/v2"
	"os"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth    = 2560
	screenHeight   = 1600
	lineSeparation = "====================="

	StatePlaying  = 0
	StateGameOver = 1
)

type player struct {
	img           *ebiten.Image
	angle         float64
	x, y, r       float64
	pickupRadius  float64 // <-- NOUVEAU : Le rayon de ramassage
	speed         float64
	shootRange    float64
	shootCooldown int
	cadence       float64
	lifes         int
	diamond       int
	diamondQuota  int
	clr           color.RGBA
}

type card struct {
	x, y, w, h  float64
	description string
	name        string
	clr         color.RGBA
}

type miniatureCard struct {
	x, y, s float64
	img     *ebiten.Image
}

type diamond struct {
	x, y, r                float64
	detectedInPickupRadius bool
	img                    *ebiten.Image
}

type tree struct {
	Img     *ebiten.Image
	x, y, s float64
}

type bullet struct {
	img        *ebiten.Image
	x, y, w, h float64
	angle      float64
	vx, vy     float64
	clr        color.RGBA
}

type zombie struct {
	img        *ebiten.Image
	x, y, r, s float64
	speed      float64
	angle      float64
	health     int

	boss bool
}

type Game struct {
	bossCooldown        int
	state               int
	addZombieCooldown   float64
	zombieSpawnCooldown float64
	player              player
	cards               []card
	diamonds            []diamond
	bullets             []bullet
	zombies             []zombie
	trees               []tree
	miniatureCard       miniatureCard
	upgrades            map[string]int
	clicPrecedent       bool
	mapX, mapY          float64
}

var (
	diamondImg *ebiten.Image
	treeImg    *ebiten.Image
	cardImg    *ebiten.Image
	zombieImg  *ebiten.Image
	playerImg  *ebiten.Image
	bulletImg  *ebiten.Image
	fenceImg   *ebiten.Image
	bossImg    *ebiten.Image
)

var mplusSource *text.GoTextFaceSource

//go:embed RobotoMono-VariableFont_wght.ttf
var roboto []byte

var faceSource *text.GoTextFaceSource

func loadImage(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}

func antiCheatLimit(cadence *float64, speed *float64) {
	if *cadence > 60 {
		*cadence = 60
	}

	if *cadence < 20 {
		*cadence = 20
	}

	if *speed > 20 {
		*speed = 20
	}
}

func (g *Game) reset() {
	g.state = StatePlaying
	g.addZombieCooldown = 60
	g.player.lifes = 100
	g.player.pickupRadius = 100 // <-- NOUVEAU : Initialisation
	g.player.diamond = 0
	g.player.diamondQuota = 7
	g.player.cadence = 60
	g.player.shootCooldown = 60
	g.player.speed = 10
	g.bossCooldown = 1800
	g.bullets = []bullet{}
	g.diamonds = []diamond{}
	g.cards = []card{}
	g.zombies = []zombie{}
	g.upgrades = map[string]int{
		"pierce":  0,
		"vampire": 0,
	}
	g.mapX = 0
	g.mapY = 0
}

func (g *Game) Update() error {
	if g.state == StatePlaying {

		if g.player.lifes <= 0 {
			g.state = StateGameOver
		}

		if g.player.lifes > 100 {
			g.player.lifes = 100
		}

		if len(g.cards) == 0 {
			g.bossCooldown--

			MovePlayer(&g.mapX, &g.mapY, g.player.speed, g.bossCooldown, g.player.x, g.player.y, g.player.r)

			// Position du joueur dans le "monde" (en tenant compte de la caméra)
			playerWorldX := g.player.x - g.mapX
			playerWorldY := g.player.y - g.mapY

			g.zombies = ZombiesMovement(g.zombies, playerWorldX, playerWorldY)

			g.zombies = ZombieSpawn(g.zombies, &g.addZombieCooldown)

			g.bullets = CreateBullet(g.player.x, g.player.y, playerWorldX, playerWorldY, &g.player.angle, g.player.shootRange, g.player.cadence, g.zombies, g.bullets, &g.player.shootCooldown)
			MoveBullets(g.bullets, g.player.x, g.player.y)

			bulletHitZombie, zi, bi := BulletHitZombie(g.zombies, g.bullets, g.mapX, g.mapY)

			if bulletHitZombie {
				SpawnDiamond(&g.diamonds, g.zombies[zi].x, g.zombies[zi].y, 56) // <-- MODIFIÉ : Rayon du diamant fixe à 56

				g.zombies, g.bullets = BulletHitZombieReaction(zi, bi, g.zombies, g.bullets, g.upgrades, &g.player.lifes, &g.bossCooldown)
			}

			// <-- MODIFIÉ : Utilise g.player.pickupRadius au lieu de g.player.r pour ramasser
			diamondPickup, di := DiamondXpickupRadius(g.player.x, g.player.y, g.player.pickupRadius, g.mapX, g.mapY, g.diamonds)

			if diamondPickup {
				fmt.Println(lineSeparation)
				fmt.Println("diamond/player collision")

				g.diamonds[di].detectedInPickupRadius = true
			}

			g.diamonds = DragDiamondsToPlayer(g.diamonds, playerWorldX, playerWorldY, &g.player.diamond)

			CreateCards(&g.cards, &g.player.diamond, &g.player.diamondQuota)

			ZombieAttack(playerWorldX, playerWorldY, g.player.r, &g.zombies, &g.player.lifes)

			antiCheatLimit(&g.player.cadence, &g.player.speed)

			g.zombieSpawnCooldown--

			if g.bossCooldown <= 0 {

				var remainingZombies []zombie
				for _, z := range g.zombies {
					if z.boss {
						remainingZombies = append(remainingZombies, z)
					}
				}

				g.zombies = remainingZombies
				g.trees = []tree{}

				if g.bossCooldown == 0 {

					g.zombies = append(g.zombies, zombie{
						x:      playerWorldX + 1000,
						y:      playerWorldY,
						r:      200,
						img:    bossImg,
						speed:  g.player.speed - 1,
						angle:  0,
						health: 10,
						s:      2,
						boss:   true,
					})

					g.mapX = 0
					g.mapY = 0
				}

				g.player.angle = math.Atan2(-(g.zombies[0].y - playerWorldY), -(g.zombies[0].x - playerWorldX))

				g.bossCooldown--

			}
		} else {
			if len(g.cards) > 3 {
				panic("have more than 3 cards")
			}

			g.upgrades = DetectClickOnCard(&g.cards, g.upgrades, &g.player.cadence, &g.player.speed, &g.player.shootRange, &g.clicPrecedent, &g.player.pickupRadius)
		}
	} else {
		touches := inpututil.AppendJustPressedKeys(nil)

		if len(touches) > 0 {
			g.reset()
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	if g.bossCooldown <= 0 {

		// Récupérer les dimensions originales de l'image
		w, h := fenceImg.Size()
		scale := 7.0

		op := &ebiten.DrawImageOptions{}

		// 1. Centrer l'image sur le point (0,0)
		op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

		// 2. Appliquer le redimensionnement (Scale)
		op.GeoM.Scale(scale, scale)

		// 3. Déplacer au centre de l'écran
		op.GeoM.Translate(screenWidth/2+g.mapX, screenHeight/2+g.mapY)

		screen.DrawImage(fenceImg, op)

	} else {

		// debug
		vector.StrokeCircle(screen, float32(g.player.x), float32(g.player.y), float32(g.player.pickupRadius), 2, color.RGBA{0, 0, 120, 120}, true)

		for _, d := range g.diamonds {
			op := &ebiten.DrawImageOptions{}

			op.GeoM.Scale(0.3, 0.3)

			op.GeoM.Translate((d.x+g.mapX)-70, (d.y+g.mapY)-70)

			if d.img == nil {
				panic(g.diamonds)
			}

			screen.DrawImage(d.img, op)
		}

		for _, t := range g.trees {
			op := &ebiten.DrawImageOptions{}

			op.GeoM.Scale(t.s, t.s)

			op.GeoM.Translate(t.x+g.mapX, t.y+g.mapY)

			screen.DrawImage(t.Img, op)
		}

		vector.StrokeRect(screen, 700, 50, 500, 60, 10, color.RGBA{255, 255, 255, 255}, true)

		vector.DrawFilledRect(screen, 700, 50, float32(g.player.diamond*(500/g.player.diamondQuota)), 54, color.RGBA{38, 115, 211, 255}, true)
		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(g.miniatureCard.s, g.miniatureCard.s)

		op.GeoM.Translate(g.miniatureCard.x, g.miniatureCard.y)

		screen.DrawImage(g.miniatureCard.img, op)

		if len(g.cards) > 0 {
			vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 50, 50, 240}, false)

			for _, c := range g.cards {
				vector.StrokeRect(screen, float32(c.x), float32(c.y), float32(c.w), float32(c.h), 30, c.clr, true)

				gameutil.DrawText(c.name, 100, int(c.x+c.w), c.x+30, c.y+100, 0, screen, color.RGBA{255, 255, 255, 255}, mplusSource)

				gameutil.DrawText(c.description, 40, int(c.x+c.w), c.x+30, c.y+400, 0, screen, color.RGBA{255, 255, 255, 255}, mplusSource)
			}
		}

	}

	// zombies
	for _, z := range g.zombies {

		op := &ebiten.DrawImageOptions{}

		w := z.img.Bounds().Dx()
		h := z.img.Bounds().Dy()

		op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

		op.GeoM.Rotate(z.angle + math.Pi)

		op.GeoM.Scale(z.s, z.s)

		op.GeoM.Translate(z.x+g.mapX, z.y+g.mapY)

		screen.DrawImage(z.img, op)

		if z.boss {
			// jauge de vie du boss
			vector.StrokeRect(screen, (float32(z.x)-100)+float32(g.mapX), (float32(z.y)-300)+float32(g.mapY), 260, 60, 10, color.RGBA{255, 255, 255, 255}, true)

			vector.DrawFilledRect(screen, ((float32(z.x)-100)+5)+float32(g.mapX), ((float32(z.y)-300)+5)+float32(g.mapY), float32(z.health)*(250/10), 54, color.RGBA{0, 255, 0, 255}, true)
		}
	}

	// player
	op := &ebiten.DrawImageOptions{}

	w := g.player.img.Bounds().Dx()
	h := g.player.img.Bounds().Dy()

	op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

	op.GeoM.Rotate(g.player.angle)

	op.GeoM.Scale(.5, .5)

	op.GeoM.Translate(g.player.x, g.player.y)

	screen.DrawImage(g.player.img, op)

	for _, b := range g.bullets {
		if b.img == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}

		if b.img == nil {
			print("")
		}

		w := b.img.Bounds().Dx()
		h := b.img.Bounds().Dy()

		op.GeoM.Translate(-float64(w)/2, -float64(h)/2)

		op.GeoM.Rotate(b.angle)

		op.GeoM.Scale(.14, .14)

		op.GeoM.Translate(b.x, b.y)

		screen.DrawImage(b.img, op)
	}

	// jauge de vie du player
	vector.StrokeRect(screen, 10, 10, 510, 60, 10, color.RGBA{255, 255, 255, 255}, true)

	vector.DrawFilledRect(screen, 15, 15, float32(g.player.lifes)*(500/100), 54, color.RGBA{0, 255, 0, 255}, true)

	if g.state == StateGameOver {
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 50, 50, 240}, true)

		gameutil.DrawText("GAME OVER", 260, screenWidth, 180, 300, 0, screen, color.RGBA{255, 0, 0, 255}, mplusSource)

		gameutil.DrawText("Press any key to restart a game!", 70, screenWidth-200, 200, screenHeight-300, 0, screen, color.RGBA{160, 160, 160, 255}, mplusSource)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFullscreen(true)

	diamondImg = loadImage("blue diamond.png")
	treeImg = loadImage("tree.png")
	cardImg = loadImage("card.png")
	zombieImg = loadImage("zombie.png")
	playerImg = loadImage("player.png")
	bulletImg = loadImage("bullet.png")
	fenceImg = loadImage("fence.png")
	bossImg = loadImage("zombieKing.png")

	s2, _ := text.NewGoTextFaceSource(bytes.NewReader(roboto))

	faceSource = s2
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))

	if err != nil {
		log.Fatal(err)
	}

	mplusSource = s

	g := &Game{
		state:        StatePlaying,
		bossCooldown: 1800,
		upgrades: map[string]int{
			"pierce":  0,
			"vampire": 0,
		},

		miniatureCard: miniatureCard{
			x:   1130,
			y:   20,
			s:   .3,
			img: cardImg,
		},

		player: player{
			img:           playerImg,
			x:             screenWidth / 2,
			y:             screenHeight / 2,
			r:             70,
			pickupRadius:  80, // <-- NOUVEAU : Valeur de départ
			cadence:       60,
			shootRange:    700,
			shootCooldown: 60,
			diamondQuota:  7,
			speed:         10,
			lifes:         100,
			clr:           color.RGBA{255, 255, 0, 255},
		},
	}

	for i := 0; i < 8; i++ {
		g.trees = append(g.trees, tree{
			x:   float64(rand.IntN(screenWidth+1800) + -1300),
			y:   float64(rand.IntN(screenHeight+1800) + -1300),
			s:   1,
			Img: treeImg,
		})
	}

	ebiten.SetWindowTitle("zombie_rush")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
