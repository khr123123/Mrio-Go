package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ==================== 所有可调参数都在这里！ =====================

const (
	// 窗口大小
	screenWidth  = 960
	screenHeight = 540

	// 马里奥水平移动速度
	playerSpeed = 3.0

	// 跳跃参数
	jumpVelocity = -9.0 // 跳跃初速度（负值越大跳越高）
	gravity      = 0.45 // 重力（每帧向下加速）
	maxFallSpeed = 10.0 // 最大下落速度

	// 地面基准位置（马里奥站立时的 playerY）
	groundY       = 420.0
	groundOffsetY = 0.0 // 额外偏移：正=抬高马里奥，负=降低

	// 马里奥整体缩放
	marioScale = 1.9

	// 动画速度（毫秒/帧）
	animationFrameTime = 60 * time.Millisecond

	// 马里奥精灵表切帧参数
	marioFrameWidth  = 16
	marioFrameHeight = 32
	marioNumFrames   = 3
	marioFrameStartX = 80
	marioFrameStartY = 0 // 动画起始 Y 行（0=第一行，32=第二行...）

	// 背景参数
	bgScale   = 2.4
	bgOffsetX = 200.0

	// 相机跟随平滑度
	cameraLerpFactor = 0.12

	// 初始玩家位置（世界坐标）
	initialPlayerX = 300.0

	// 怪物生成参数
	monsterSpawnTime   = 2 * time.Second        // 生成间隔
	monsterAnimSpeed   = 150 * time.Millisecond // 怪物动画速度
	monsterGroundY     = groundY + 20           // 怪物地面位置基准
	maxMonsters        = 15                     // 最大怪物数量
	monsterSpawnRangeX = 800.0                  // 怪物生成范围（相对于玩家）
)

// ================================================================

// MonsterType 怪物类型定义
type MonsterType struct {
	Name        string  // 怪物名称
	FrameWidth  int     // 帧宽度
	FrameHeight int     // 帧高度
	NumFrames   int     // 动画帧数
	StartX      int     // 精灵图起始X坐标
	StartY      int     // 精灵图起始Y坐标
	Scale       float64 // 缩放比例
	Speed       float64 // 移动速度
	YOffset     float64 // Y轴偏移量（用于调整地面高度）
}

// 预定义怪物类型配置
var monsterTypes = []MonsterType{
	// 怪物类型1：蘑菇怪（Goomba）
	{
		Name:        "Goomba",
		FrameWidth:  16,
		FrameHeight: 28,
		NumFrames:   2,
		StartX:      0,  // 精灵图中的X起始位置
		StartY:      12, // 精灵图中的Y起始位置
		Scale:       2.0,
		Speed:       1.2,
		YOffset:     0, // 与地面齐平
	},
}

// Monster 怪物实例
type Monster struct {
	monsterType   *MonsterType // 怪物类型
	x             float64
	y             float64
	vx            float64 // 水平速度（负=向左移动）
	frame         int
	lastFrameTime time.Time
	active        bool
}

type Game struct {
	marioImg         *ebiten.Image
	bgImg            *ebiten.Image
	monsterImg       *ebiten.Image
	cameraX          float64
	playerX          float64
	playerY          float64
	vy               float64 // 垂直速度
	onGround         bool    // 是否在地面
	canInfiniteJump  bool    // 无限连跳开关
	frame            int
	lastFrameTime    time.Time
	bgScale          float64
	monsters         []*Monster
	lastMonsterSpawn time.Time
}

func (g *Game) Update() error {
	// 水平移动
	dx := float64(0)
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx += 1
	}
	g.playerX += dx * playerSpeed

	// 跳跃逻辑
	if g.canInfiniteJump {
		if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
			g.vy = jumpVelocity
			g.onGround = false
		}
	} else {
		if g.onGround && (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)) {
			g.vy = jumpVelocity
			g.onGround = false
		}
	}

	// 切换无限连跳模式（按 J 键切换）
	if ebiten.IsKeyPressed(ebiten.KeyJ) {
		time.Sleep(150 * time.Millisecond)
		g.canInfiniteJump = !g.canInfiniteJump
	}

	// 重力
	g.vy += gravity
	if g.vy > maxFallSpeed {
		g.vy = maxFallSpeed
	}

	// 更新 Y 位置
	g.playerY += g.vy

	// 落地检测
	if g.playerY >= groundY+groundOffsetY {
		g.playerY = groundY + groundOffsetY
		g.vy = 0
		g.onGround = true
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
		g.frame = 0 // 站立帧
	}

	// 怪物生成逻辑
	if time.Since(g.lastMonsterSpawn) > monsterSpawnTime && len(g.monsters) < maxMonsters {
		g.spawnMonster()
		g.lastMonsterSpawn = time.Now()
	}

	// 更新怪物
	g.updateMonsters()

	return nil
}

// spawnMonster 生成新怪物（随机选择类型）
func (g *Game) spawnMonster() {
	// 随机选择怪物类型
	monsterType := &monsterTypes[rand.Intn(len(monsterTypes))]

	// 在玩家前方随机位置生成
	spawnX := g.playerX + monsterSpawnRangeX + rand.Float64()*300

	monster := &Monster{
		monsterType:   monsterType,
		x:             spawnX,
		y:             monsterGroundY + monsterType.YOffset,
		vx:            -monsterType.Speed, // 向左移动
		frame:         0,
		lastFrameTime: time.Now(),
		active:        true,
	}

	g.monsters = append(g.monsters, monster)
}

// updateMonsters 更新所有怪物
func (g *Game) updateMonsters() {
	for i := len(g.monsters) - 1; i >= 0; i-- {
		monster := g.monsters[i]

		if !monster.active {
			// 移除不活跃的怪物
			g.monsters = append(g.monsters[:i], g.monsters[i+1:]...)
			continue
		}

		// 移动怪物
		monster.x += monster.vx

		// 如果怪物离开屏幕太远，标记为不活跃
		if monster.x < g.cameraX-200 || monster.x > g.playerX+monsterSpawnRangeX+600 {
			monster.active = false
		}

		// 动画更新
		if time.Since(monster.lastFrameTime) > monsterAnimSpeed {
			monster.frame = (monster.frame + 1) % monster.monsterType.NumFrames
			monster.lastFrameTime = time.Now()
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
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

	// 绘制怪物
	g.drawMonsters(screen)

	// 马里奥
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Reset()

	drawX := g.playerX - g.cameraX
	drawY := g.playerY

	sx := marioFrameStartX + g.frame*marioFrameWidth
	sy := marioFrameStartY

	op.GeoM.Scale(marioScale, marioScale)

	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(marioFrameWidth)*marioScale, 0)
	}

	op.GeoM.Translate(drawX, drawY)

	subImg := g.marioImg.SubImage(image.Rect(sx, sy, sx+marioFrameWidth, sy+marioFrameHeight)).(*ebiten.Image)
	screen.DrawImage(subImg, op)

	// 调试信息
	jumpMode := "普通"
	if g.canInfiniteJump {
		jumpMode = "无限连跳"
	}

	// 统计各类怪物数量
	monsterStats := make(map[string]int)
	for _, m := range g.monsters {
		if m.active {
			monsterStats[m.monsterType.Name]++
		}
	}

	statsStr := ""
	for name, count := range monsterStats {
		statsStr += fmt.Sprintf("%s:%d ", name, count)
	}

	ebitenutil.DebugPrint(screen,
		fmt.Sprintf("FPS: %.1f | X:%.0f Y:%.0f | 跳跃:%s (J切换) | 怪物总数:%d\n%s",
			ebiten.ActualFPS(), g.playerX, g.playerY, jumpMode, len(g.monsters), statsStr))
}

// drawMonsters 绘制所有怪物
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

		// 只绘制在屏幕范围内的怪物
		if drawX < -100 || drawX > float64(screenWidth)+100 {
			continue
		}

		mType := monster.monsterType

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Reset()
		op.GeoM.Scale(mType.Scale, mType.Scale)
		op.GeoM.Translate(drawX, drawY)

		// 从精灵表中提取当前帧
		sx := mType.StartX + monster.frame*mType.FrameWidth
		sy := mType.StartY
		subImg := g.monsterImg.SubImage(image.Rect(
			sx, sy,
			sx+mType.FrameWidth,
			sy+mType.FrameHeight,
		)).(*ebiten.Image)

		screen.DrawImage(subImg, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Mario-Go")

	game := &Game{
		playerX:          initialPlayerX,
		playerY:          groundY + groundOffsetY,
		onGround:         true,
		canInfiniteJump:  true, // 默认开启无限连跳
		bgScale:          bgScale,
		lastFrameTime:    time.Now(),
		monsters:         make([]*Monster, 0),
		lastMonsterSpawn: time.Now(),
	}

	// 加载马里奥精灵表
	marioImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/mario_bros.png")
	if err != nil {
		log.Fatal(err)
	}
	game.marioImg = marioImg

	// 加载背景
	bgImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/level_1.png")
	if err != nil {
		log.Fatal(err)
	}
	game.bgImg = bgImg

	// 加载怪物精灵图
	monsterImg, _, err := ebitenutil.NewImageFromFile("resources/graphics/enemies.png")
	if err != nil {
		log.Println("警告：无法加载怪物图片，怪物将不显示")
	} else {
		game.monsterImg = monsterImg
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
