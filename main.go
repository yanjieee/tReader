package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type App struct {
	app              *tview.Application
	mainView         *tview.TextView
	searchInput      *tview.InputField
	layout           *tview.Flex
	readingMode      bool
	searchMode       bool
	currentNovelLine int
	novelContent     []string
	fakeBuffer       []string
	novelFile        string
	opacity          int
}

// ============================================================================
// 应用初始化
// ============================================================================

func NewApp(novelFile string) *App {
	app := &App{
		app:              tview.NewApplication(),
		readingMode:      false,
		searchMode:       false,
		currentNovelLine: 0,
		novelFile:        novelFile,
		fakeBuffer:       make([]string, 0, 25),
		opacity:          3,
	}
	app.loadNovelContent()
	app.createSearchInput()
	return app
}

// ============================================================================
// 文件加载
// ============================================================================

func (a *App) loadNovelContent() {
	filename := "novel.txt"
	if a.novelFile != "" {
		filename = a.novelFile
	}
	a.loadTextContent(filename)
}

func (a *App) tryGBKDecode(data []byte) (string, bool) {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", false
	}
	text := string(decoded)

	// 如果GBK解码后没有替换字符，认为解码成功
	if !strings.Contains(text, "\uFFFD") {
		return text, true
	}

	return "", false
}

func (a *App) loadTextContent(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		a.novelContent = []string{
			"未找到小说文件: " + filename,
			"",
			"支持格式: UTF-8 或 GBK 编码的 .txt 文件",
			"使用方法: ./tReader [文件路径]",
			"",
			"快捷键：",
			"- h: 切换阅读模式（老板键）",
			"- j/k 或 ↑↓: 滚动小说内容",
			"- [/]: 调节透明度",
			"- /: 搜索文本",
			"- Ctrl+C: 退出程序",
		}
		return
	}

	var text string

	// 检查是否为有效的UTF-8
	if utf8.Valid(data) {
		text = string(data)
	} else {
		// 不是有效UTF-8，尝试GBK解码
		if gbkText, ok := a.tryGBKDecode(data); ok {
			text = gbkText
		} else {
			// GBK也失败，使用UTF-8强制解码（会有替换字符）
			text = string(data)
		}
	}

	// 处理换行符并分割为行
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")

	var content []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			// 将长行分割成多行显示
			splitLines := splitLongLine(line, 70)
			content = append(content, splitLines...)
		}
	}

	if len(content) == 0 {
		a.novelContent = []string{"文件为空或编码不支持"}
		return
	}

	a.novelContent = content
}

// ============================================================================
// 搜索功能
// ============================================================================

func (a *App) createSearchInput() {
	a.searchInput = tview.NewInputField().
		SetLabel("搜索: ").
		SetPlaceholder("输入要搜索的文字...").
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite)
	
	a.searchInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			searchText := a.searchInput.GetText()
			if searchText != "" {
				a.searchNovel(searchText)
			}
		}
		// Enter 或 Escape 都会隐藏搜索框
		a.hideSearchInput()
	})
}

func (a *App) showSearchInput() {
	a.searchMode = true
	a.layout.Clear()
	a.layout.AddItem(a.mainView, 0, 1, false)
	a.layout.AddItem(a.searchInput, 1, 0, true)
	a.searchInput.SetText("")
	a.app.SetFocus(a.searchInput)
}

func (a *App) hideSearchInput() {
	a.searchMode = false
	a.layout.Clear()
	a.layout.AddItem(a.mainView, 0, 1, true)
	a.app.SetFocus(a.mainView)
	a.updateDisplay()
}

func (a *App) searchNovel(searchText string) {
	searchText = strings.ToLower(searchText)
	
	for i, line := range a.novelContent {
		if strings.Contains(strings.ToLower(line), searchText) {
			a.currentNovelLine = i
			// 确保不超出范围
			maxLine := len(a.novelContent) - 15
			if maxLine < 0 {
				maxLine = 0
			}
			if a.currentNovelLine > maxLine {
				a.currentNovelLine = maxLine
			}
			return
		}
	}
}

// ============================================================================
// 字符串处理工具
// ============================================================================

func safeSubstring(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	// 从maxLen位置向前找到一个有效的UTF-8字符边界
	for i := maxLen; i > 0; i-- {
		if utf8.ValidString(text[:i]) {
			return text[:i]
		}
	}
	return text[:maxLen]
}

// 将长行分割成多行，每行最多maxLen字符
func splitLongLine(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var lines []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= maxLen {
			lines = append(lines, remaining)
			break
		}

		// 找到安全的UTF-8字符边界
		cutPoint := maxLen
		for i := maxLen; i > 0; i-- {
			if utf8.ValidString(remaining[:i]) {
				cutPoint = i
				break
			}
		}

		lines = append(lines, remaining[:cutPoint])
		remaining = remaining[cutPoint:]
	}

	return lines
}

// ============================================================================
// 颜色和透明度
// ============================================================================

func (a *App) getNovelColor() string {
	colors := []string{
		"#101010", "#202020", "#303030", "#404040", "#505050",
		"#606060", "#707070", "#808080", "#909090", "#a0a0a0",
	}
	return colors[a.opacity]
}

// ============================================================================
// 伪装界面生成
// ============================================================================

func (a *App) addFakeLine() {
	fakeMessages := []string{
		"├─ [INFO] Processing microservice requests...",
		"├─ [DEBUG] Database connection pool: %d/20 active",
		"├─ [WARN] Memory usage: %d.%d%%",
		"├─ [INFO] Redis cache hit ratio: %d.%d%%",
		"├─ [DEBUG] Executing SQL: SELECT * FROM user_sessions",
		"├─ [INFO] API Gateway response time: %dms",
		"├─ [WARN] Queue depth: %d messages pending",
		"├─ [DEBUG] JWT token validation successful",
		"├─ [INFO] Elasticsearch index updated: %d documents",
		"├─ [DEBUG] Kafka consumer lag: %d messages",
		"├─ [INFO] Container health check passed: app-server-%d",
		"├─ [WARN] CPU threshold exceeded: %d%% on node-3",
		"├─ [DEBUG] Load balancer distributing to %d backend servers",
		"├─ [INFO] Backup completed: %d.%d GB transferred",
		"├─ [DEBUG] WebSocket connections: %d active",
		"├─ [INFO] CDN cache refresh initiated for region: us-east-1",
		"├─ [WARN] Disk I/O latency: %dms (threshold: 100ms)",
		"├─ [DEBUG] Auto-scaling triggered: launching %d new instances",
		"├─ [INFO] SSL certificate renewal scheduled",
		"├─ [DEBUG] Circuit breaker status: CLOSED",
	}

	msg := fakeMessages[rand.Intn(len(fakeMessages))]

	var fakeLine string
	switch strings.Count(msg, "%d") {
	case 1:
		fakeLine = fmt.Sprintf(msg, rand.Intn(100)+1)
	case 2:
		fakeLine = fmt.Sprintf(msg, rand.Intn(90)+10, rand.Intn(10))
	default:
		fakeLine = msg
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fullLine := fmt.Sprintf("[gray]%s[white] %s", timestamp, fakeLine)

	a.fakeBuffer = append(a.fakeBuffer, fullLine)
	if len(a.fakeBuffer) > 25 {
		a.fakeBuffer = a.fakeBuffer[1:]
	}
}

// ============================================================================
// UI界面管理
// ============================================================================

func (a *App) createMainView() *tview.TextView {
	a.mainView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(false).
		SetChangedFunc(func() {
			a.app.Draw()
		})

	a.mainView.SetBorder(true).SetTitle(" [green]Production System Monitor v3.2.1[white] ").SetTitleAlign(tview.AlignLeft)

	go func() {
		for {
			a.addFakeLine()
			a.updateDisplay()
			time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
		}
	}()

	return a.mainView
}

func (a *App) createLayout() *tview.Flex {
	a.layout = tview.NewFlex().SetDirection(tview.FlexRow)
	
	// 主视图
	mainView := a.createMainView()
	a.layout.AddItem(mainView, 0, 1, true)
	
	return a.layout
}

func (a *App) updateDisplay() {
	a.mainView.Clear()

	if !a.readingMode {
		for _, line := range a.fakeBuffer {
			fmt.Fprintln(a.mainView, line)
		}
	} else {
		fakeIndex := 0
		novelIndex := a.currentNovelLine
		linesDisplayed := 0
		maxLines := 22

		for linesDisplayed < maxLines && fakeIndex < len(a.fakeBuffer) {
			if linesDisplayed%2 == 0 && novelIndex < len(a.novelContent) {
				fakeLine := a.fakeBuffer[min(fakeIndex, len(a.fakeBuffer)-1)]
				novelLine := a.novelContent[novelIndex]
				fmt.Fprintf(a.mainView, "%s\n[%s]│  %s[white]\n", fakeLine, a.getNovelColor(), novelLine)
				novelIndex++
				linesDisplayed += 2
			} else {
				if fakeIndex < len(a.fakeBuffer) {
					fmt.Fprintln(a.mainView, a.fakeBuffer[fakeIndex])
				}
				linesDisplayed++
			}
			fakeIndex++
		}

		for linesDisplayed < maxLines {
			if linesDisplayed%2 == 0 && novelIndex < len(a.novelContent) {
				novelLine := a.novelContent[novelIndex]
				fmt.Fprintf(a.mainView, "[gray]%s ├─ [INFO] Processing background tasks...[white]\n", time.Now().Format("15:04:05"))
				fmt.Fprintf(a.mainView, "[%s]│  %s[white]\n", a.getNovelColor(), novelLine)
				novelIndex++
				linesDisplayed += 2
			} else {
				fmt.Fprintf(a.mainView, "[gray]%s ├─ [DEBUG] System status: OK[white]\n", time.Now().Format("15:04:05"))
				linesDisplayed++
			}
		}

		percentage := float64(a.currentNovelLine) / float64(len(a.novelContent)) * 100
		fmt.Fprintf(a.mainView, "\n[%s]━━ 阅读模式 (%.1f%%) | j/k ↑↓ 滚动 | h 切换 | [/] 透明度 | / 搜索 ━━[white]", a.getNovelColor(), percentage)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// 键盘控制
// ============================================================================

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 如果在搜索模式，让搜索框处理输入
		if a.searchMode {
			return event
		}
		
		switch {
		case event.Key() == tcell.KeyRune && event.Rune() == '/':
			a.showSearchInput()
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == 'h':
			a.readingMode = !a.readingMode
			a.updateDisplay()
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == '[':
			if a.opacity > 0 {
				a.opacity--
				a.updateDisplay()
			}
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == ']':
			if a.opacity < 9 {
				a.opacity++
				a.updateDisplay()
			}
			return nil
		case a.readingMode && (event.Key() == tcell.KeyRune && event.Rune() == 'k'):
			if a.currentNovelLine > 0 {
				a.currentNovelLine--
				a.updateDisplay()
			}
			return nil
		case a.readingMode && (event.Key() == tcell.KeyRune && event.Rune() == 'j'):
			maxLine := len(a.novelContent) - 15
			if maxLine < 0 {
				maxLine = 0
			}
			if a.currentNovelLine < maxLine {
				a.currentNovelLine++
				a.updateDisplay()
			}
			return nil
		case a.readingMode && event.Key() == tcell.KeyUp:
			if a.currentNovelLine > 0 {
				a.currentNovelLine--
				a.updateDisplay()
			}
			return nil
		case a.readingMode && event.Key() == tcell.KeyDown:
			maxLine := len(a.novelContent) - 15
			if maxLine < 0 {
				maxLine = 0
			}
			if a.currentNovelLine < maxLine {
				a.currentNovelLine++
				a.updateDisplay()
			}
			return nil
		case a.readingMode && event.Key() == tcell.KeyPgUp:
			a.currentNovelLine -= 5
			if a.currentNovelLine < 0 {
				a.currentNovelLine = 0
			}
			a.updateDisplay()
			return nil
		case a.readingMode && event.Key() == tcell.KeyPgDn:
			maxLine := len(a.novelContent) - 15
			if maxLine < 0 {
				maxLine = 0
			}
			a.currentNovelLine += 5
			if a.currentNovelLine > maxLine {
				a.currentNovelLine = maxLine
			}
			a.updateDisplay()
			return nil
		case event.Key() == tcell.KeyCtrlC:
			a.app.Stop()
			return nil
		}
		return event
	})
}

// ============================================================================
// 应用运行
// ============================================================================

func (a *App) Run() error {
	layout := a.createLayout()
	a.setupKeyBindings()
	return a.app.SetRoot(layout, true).SetFocus(a.mainView).Run()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var novelFile string
	if len(os.Args) > 1 {
		novelFile = os.Args[1]
	}

	app := NewApp(novelFile)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
