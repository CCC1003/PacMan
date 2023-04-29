package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/danicat/simpleansi"
	"github.com/eiannone/keyboard"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

// 迷宫
var maze []string

// 命令行参数的解析
var (
	configFile = flag.String("config-file", "config.json", "path to custom configuration file")
	mazeFile   = flag.String("maze-file", "maze01.txt", "path to a custom maze file")
)

// 坐标类型结构体
type sprite struct {
	row      int
	col      int
	startRow int
	startCol int
}

// 配置包含表情符号的配置
type Config struct {
	Player           string        `json:"player"`
	Ghost            string        `json:"ghost"`
	Wall             string        `json:"wall"`
	Dot              string        `json:"dot"`
	Pill             string        `json:"pill"`
	Death            string        `json:"death"`
	Space            string        `json:"space"`
	UseEmoji         bool          `json:"use_emoji"`
	GhostRight       string        `json:"ghost_right"`
	PillDurationSecs time.Duration `json:"pill_duration_secs"`
}

// 玩家位置
var player sprite

// 敌人坐标数组
var ghosts []*ghost

// 分数
var score int

// 豆子
var numDots int

// 生命值
var lives = 3

// email配置
var cfg Config

// 敌人的状态
type GhostStatus string

// 敌人状态常量
const (
	GhostStatusNormal GhostStatus = "Normal"
	GhostStatusRight  GhostStatus = "Blue"
)

// 敌人结构体 含位置信息和状态信息
type ghost struct {
	position sprite
	status   GhostStatus
}

// 药丸计时器
var pillTimer *time.Timer

// 互斥锁
var pillMx sync.Mutex

// 读写互斥锁
var ghostsStatusMx sync.RWMutex

// 加载配置
func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}
	return nil
}

// 加载迷宫
func loadMaze(file string) error {
	//Open（） 函数返回一对值：一个文件和一个错误。
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	//函数结束时延迟后调用Close()函数
	defer f.Close()
	//扫描仪是读取文件的一种非常方便的方式
	scanner := bufio.NewScanner(f)
	//Scan（） 将在从文件和扫描器中读取某些内容时返回 true。
	//Text（） 将返回下一行输入。
	for scanner.Scan() {
		line := scanner.Text()
		//追加内置函数负责向迷宫切片添加新元素。
		maze = append(maze, line)
	}
	for row, line := range maze {
		for col, char := range line {
			switch char {
			case 'P':
				player = sprite{row, col, row, col}
				fmt.Println(player.row, player.col)
			case 'G':
				ghosts = append(ghosts, &ghost{sprite{row, col, row, col}, GhostStatusNormal})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

// 遍历迷宫切片中的每个条目并打印出来。
func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '#':
				fmt.Print(simpleansi.WithBlueBackground(cfg.Wall))
			case '.':
				fmt.Print(cfg.Dot)
			case 'X':
				fmt.Print(cfg.Pill)
			default:
				fmt.Print(cfg.Space)
			}
		}
		fmt.Println()
	}
	moveCursor(player.row-2, player.col)
	fmt.Print(cfg.Player)

	ghostsStatusMx.RLock()
	moveCursor(len(maze)+1, 0)
	for _, g := range ghosts {
		moveCursor(g.position.row-2, g.position.col)
		if g.status == GhostStatusNormal {
			fmt.Print(cfg.Ghost)
		} else if g.status == GhostStatusRight {
			fmt.Printf(cfg.GhostRight)
		}
	}
	ghostsStatusMx.RUnlock()
	moveCursor(len(maze)+1, 0)
	livesRemaining := strconv.Itoa(lives) // 将生命 int 转换为字符串
	if cfg.UseEmoji {
		livesRemaining = getLivesAsEmoji()
	}
	fmt.Println("Score:", score, "\tLives:", livesRemaining)
}

// 初始化使用keyboard库函数进行键盘监听
func initialise() {
	err := keyboard.Open()
	if err != nil {
		panic(err)
	}
}

// 读取输入
func readInput() (string, error) {
	//GetKey()是库中的一个函数，用于读取键盘事件。该函数在读取到键盘事件时会返回字符、键码和错误。
	char, key, err := keyboard.GetKey()
	_ = char
	switch key {
	case 65517:
		return "UP", err
	case 65516:
		return "DOWN", err
	case 65515:
		return "LEFT", err
	case 65514:
		return "RIGHT", err
	case 27:
		return "ESC", err
	}
	return "", err
}

// 处理移动
func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int) {
	newRow, newCol = oldRow, oldCol
	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(maze) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(maze) {
			newRow = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(maze[0]) - 1
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(maze[0]) {
			newCol = 0
		}
	}
	if maze[newRow][newCol] == '#' {
		newRow = oldRow
		newCol = oldCol
	}
	return newRow, newCol
}

// 调整水平移动
func moveCursor(row, col int) {
	if cfg.UseEmoji {
		simpleansi.MoveCursor(row, col*2)
	} else {
		simpleansi.MoveCursor(row, col)
	}
}

// 玩家移动
func movePlayer(dir string) {
	player.row, player.col = makeMove(player.row, player.col, dir)
	removeDot := func(row, col int) {
		maze[row] = maze[row][0:col] + " " + maze[row][col+1:]
	}
	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		removeDot(player.row, player.col)
	case 'X':
		score += 10
		removeDot(player.row, player.col)
		go processPill()
	}
}

// 敌人移动
func moveGhosts() {
	for _, g := range ghosts {
		dir := drawDirection()
		g.position.row, g.position.col = makeMove(g.position.row, g.position.col, dir)
	}
}

// 随机生成器控制敌人
func drawDirection() string {
	dir := rand.Intn(4)
	move := map[int]string{
		0: "UP",
		1: "DOWN",
		2: "RIGHT",
		3: "LEFT",
	}
	return move[dir]
}

// 吃豆人吃完能量药丸，敌人状态改变方法
func processPill() {
	pillMx.Lock()
	updateGhost(ghosts, GhostStatusRight)
	if pillTimer != nil {
		pillTimer.Stop()
	}
	pillTimer = time.NewTimer(time.Second * cfg.PillDurationSecs)
	pillMx.Unlock()
	<-pillTimer.C
	pillMx.Lock()
	pillTimer.Stop()
	updateGhost(ghosts, GhostStatusNormal)
	pillMx.Unlock()
}

// 更新敌人状态
func updateGhost(ghosts []*ghost, ghostStatus GhostStatus) {
	ghostsStatusMx.Lock()
	defer ghostsStatusMx.Unlock()
	for _, g := range ghosts {
		g.status = ghostStatus
	}
}

// 根据生命连接正确数量的玩家表情符号
func getLivesAsEmoji() string {
	buf := bytes.Buffer{}
	for i := lives; i > 0; i-- {
		buf.WriteString(cfg.Player)
	}
	return buf.String()
}

func main() {
	//解析命令行参数
	flag.Parse()
	//初始化 游戏
	initialise()
	defer func() {
		// 关闭键盘监听
		keyboard.Close()
	}()
	//加载资源
	err := loadMaze(*mazeFile)
	if err != nil {
		log.Println("failed to load maze:", err)
	}
	errj := loadConfig(*configFile)
	if errj != nil {
		log.Println("failed to load configuration:", errj)
		return
	}
	//输入过程（并发）
	input := make(chan string)
	//声明一个只能接收的通道类型，并赋值为ch
	go func(ch chan<- string) {
		for {
			//输入过程
			key, err := readInput()
			if err != nil {
				log.Println("readInput are error :", err)
				input <- "ESC"
			}
			input <- key
		}
	}(input)
	//游戏循环
	for {
		//移动过程
		select {
		case inp := <-input:
			if inp == "ESC" {
				lives = 0
			}
			movePlayer(inp)
		default:

		}
		moveGhosts()

		//碰撞过程
		for _, g := range ghosts {
			if player.row == g.position.row && player.col == g.position.col {
				ghostsStatusMx.RLock()
				if g.status == GhostStatusNormal {
					lives = lives - 1
					if lives != 0 {
						moveCursor(player.row, player.col)
						fmt.Print(cfg.Death)
						moveCursor(len(maze)+2, 0)
						ghostsStatusMx.RUnlock()
						updateGhost(ghosts, GhostStatusNormal)
						time.Sleep(1000 * time.Millisecond)
						player.row, player.col = player.startRow, player.startCol
					}
				} else if g.status == GhostStatusRight {
					ghostsStatusMx.RUnlock()
					updateGhost([]*ghost{g}, GhostStatusNormal)
					g.position.row, g.position.col = g.position.startRow, g.position.startCol
				}

			}
		}
		//更新屏幕
		printScreen()
		//检测 游戏结束
		if numDots == 0 || lives <= 0 {
			if lives == 0 {
				moveCursor(player.row-4, player.col)
				fmt.Print(cfg.Death)
				moveCursor(player.startRow-3, player.startCol-1)
				fmt.Print("GAME OVER")
				moveCursor(len(maze)+2, 0)
			}
			//打破无限循环
			break
		}
		//重复过程
		time.Sleep(600 * time.Millisecond)
	}
}
