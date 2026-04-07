package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math/rand/v2"
	"os"
	"zombie_rush/bullet"
	"zombie_rush/card"
	"zombie_rush/diamond"
	"zombie_rush/player"
	"zombie_rush/zombie"

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

type miniatureCard struct {
	x, y, s float64
	img     *ebiten.Image
}

type tree struct {
	Img     *ebiten.Image
	x, y, s float64
}

type Game struct {
	bossCooldown        int
	state               int
	addZombieCooldown   float64
	zombieSpawnCooldown float64
	player              player.Player
	cards               []card.Card
	diamonds            []diamond.Diamond
	bullets             []bullet.Bullet
	zombies             []zombie.Zombie
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
	g.player.Lifes = 100
	g.player.PickupRadius = 100 // <-- NOUVEAU : Initialisation
	g.player.Diamond = 0
	g.player.DiamondQuota = 7
	g.player.Cadence = 60
	g.player.ShootCooldown = 60
	g.player.Speed = 10
	g.player.ShootRange = 700
	g.bossCooldown = 1800
	g.bullets = []bullet.Bullet{}
	g.diamonds = []diamond.Diamond{}
	g.cards = []card.Card{}
	g.zombies = []zombie.Zombie{}
	g.upgrades = map[string]int{
		"pierce":  0,
		"vampire": 0,
	}
	g.mapX = 0
	g.mapY = 0
}

func (g *Game) Update() error {
	if g.state == StatePlaying {

		if g.player.Lifes <= 0 {
			g.state = StateGameOver
		}

		if g.player.Lifes > 100 {
			g.player.Lifes = 100
		}

		if len(g.cards) == 0 {
			g.bossCooldown--

			player.Move(g.player.X, g.player.Y, g.player.R, &g.mapX, &g.mapY, g.player.Speed, g.bossCooldown, fenceImg, screenWidth, screenHeight)

			// Position du joueur dans le "monde" (en tenant compte de la caméra)
			playerWorldX := g.player.X - g.mapX
			playerWorldY := g.player.Y - g.mapY

			g.zombies = zombie.Movement(g.zombies, playerWorldX, playerWorldY)

			g.zombies = zombie.Spawn(g.zombies, &g.addZombieCooldown, zombieImg, screenWidth, screenHeight)

			g.bullets = bullet.Create(g.player.X, g.player.Y, playerWorldX, playerWorldY, &g.player.Angle, g.player.ShootRange, g.player.Cadence, g.zombies, g.bullets, &g.player.ShootCooldown, bulletImg)
			bullet.Move(g.bullets, g.player.X, g.player.Y)

			bulletHitZombie, zi, bi := bullet.HitZombie(g.zombies, g.bullets, g.mapX, g.mapY)

			if bulletHitZombie {
				diamond.Spawn(&g.diamonds, g.zombies[zi].X, g.zombies[zi].Y, 56, diamondImg)

				g.zombies, g.bullets = bullet.HitZombieReaction(zi, bi, g.zombies, g.bullets, g.upgrades, &g.player.Lifes, &g.bossCooldown)
			}

			diamondPickup, di := diamond.PickupRadius(g.player.X, g.player.Y, g.player.PickupRadius, g.mapX, g.mapY, g.diamonds)

			if diamondPickup {
				fmt.Println(lineSeparation)
				fmt.Println("diamond/player collision")

				g.diamonds[di].DetectedInPickupRadius = true
			}

			g.diamonds = diamond.DragToPlayer(g.diamonds, playerWorldX, playerWorldY, &g.player.Diamond)

			card.Create(&g.cards, &g.player.Diamond, &g.player.DiamondQuota, screenHeight, g.bossCooldown)

			zombie.Attack(playerWorldX, playerWorldY, g.player.R, &g.zombies, &g.player.Lifes)

			antiCheatLimit(&g.player.Cadence, &g.player.Speed)

			g.zombieSpawnCooldown--

			if g.bossCooldown <= 0 {

				g.trees = []tree{}

				zombie.UpdateBossPhase(&g.zombies, &g.bossCooldown, &g.player.Angle, g.player.X, g.player.Y, g.player.Speed, bossImg, &g.mapX, &g.mapY)
			}
		}
	} else {
		touches := inpututil.AppendJustPressedKeys(nil)

		if len(touches) > 0 {
			g.reset()
		}
	}

	g.upgrades = card.DetectClick(&g.cards, g.upgrades, &g.player.Cadence, &g.player.Speed, &g.player.ShootRange, &g.clicPrecedent, &g.player.PickupRadius)

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
		vector.StrokeCircle(screen, float32(g.player.X), float32(g.player.Y), float32(g.player.PickupRadius), 2, color.RGBA{0, 0, 120, 120}, true)

		for _, d := range g.diamonds {
			d.Draw(screen, g.mapX, g.mapY)
		}

		for _, t := range g.trees {
			op := &ebiten.DrawImageOptions{}

			op.GeoM.Scale(t.s, t.s)

			op.GeoM.Translate(t.x+g.mapX, t.y+g.mapY)

			screen.DrawImage(t.Img, op)
		}

		vector.StrokeRect(screen, 700, 50, 500, 60, 10, color.RGBA{255, 255, 255, 255}, true)

		vector.DrawFilledRect(screen, 700, 50, float32(g.player.Diamond*(500/g.player.DiamondQuota)), 54, color.RGBA{38, 115, 211, 255}, true)
		op := &ebiten.DrawImageOptions{}

		op.GeoM.Scale(g.miniatureCard.s, g.miniatureCard.s)

		op.GeoM.Translate(g.miniatureCard.x, g.miniatureCard.y)

		screen.DrawImage(g.miniatureCard.img, op)

		if len(g.cards) > 0 {

			vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 50, 50, 240}, false)

			for _, c := range g.cards {
				c.Draw(screen, mplusSource)
			}
		}

	}

	// zombies
	for _, z := range g.zombies {
		z.Draw(screen, g.mapX, g.mapY)

		if z.Boss {
			print("")
		}
	}

	// player
	g.player.Draw(screen)

	for _, b := range g.bullets {
		b.Draw(screen)
	}

	// jauge de vie du player
	vector.StrokeRect(screen, 10, 10, 510, 60, 10, color.RGBA{255, 255, 255, 255}, true)

	vector.DrawFilledRect(screen, 15, 15, float32(g.player.Lifes)*(500/100), 54, color.RGBA{0, 255, 0, 255}, true)

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
	}

	g.player.Initialization(playerImg, screenWidth, screenHeight)

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
