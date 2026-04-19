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
	"zombie_rush/ice"
	"zombie_rush/player"
	"zombie_rush/zombie"

	"github.com/JeremyBelleIsle/gameutil"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
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

type Shake struct {
	duration  int
	intensity float64
}

type Game struct {
	bossCooldown          int
	bossCnt               int
	setBossCntValue       int
	state                 int
	setSpawnZombieCadence float64
	addZombieCooldown     float64
	player                player.Player
	cards                 []card.Card
	diamonds              []diamond.Diamond
	bullets               []bullet.Bullet
	zombies               []zombie.Zombie
	trees                 []tree
	BooomSnd              *audio.Player
	BossAlarmSnd          *audio.Player
	tingSnd               *audio.Player
	impactSnd             *audio.Player
	GameOverSnd           *audio.Player
	principalMusic        *audio.Player
	ice                   ice.Ice
	miniatureCard         miniatureCard
	upgrades              map[string]int
	clicPrecedent         bool
	mapX, mapY            float64
	musicStarted          bool
	shake                 Shake
}

var (
	// images
	diamondImg   *ebiten.Image
	treeImg      *ebiten.Image
	cardImg      *ebiten.Image
	zombieImg    *ebiten.Image
	playerImg    *ebiten.Image
	bulletImg    *ebiten.Image
	fenceImg     *ebiten.Image
	bossImg      *ebiten.Image
	iceCircleImg *ebiten.Image
)

// sound

var audioContext *audio.Context

// fonts

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
	g.setSpawnZombieCadence = 60
	g.player.Lifes = 100
	g.player.MaxHealth = 100
	g.player.PickupRadius = 100
	g.player.Diamond = 0
	g.player.DiamondQuota = 7
	g.player.Cadence = 60
	g.player.ShootCooldown = 60
	g.player.Speed = 10
	g.player.ShootRange = 700
	g.bossCooldown = 1800
	g.bossCnt = 0
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
	g.musicStarted = false
	g.shake = Shake{}
}

func (g *Game) Update() error {
	// Play principal music on loop
	if !g.musicStarted {
		g.musicStarted = true
		g.principalMusic.Rewind()
		g.principalMusic.Play()
	}

	// Restart music if it finished
	if !g.principalMusic.IsPlaying() && g.musicStarted {
		g.principalMusic.Rewind()
		g.principalMusic.Play()
	}

	if g.ice.Life > 0 {
		g.ice.Life--
	} else {
		g.ice = ice.Ice{}
	}

	if g.shake.duration > 0 {
		g.shake.duration--
	}

	if g.state == StatePlaying {

		if g.player.Lifes < 0 && g.player.Lifes != -1000 {
			g.state = StateGameOver
			g.player.Lifes = -1000
			g.GameOverSnd.Rewind()
			g.GameOverSnd.Play()
		}

		if g.player.Lifes > g.player.MaxHealth {
			g.player.Lifes = g.player.MaxHealth
		}

		if len(g.cards) == 0 {
			g.setSpawnZombieCadence -= .0006
			g.bossCooldown--

			player.Move(g.player.X, g.player.Y, g.player.R, &g.mapX, &g.mapY, g.player.Speed, g.bossCooldown, fenceImg, screenWidth, screenHeight)

			// Position du joueur dans le "monde" (en tenant compte de la caméra)
			playerWorldX := g.player.X - g.mapX
			playerWorldY := g.player.Y - g.mapY

			g.zombies = zombie.Movement(g.zombies, playerWorldX, playerWorldY, g.ice)

			g.zombies = zombie.Spawn(g.zombies, &g.addZombieCooldown, g.setSpawnZombieCadence, zombieImg, screenWidth, screenHeight, g.mapX, g.mapY)

			previewSlice := g.bullets
			g.bullets = bullet.Create(g.player.X, g.player.Y, playerWorldX, playerWorldY, &g.player.Angle, g.player.ShootRange, g.player.Cadence, g.zombies, g.bullets, &g.player.ShootCooldown, bulletImg)
			if len(previewSlice) != len(g.bullets) {
				g.BooomSnd.Rewind()
				g.BooomSnd.Play()
			}
			bullet.Move(g.bullets, g.player.X, g.player.Y)

			bulletHitZombie, zi, bi := bullet.HitZombie(g.zombies, g.bullets, g.mapX, g.mapY)

			if bulletHitZombie {
				diamond.Spawn(&g.diamonds, g.zombies[zi].X, g.zombies[zi].Y, 56, diamondImg, g.bossCooldown)

				g.zombies, g.bullets, g.ice = bullet.HitZombieReaction(zi, bi, g.zombies, g.bullets, g.upgrades, &g.player.Lifes, &g.bossCooldown, iceCircleImg, g.ice)
			}

			diamondPickup, di := diamond.PickupRadius(g.player.X, g.player.Y, g.player.PickupRadius, g.mapX, g.mapY, g.diamonds, g.bossCooldown)

			if diamondPickup {
				fmt.Println(lineSeparation)
				fmt.Println("diamond/player collision")

				g.diamonds[di].DetectedInPickupRadius = true
			}

			collect := false

			g.diamonds, collect = diamond.DragToPlayer(g.diamonds, playerWorldX, playerWorldY, &g.player.Diamond, g.bossCooldown)

			if collect {
				g.tingSnd.Rewind()
				g.tingSnd.Play()
			}

			card.Create(&g.cards, &g.player.Diamond, &g.player.DiamondQuota, screenHeight, g.bossCooldown, false)

			oldLife := g.player.Lifes
			zombie.Attack(playerWorldX, playerWorldY, g.player.R, &g.zombies, &g.player.Lifes, g.impactSnd)
			if g.player.Lifes < oldLife {
				g.shake.duration = 22
				g.shake.intensity = 30
			}

			antiCheatLimit(&g.player.Cadence, &g.player.Speed)

			if g.bossCooldown <= 0 {

				g.trees = []tree{}

				if g.bossCooldown == 0 {
					g.bossCnt++
				}

				zombie.UpdateBossPhase(&g.zombies, &g.bossCooldown, &g.player.Angle, g.player.X, g.player.Y, g.player.Speed, bossImg, &g.mapX, &g.mapY, g.BossAlarmSnd)
			}

			if g.bossCnt == 1 && g.bossCooldown > 0 {
				g.setBossCntValue--
				g.bossCnt = 0 + g.setBossCntValue
				card.Create(&g.cards, &g.player.Diamond, &g.player.DiamondQuota, screenHeight, g.bossCooldown, true)
			}

		}
	} else {
		touches := inpututil.AppendJustPressedKeys(nil)

		if len(touches) > 0 {
			g.reset()
		}
	}

	g.upgrades = card.DetectClick(&g.cards, g.upgrades, &g.player.Cadence, &g.player.Speed, &g.player.ShootRange, &g.clicPrecedent, &g.player.PickupRadius, &g.player.MaxHealth, &g.player.Lifes)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var shakeX, shakeY float64
	if g.shake.duration > 0 {
		shakeX = (rand.Float64()*2 - 1) * g.shake.intensity
		shakeY = (rand.Float64()*2 - 1) * g.shake.intensity
	}

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
		op.GeoM.Translate(screenWidth/2+g.mapX+shakeX, screenHeight/2+g.mapY+shakeY)

		screen.DrawImage(fenceImg, op)

	} else {

		// debug
		vector.StrokeCircle(screen, float32(g.player.X+shakeX), float32(g.player.Y+shakeY), float32(g.player.PickupRadius), 2, color.RGBA{0, 0, 120, 120}, true)

		for _, d := range g.diamonds {
			d.Draw(screen, g.mapX+shakeX, g.mapY+shakeY)
		}

		for _, t := range g.trees {
			op := &ebiten.DrawImageOptions{}

			op.GeoM.Scale(t.s, t.s)

			op.GeoM.Translate(t.x+g.mapX+shakeX, t.y+g.mapY+shakeY)

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
		z.Draw(screen, g.mapX+shakeX, g.mapY+shakeY)

		if z.Boss {
			print("")
		}
	}

	// player
	player.Player.Draw(g.player, screen, g.shake.intensity, g.shake.duration)

	for _, b := range g.bullets {
		b.Draw(screen)
	}

	// jauge de vie du player
	vector.StrokeRect(screen, 10, 10, 510, 60, 10, color.RGBA{255, 255, 255, 255}, true)

	vector.DrawFilledRect(screen, 15, 15, float32(g.player.Lifes)*(500/float32(g.player.MaxHealth)), 54, color.RGBA{0, 255, 0, 255}, true)

	if g.state == StateGameOver {
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 50, 50, 240}, true)

		gameutil.DrawText("GAME OVER", 260, screenWidth, 180, 300, 0, screen, color.RGBA{255, 0, 0, 255}, mplusSource)

		gameutil.DrawText("Press any key to restart a game!", 70, screenWidth-200, 200, screenHeight-300, 0, screen, color.RGBA{160, 160, 160, 255}, mplusSource)
	}

	g.ice.Draw(screen, g.mapX+shakeX, g.mapY+shakeY)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	fmt.Println("game started with success")

	audioContext = audio.NewContext(44100)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetFullscreen(true)

	s2, _ := text.NewGoTextFaceSource(bytes.NewReader(roboto))

	faceSource = s2
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))

	if err != nil {
		log.Fatal(err)
	}

	mplusSource = s

	// load images
	diamondImg = loadImage("blue diamond.png")
	treeImg = loadImage("tree.png")
	cardImg = loadImage("card.png")
	zombieImg = loadImage("zombie.png")
	playerImg = loadImage("player.png")
	bulletImg = loadImage("bullet.png")
	fenceImg = loadImage("fence.png")
	bossImg = loadImage("zombieKing.png")
	iceCircleImg = loadImage("iceCircle.png")

	g := &Game{
		state:                 StatePlaying,
		bossCooldown:          100,
		setSpawnZombieCadence: 60,
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
	// load sounds
	g.BooomSnd, err = gameutil.LoadSound(44100, audioContext, "miniBoom.wav")
	g.BossAlarmSnd, err = gameutil.LoadSound(44100, audioContext, "alarm.mp3")
	g.tingSnd, err = gameutil.LoadSound(44100, audioContext, "ting.mp3")
	g.impactSnd, err = gameutil.LoadSound(44100, audioContext, "impact.wav")
	g.GameOverSnd, err = gameutil.LoadSound(44100, audioContext, "gameOver.wav")
	g.principalMusic, err = gameutil.LoadSound(44100, audioContext, "principal music.wav")

	g.principalMusic.SetVolume(.2)

	if err != nil {
		panic(err)
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

	fmt.Println("game build started with success")

	ebiten.SetWindowTitle("zombie_rush")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
