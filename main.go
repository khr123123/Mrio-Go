package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// ==================== 所有可调参数都在这里！ =====================

const (
	// 窗口大小
	screenWidth  = 960
	screenHeight = 540

	// 马里奥水平移动速度
	playerSpeed = 3.5

	// 跳跃参数
	jumpVelocity = -9.5
	gravity      = 0.45
	maxFallSpeed = 10.0

	// 地面基准位置
	groundY       = 420.0
	groundOffsetY = 0.0

	// 马里奥整体缩放
	marioScale = 1.9

	// 动画速度
	animationFrameTime = 60 * time.Millisecond

	// 马里奥精灵表切帧参数
	marioFrameWidth  = 16
	marioFrameHeight = 32
	marioNumFrames   = 3
	marioFrameStartX = 80
	marioFrameStartY = 0

	// 背景参数
	bgScale   = 2.4
	bgOffsetX = 200.0

	// 相机跟随平滑度
	cameraLerpFactor = 0.12

	// 初始玩家位置
	initialPlayerX = 300.0

	// 怪物生成参数
	monsterSpawnTime   = 2 * time.Second
	monsterAnimSpeed   = 150 * time.Millisecond
	monsterGroundY     = groundY + 20
	maxMonsters        = 15
	monsterSpawnRangeX = 800.0

	// 游戏参数
	initialLives     = 3
	invincibleTime   = 2 * time.Second
	killScoreBonus   = 100
	coinScoreBonus   = 50
	coinSpawnTime    = 3 * time.Second
	maxCoins         = 10
	powerUpSpawnTime = 8 * time.Second
	maxPowerUps      = 3
	superMarioTime   = 10 * time.Second
)

// ================================================================

// GameState 游戏状态
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// PowerUpType 道具类型
type PowerUpType int

const (
	PowerUpMushroom PowerUpType = iota // 蘑菇（变大）
	PowerUpStar                        // 星星（无敌）
)

// MonsterType 怪物类型定义
type MonsterType struct {
	Name        string
	FrameWidth  int
	FrameHeight int
	NumFrames   int
	StartX      int
	StartY      int
	Scale       float64
	Speed       float64
	YOffset     float64
	HitboxW     float64 // 碰撞箱宽度
	HitboxH     float64 // 碰撞箱高度
}

// 预定义怪物类型配置
var monsterTypes = []MonsterType{
	{
		Name:        "Goomba",
		FrameWidth:  16,
		FrameHeight: 28,
		NumFrames:   2,
		StartX:      0,
		StartY:      12,
		Scale:       2.0,
		Speed:       1.2,
		YOffset:     0,
		HitboxW:     28,
		HitboxH:     52,
	},
	{
		Name:        "Koopa",
		FrameWidth:  16,
		FrameHeight: 32,
		NumFrames:   2,
		StartX:      96,
		StartY:      8,
		Scale:       2.0,
		Speed:       1.0,
		YOffset:     -8,
		HitboxW:     28,
		HitboxH:     60,
	},
}

// Monster 怪物实例
type Monster struct {
	monsterType   *MonsterType
	x             float64
	y             float64
	vx            float64
	frame         int
	lastFrameTime time.Time
	active        bool
	dying         bool
	dyingTime     time.Time
}

// Coin 金币
type Coin struct {
	x             float64
	y             float64
	frame         int
	lastFrameTime time.Time
	active        bool
	collected     bool
}

// PowerUp 道具
type PowerUp struct {
	powerType     PowerUpType
	x             float64
	y             float64
	vy            float64
	frame         int
	lastFrameTime time.Time
	active        bool
	collected     bool
}

// Particle 粒子效果
type Particle struct {
	x       float64
	y       float64
	vx      float64
	vy      float64
	life    float64
	maxLife float64
	color   color.Color
}

type Game struct {
	// 图片资源
	marioImg   *ebiten.Image
	bgImg      *ebiten.Image
	monsterImg *ebiten.Image

	// 游戏状态
	state           GameState
	score           int
	lives           int
	coins           int
	highScore       int
	invincible      bool
	invincibleStart time.Time
	superMario      bool
	superMarioStart time.Time
	gameOverTime    time.Time
	canInfiniteJump bool
	godMode         bool // 上帝模式（无敌+穿墙）
	showHitbox      bool // 显示碰撞箱

	// 玩家状态
	cameraX       float64
	playerX       float64
	playerY       float64
	vy            float64
	onGround      bool
	frame         int
	lastFrameTime time.Time
	facingRight   bool

	// 背景
	bgScale float64

	// 游戏对象
	monsters         []*Monster
	lastMonsterSpawn time.Time
	coinsList        []*Coin
	lastCoinSpawn    time.Time
	powerUps         []*PowerUp
	lastPowerUpSpawn time.Time
	particles        []*Particle

	// 输入状态
	lastJumpPress  bool
	lastPauseKey   bool
	lastRestartKey bool
}

func (g *Game) Init() {
	g.state = StateMenu
	g.lives = initialLives
	g.score = 0
	g.coins = 0
	g.playerX = initialPlayerX
	g.playerY = groundY + groundOffsetY
	g.onGround = true
	g.canInfiniteJump = true
	g.bgScale = bgScale
	g.lastFrameTime = time.Now()
	g.monsters = make([]*Monster, 0)
	g.coinsList = make([]*Coin, 0)
	g.powerUps = make([]*PowerUp, 0)
	g.particles = make([]*Particle, 0)
	g.lastMonsterSpawn = time.Now()
	g.lastCoinSpawn = time.Now()
	g.lastPowerUpSpawn = time.Now()
	g.facingRight = true
	g.invincible = false
	g.superMario = false
	g.godMode = false
	g.showHitbox = false
}

func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		g.updatePaused()
	case StateGameOver:
		g.updateGameOver()
	}
	return nil
}

func (g *Game) updateMenu() {
	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Init()
		g.state = StatePlaying
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return
	}
}

func (g *Game) updatePaused() {
	pausePressed := ebiten.IsKeyPressed(ebiten.KeyP)
	if pausePressed && !g.lastPauseKey {
		g.state = StatePlaying
	}
	g.lastPauseKey = pausePressed
}

func (g *Game) updateGameOver() {
	restartPressed := ebiten.IsKeyPressed(ebiten.KeyR)
	if restartPressed && !g.lastRestartKey {
		g.Init()
		g.state = StatePlaying
	}
	g.lastRestartKey = restartPressed
}

func (g *Game) updatePlaying() {
	// 暂停
	pausePressed := ebiten.IsKeyPressed(ebiten.KeyP)
	if pausePressed && !g.lastPauseKey {
		g.state = StatePaused
	}
	g.lastPauseKey = pausePressed

	// 切换上帝模式 (G键)
	if ebiten.IsKeyPressed(ebiten.KeyG) {
		time.Sleep(150 * time.Millisecond)
		g.godMode = !g.godMode
	}

	// 切换显示碰撞箱 (H键)
	if ebiten.IsKeyPressed(ebiten.KeyH) {
		time.Sleep(150 * time.Millisecond)
		g.showHitbox = !g.showHitbox
	}

	// 水平移动
	dx := float64(0)
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1
		g.facingRight = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1
		g.facingRight = true
	}
	g.playerX += dx * playerSpeed

	// 跳跃逻辑
	jumpPressed := ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	if g.canInfiniteJump {
		if jumpPressed && !g.lastJumpPress {
			g.vy = jumpVelocity
			g.onGround = false
		}
	} else {
		if g.onGround && jumpPressed && !g.lastJumpPress {
			g.vy = jumpVelocity
			g.onGround = false
		}
	}
	g.lastJumpPress = jumpPressed

	// 切换无限连跳模式（J键）
	if ebiten.IsKeyPressed(ebiten.KeyJ) {
		time.Sleep(150 * time.Millisecond)
		g.canInfiniteJump = !g.canInfiniteJump
	}

	// 重力
	g.vy += gravity
	if g.vy > maxFallSpeed {
		g.vy = maxFallSpeed
	}

	// 更新Y位置
	g.playerY += g.vy

	// 落地检测
	if g.playerY >= groundY+groundOffsetY {
		g.playerY = groundY + groundOffsetY
		g.vy = 0
		g.onGround = true
	}

	// 死亡检测（掉出地图）
	if g.playerY > screenHeight+100 {
		g.loseLife()
	}

	// 相机跟随
	targetCameraX := g.playerX - float64(screenWidth)/3
	g.cameraX += (targetCameraX - g.cameraX) * cameraLerpFactor

	// 动画
	if dx != 0 || !g.onGround {
		if time.Since(g.lastFrameTime) > animationFrameTime {
			g.frame = (g.frame + 1) % marioNumFrames
			g.lastFrameTime = time.Now()
		}
	} else {
		g.frame = 0
	}

	// 生成怪物
	if time.Since(g.lastMonsterSpawn) > monsterSpawnTime && len(g.monsters) < maxMonsters {
		g.spawnMonster()
		g.lastMonsterSpawn = time.Now()
	}

	// 生成金币
	if time.Since(g.lastCoinSpawn) > coinSpawnTime && len(g.coinsList) < maxCoins {
		g.spawnCoin()
		g.lastCoinSpawn = time.Now()
	}

	// 生成道具
	if time.Since(g.lastPowerUpSpawn) > powerUpSpawnTime && len(g.powerUps) < maxPowerUps {
		g.spawnPowerUp()
		g.lastPowerUpSpawn = time.Now()
	}

	// 更新游戏对象
	g.updateMonsters()
	g.updateCoins()
	g.updatePowerUps()
	g.updateParticles()

	// 检测碰撞
	if !g.godMode {
		g.checkCollisions()
	}

	// 检查无敌状态
	if g.invincible && time.Since(g.invincibleStart) > invincibleTime {
		g.invincible = false
	}

	// 检查超级马里奥状态
	if g.superMario && time.Since(g.superMarioStart) > superMarioTime {
		g.superMario = false
	}
}

func (g *Game) spawnMonster() {
	monsterType := &monsterTypes[rand.Intn(len(monsterTypes))]
	spawnX := g.playerX + monsterSpawnRangeX + rand.Float64()*300

	monster := &Monster{
		monsterType:   monsterType,
		x:             spawnX,
		y:             monsterGroundY + monsterType.YOffset,
		vx:            -monsterType.Speed,
		frame:         0,
		lastFrameTime: time.Now(),
		active:        true,
		dying:         false,
	}

	g.monsters = append(g.monsters, monster)
}

func (g *Game) spawnCoin() {
	spawnX := g.playerX + monsterSpawnRangeX + rand.Float64()*400
	spawnY := groundY - 100 - rand.Float64()*200

	coin := &Coin{
		x:             spawnX,
		y:             spawnY,
		frame:         0,
		lastFrameTime: time.Now(),
		active:        true,
		collected:     false,
	}

	g.coinsList = append(g.coinsList, coin)
}

func (g *Game) spawnPowerUp() {
	spawnX := g.playerX + monsterSpawnRangeX + rand.Float64()*500
	spawnY := groundY - 50

	powerType := PowerUpMushroom
	if rand.Float64() < 0.3 {
		powerType = PowerUpStar
	}

	powerUp := &PowerUp{
		powerType:     powerType,
		x:             spawnX,
		y:             spawnY,
		vy:            0,
		frame:         0,
		lastFrameTime: time.Now(),
		active:        true,
		collected:     false,
	}

	g.powerUps = append(g.powerUps, powerUp)
}

func (g *Game) updateMonsters() {
	for i := len(g.monsters) - 1; i >= 0; i-- {
		monster := g.monsters[i]

		if !monster.active {
			g.monsters = append(g.monsters[:i], g.monsters[i+1:]...)
			continue
		}

		if monster.dying {
			if time.Since(monster.dyingTime) > 500*time.Millisecond {
				monster.active = false
			}
			continue
		}

		monster.x += monster.vx

		if monster.x < g.cameraX-200 || monster.x > g.playerX+monsterSpawnRangeX+600 {
			monster.active = false
		}

		if time.Since(monster.lastFrameTime) > monsterAnimSpeed {
			monster.frame = (monster.frame + 1) % monster.monsterType.NumFrames
			monster.lastFrameTime = time.Now()
		}
	}
}

func (g *Game) updateCoins() {
	for i := len(g.coinsList) - 1; i >= 0; i-- {
		coin := g.coinsList[i]

		if !coin.active {
			g.coinsList = append(g.coinsList[:i], g.coinsList[i+1:]...)
			continue
		}

		if coin.x < g.cameraX-200 || coin.x > g.playerX+monsterSpawnRangeX+800 {
			coin.active = false
		}

		if time.Since(coin.lastFrameTime) > 100*time.Millisecond {
			coin.frame = (coin.frame + 1) % 4
			coin.lastFrameTime = time.Now()
		}
	}
}

func (g *Game) updatePowerUps() {
	for i := len(g.powerUps) - 1; i >= 0; i-- {
		powerUp := g.powerUps[i]

		if !powerUp.active {
			g.powerUps = append(g.powerUps[:i], g.powerUps[i+1:]...)
			continue
		}

		powerUp.vy += gravity * 0.5
		if powerUp.vy > maxFallSpeed {
			powerUp.vy = maxFallSpeed
		}
		powerUp.y += powerUp.vy

		if powerUp.y >= groundY {
			powerUp.y = groundY
			powerUp.vy = 0
		}

		if powerUp.x < g.cameraX-200 || powerUp.x > g.playerX+monsterSpawnRangeX+800 {
			powerUp.active = false
		}

		if time.Since(powerUp.lastFrameTime) > 150*time.Millisecond {
			powerUp.frame = (powerUp.frame + 1) % 2
			powerUp.lastFrameTime = time.Now()
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.2
		p.life -= 0.02

		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) checkCollisions() {
	// 玩家碰撞箱
	playerW := float64(marioFrameWidth) * marioScale * 0.7
	playerH := float64(marioFrameHeight) * marioScale * 0.8
	playerLeft := g.playerX + playerW*0.15
	playerRight := g.playerX + playerW*1.15
	playerTop := g.playerY
	playerBottom := g.playerY + playerH

	// 检测与怪物的碰撞
	for _, monster := range g.monsters {
		if !monster.active || monster.dying {
			continue
		}

		mLeft := monster.x
		mRight := monster.x + monster.monsterType.HitboxW
		mTop := monster.y
		mBottom := monster.y + monster.monsterType.HitboxH

		if playerRight > mLeft && playerLeft < mRight &&
			playerBottom > mTop && playerTop < mBottom {

			// 从上方踩踏怪物
			if g.vy > 0 && playerBottom-mTop < 20 {
				g.killMonster(monster)
				g.vy = jumpVelocity * 0.6
				g.score += killScoreBonus
			} else if !g.invincible {
				// 被怪物伤害
				g.loseLife()
			}
		}
	}

	// 检测与金币的碰撞
	for _, coin := range g.coinsList {
		if !coin.active || coin.collected {
			continue
		}

		coinW := 30.0
		coinH := 30.0
		cLeft := coin.x
		cRight := coin.x + coinW
		cTop := coin.y
		cBottom := coin.y + coinH

		if playerRight > cLeft && playerLeft < cRight &&
			playerBottom > cTop && playerTop < cBottom {
			g.collectCoin(coin)
		}
	}

	// 检测与道具的碰撞
	for _, powerUp := range g.powerUps {
		if !powerUp.active || powerUp.collected {
			continue
		}

		pW := 32.0
		pH := 32.0
		pLeft := powerUp.x
		pRight := powerUp.x + pW
		pTop := powerUp.y
		pBottom := powerUp.y + pH

		if playerRight > pLeft && playerLeft < pRight &&
			playerBottom > pTop && playerTop < pBottom {
			g.collectPowerUp(powerUp)
		}
	}
}

func (g *Game) killMonster(monster *Monster) {
	monster.dying = true
	monster.dyingTime = time.Now()

	// 生成粒子效果
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		speed := 2.0 + rand.Float64()*2
		g.particles = append(g.particles, &Particle{
			x:       monster.x + monster.monsterType.HitboxW/2,
			y:       monster.y + monster.monsterType.HitboxH/2,
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle)*speed - 2,
			life:    1.0,
			maxLife: 1.0,
			color:   color.RGBA{255, 200, 0, 255},
		})
	}
}

func (g *Game) collectCoin(coin *Coin) {
	coin.collected = true
	coin.active = false
	g.coins++
	g.score += coinScoreBonus

	// 生成粒子效果
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3
		speed := 1.5 + rand.Float64()
		g.particles = append(g.particles, &Particle{
			x:       coin.x + 15,
			y:       coin.y + 15,
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle)*speed - 1,
			life:    1.0,
			maxLife: 1.0,
			color:   color.RGBA{255, 215, 0, 255},
		})
	}
}

func (g *Game) collectPowerUp(powerUp *PowerUp) {
	powerUp.collected = true
	powerUp.active = false

	switch powerUp.powerType {
	case PowerUpMushroom:
		g.superMario = true
		g.superMarioStart = time.Now()
		g.score += 200
	case PowerUpStar:
		g.invincible = true
		g.invincibleStart = time.Now()
		g.score += 300
	}

	// 生成粒子效果
	for i := 0; i < 10; i++ {
		angle := float64(i) * math.Pi / 5
		speed := 2.0 + rand.Float64()*2
		particleColor := color.RGBA{255, 0, 255, 255}
		if powerUp.powerType == PowerUpStar {
			particleColor = color.RGBA{255, 255, 0, 255}
		}
		g.particles = append(g.particles, &Particle{
			x:       powerUp.x + 16,
			y:       powerUp.y + 16,
			vx:      math.Cos(angle) * speed,
			vy:      math.Sin(angle)*speed - 2,
			life:    1.0,
			maxLife: 1.0,
			color:   particleColor,
		})
	}
}

func (g *Game) loseLife() {
	if g.invincible || g.godMode {
		return
	}

	g.lives--
	if g.lives <= 0 {
		g.state = StateGameOver
		g.gameOverTime = time.Now()
		if g.score > g.highScore {
			g.highScore = g.score
		}
	} else {
		// 重生
		g.invincible = true
		g.invincibleStart = time.Now()
		g.playerX = g.cameraX + float64(screenWidth)/3
		g.playerY = groundY + groundOffsetY
		g.vy = 0
		g.onGround = true
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawGame(screen)
		g.drawHUD(screen)
	case StatePaused:
		g.drawGame(screen)
		g.drawHUD(screen)
		g.drawPaused(screen)
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 100, G: 150, B: 255, A: 255})

	title := "SUPER MARIO GO"
	drawTextCentered(screen, title, screenWidth/2, screenHeight/3, 3, color.White)

	instructions := []string{
		"Press ENTER or SPACE to Start",
		"",
		"Controls:",
		"A/D or Arrow Keys - Move",
		"SPACE or UP - Jump",
		"J - Toggle Infinite Jump",
		"P - Pause",
		"G - God Mode",
		"H - Show Hitbox",
	}

	y := screenHeight/2 + 20
	for _, line := range instructions {
		drawTextCentered(screen, line, screenWidth/2, y, 1, color.White)
		y += 25
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// 背景
	bgW := float64(g.bgImg.Bounds().Dx())
	bgH := float64(g.bgImg.Bounds().Dy())
	scaledW := bgW * g.bgScale
	scaledH := bgH * g.bgScale

	offsetY := (float64(screenHeight) - scaledH) / 2
	if offsetY < 0 {
		offsetY = 0
	}

	x := -g.cameraX*g.bgScale - bgOffsetX
	for x < float64(screenWidth) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		op.GeoM.Scale(g.bgScale, g.bgScale)
		op.GeoM.Translate(x, offsetY)
		screen.DrawImage(g.bgImg, op)
		x += scaledW
	}

	// 绘制粒子
	g.drawParticles(screen)

	// 绘制金币
	g.drawCoins(screen)

	// 绘制道具
	g.drawPowerUps(screen)

	// 绘制怪物
	g.drawMonsters(screen)

	// 绘制马里奥
	g.drawMario(screen)
}

func (g *Game) drawMario(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()

	drawX := g.playerX - g.cameraX
	drawY := g.playerY

	sx := marioFrameStartX + g.frame*marioFrameWidth
	sy := marioFrameStartY

	scale := marioScale
	if g.superMario {
		scale *= 1.3
	}

	op.GeoM.Scale(scale, scale)

	if !g.facingRight {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(marioFrameWidth)*scale, 0)
	}

	op.GeoM.Translate(drawX, drawY)

	// 无敌闪烁效果
	if g.invincible && int(time.Since(g.invincibleStart).Milliseconds()/100)%2 == 0 {
		op.ColorScale.ScaleAlpha(0.5)
	}

	subImg := g.marioImg.SubImage(image.Rect(sx, sy, sx+marioFrameWidth, sy+marioFrameHeight)).(*ebiten.Image)
	screen.DrawImage(subImg, op)

	// 显示碰撞箱
	if g.showHitbox {
		g.drawHitbox(screen, drawX, drawY,
			float64(marioFrameWidth)*marioScale*0.7,
			float64(marioFrameHeight)*marioScale*0.8,
			color.RGBA{0, 255, 0, 128})
	}
}

func (g *Game) drawMonsters(screen *ebiten.Image) {
	if g.monsterImg == nil {
		return
	}

	for _, monster := range g.monsters {
		if !monster.active {
			continue
		}

		drawX := monster.x - g.cameraX
		drawY := monster.y

		if drawX < -100 || drawX > float64(screenWidth)+100 {
			continue
		}

		mType := monster.monsterType

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		op.GeoM.Scale(mType.Scale, mType.Scale)

		if monster.dying {
			op.GeoM.Rotate(math.Pi)
			op.GeoM.Translate(float64(mType.FrameWidth)*mType.Scale/2, float64(mType.FrameHeight)*mType.Scale/2)
			op.ColorScale.ScaleAlpha(0.5)
		}

		op.GeoM.Translate(drawX, drawY)

		sx := mType.StartX + monster.frame*mType.FrameWidth
		sy := mType.StartY
		subImg := g.monsterImg.SubImage(image.Rect(
			sx, sy,
			sx+mType.FrameWidth,
			sy+mType.FrameHeight,
		)).(*ebiten.Image)

		screen.DrawImage(subImg, op)

		// 显示碰撞箱
		if g.showHitbox {
			g.drawHitbox(screen, drawX, drawY, mType.HitboxW, mType.HitboxH, color.RGBA{255, 0, 0, 128})
		}
	}
}

func (g *Game) drawCoins(screen *ebiten.Image) {
	for _, coin := range g.coinsList {
		if !coin.active || coin.collected {
			continue
		}

		drawX := coin.x - g.cameraX
		drawY := coin.y

		if drawX < -100 || drawX > float64(screenWidth)+100 {
			continue
		}

		// 简单的金币动画（闪烁效果）
		size := 30.0
		alpha := uint8(200 + 55*math.Sin(float64(coin.frame)*math.Pi/2))

		ebitenutil.DrawRect(screen, drawX, drawY, size, size, color.RGBA{255, 215, 0, alpha})
		ebitenutil.DrawRect(screen, drawX+5, drawY+5, size-10, size-10, color.RGBA{255, 255, 100, alpha})
	}
}

func (g *Game) drawPowerUps(screen *ebiten.Image) {
	for _, powerUp := range g.powerUps {
		if !powerUp.active || powerUp.collected {
			continue
		}

		drawX := powerUp.x - g.cameraX
		drawY := powerUp.y

		if drawX < -100 || drawX > float64(screenWidth)+100 {
			continue
		}

		var col color.RGBA
		switch powerUp.powerType {
		case PowerUpMushroom:
			col = color.RGBA{255, 0, 0, 255}
		case PowerUpStar:
			col = color.RGBA{255, 255, 0, 255}
		}

		// 简单的道具渲染
		ebitenutil.DrawRect(screen, drawX, drawY, 32, 32, col)
		ebitenutil.DrawRect(screen, drawX+4, drawY+4, 24, 24, color.RGBA{255, 255, 255, 200})
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		drawX := p.x - g.cameraX
		drawY := p.y

		alpha := uint8(255 * p.life / p.maxLife)
		col := p.color.(color.RGBA)
		col.A = alpha

		size := 4.0 * p.life / p.maxLife
		ebitenutil.DrawRect(screen, drawX, drawY, size, size, col)
	}
}

func (g *Game) drawHitbox(screen *ebiten.Image, x, y, w, h float64, col color.Color) {
	// 绘制碰撞箱边框
	thickness := 2.0
	ebitenutil.DrawRect(screen, x, y, w, thickness, col)             // 上
	ebitenutil.DrawRect(screen, x, y+h-thickness, w, thickness, col) // 下
	ebitenutil.DrawRect(screen, x, y, thickness, h, col)             // 左
	ebitenutil.DrawRect(screen, x+w-thickness, y, thickness, h, col) // 右
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// 状态栏背景
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, 40, color.RGBA{0, 0, 0, 180})

	// 分数
	scoreText := fmt.Sprintf("Score: %d", g.score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 10, 20, color.White)

	// 生命
	livesText := fmt.Sprintf("Lives: %d", g.lives)
	text.Draw(screen, livesText, basicfont.Face7x13, 150, 20, color.White)

	// 金币
	coinsText := fmt.Sprintf("Coins: %d", g.coins)
	text.Draw(screen, coinsText, basicfont.Face7x13, 280, 20, color.White)

	// 状态显示
	statusX := 420
	if g.invincible {
		remaining := int(invincibleTime.Seconds() - time.Since(g.invincibleStart).Seconds())
		text.Draw(screen, fmt.Sprintf("INVINCIBLE:%d", remaining), basicfont.Face7x13, statusX, 20, color.RGBA{255, 255, 0, 255})
	}
	if g.superMario {
		remaining := int(superMarioTime.Seconds() - time.Since(g.superMarioStart).Seconds())
		text.Draw(screen, fmt.Sprintf("SUPER:%d", remaining), basicfont.Face7x13, statusX+150, 20, color.RGBA{255, 100, 255, 255})
	}

	// 调试信息
	debugY := 50
	if g.godMode {
		text.Draw(screen, "[GOD MODE]", basicfont.Face7x13, 10, debugY, color.RGBA{255, 0, 0, 255})
	}
	if g.canInfiniteJump {
		text.Draw(screen, "[∞ Jump]", basicfont.Face7x13, 120, debugY, color.RGBA{0, 255, 255, 255})
	}

	// FPS和怪物统计
	monsterStats := make(map[string]int)
	for _, m := range g.monsters {
		if m.active {
			monsterStats[m.monsterType.Name]++
		}
	}

	statsStr := fmt.Sprintf("FPS:%.0f Monsters:%d", ebiten.ActualFPS(), len(g.monsters))
	text.Draw(screen, statsStr, basicfont.Face7x13, screenWidth-300, 20, color.White)
}

func (g *Game) drawPaused(screen *ebiten.Image) {
	// 半透明遮罩
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150})

	drawTextCentered(screen, "PAUSED", screenWidth/2, screenHeight/2-30, 2, color.White)
	drawTextCentered(screen, "Press P to Resume", screenWidth/2, screenHeight/2+20, 1, color.White)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// 半透明遮罩
	ebitenutil.DrawRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 180})

	drawTextCentered(screen, "GAME OVER", screenWidth/2, screenHeight/2-60, 3, color.RGBA{255, 50, 50, 255})

	finalScore := fmt.Sprintf("Final Score: %d", g.score)
	drawTextCentered(screen, finalScore, screenWidth/2, screenHeight/2, 2, color.White)

	if g.score == g.highScore && g.highScore > 0 {
		drawTextCentered(screen, "NEW HIGH SCORE!", screenWidth/2, screenHeight/2+40, 1, color.RGBA{255, 215, 0, 255})
	} else if g.highScore > 0 {
		highScoreText := fmt.Sprintf("High Score: %d", g.highScore)
		drawTextCentered(screen, highScoreText, screenWidth/2, screenHeight/2+40, 1, color.RGBA{200, 200, 200, 255})
	}

	drawTextCentered(screen, "Press R to Restart", screenWidth/2, screenHeight/2+100, 1, color.White)
}

func drawTextCentered(screen *ebiten.Image, str string, cx, cy int, scale float64, col color.Color) {
	bounds := text.BoundString(basicfont.Face7x13, str)
	w := bounds.Dx()
	x := cx - int(float64(w)*scale/2)
	y := cy

	// 简单的缩放效果（绘制多次）
	if scale > 1 {
		for dx := 0; dx < int(scale); dx++ {
			for dy := 0; dy < int(scale); dy++ {
				text.Draw(screen, str, basicfont.Face7x13, x+dx, y+dy, col)
			}
		}
	} else {
		text.Draw(screen, str, basicfont.Face7x13, x, y, col)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Super Mario Go - Enhanced")

	game := &Game{}
	game.Init()

	// 加载马里奥精灵表
	marioImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/mario_bros.png")
	if err != nil {
		log.Fatal("无法加载马里奥图片:", err)
	}
	game.marioImg = marioImg

	// 加载背景
	bgImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/level_1.png")
	if err != nil {
		log.Fatal("无法加载背景图片:", err)
	}
	game.bgImg = bgImg

	// 加载怪物精灵图
	monsterImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/enemies.png")
	if err != nil {
		log.Println("警告：无法加载怪物图片，怪物将显示为占位符")
	} else {
		game.monsterImg = monsterImg
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
