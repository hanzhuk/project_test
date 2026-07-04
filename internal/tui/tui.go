// Package tui 实现脚手架的交互式终端用户界面。
// 基于 Bubble Tea 框架，在 init 命令未提供完整参数时，
// 引导用户通过键盘选择后端框架、ORM、数据库、前端框架和功能开关。
package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/example/go-scaffold/internal/project"
)

// Choice 定义 TUI 中的单个可选项。
type Choice struct {
	Label string // 显示文本
	Value string // 内部值
}

// Theme 定义 TUI 主题样式。
type Theme struct {
	SelectedPrefix string // 选中项前缀
	Cursor         string // 光标样式
}

// TUIOptions 定义 TUI 的初始选项与主题配置。
type TUIOptions struct {
	Initial project.ProjectOptions // 初始项目选项（命令行已解析部分）
	Theme   *Theme                 // 主题样式配置
	In      io.Reader              // 输入流，默认 os.Stdin
	Out     io.Writer              // 输出流，默认 os.Stdout
}

// defaultTheme 返回默认主题样式。
func defaultTheme() *Theme {
	return &Theme{
		SelectedPrefix: "✓",
		Cursor:         ">",
	}
}

// TUI 负责启动交互式终端界面，收集用户配置。
type TUI struct {
	opts TUIOptions
}

// NewTUI 创建 TUI 实例。
func NewTUI(opts TUIOptions) *TUI {
	// 应用默认值
	if opts.Theme == nil {
		opts.Theme = defaultTheme()
	}
	if opts.In == nil {
		opts.In = os.Stdin
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	return &TUI{opts: opts}
}

// Run 启动交互式配置收集流程，返回最终项目选项。
// TUI 按顺序展示以下步骤，若命令行已提供值则跳过对应步骤：
//  1. 后端框架选择：Echo / Fiber / Gin
//  2. ORM 选择：Ent / sqlc / GORM
//  3. 数据库选择：PostgreSQL / MySQL / SQLite
//  4. 前端框架选择：React / Vue / Svelte
//  5. 功能开关：JWT 认证、Docker 配置、GitHub Actions CI/CD
func (t *TUI) Run() (project.ProjectOptions, error) {
	opts := t.opts.Initial

	// 步骤 1：后端框架选择
	if opts.Backend == "" {
		val, err := t.RunStep("请选择后端框架", []Choice{
			{Label: "Echo (默认，高性能，错误处理集中)", Value: "echo"},
			{Label: "Fiber (最高性能，Express 风格)", Value: "fiber"},
			{Label: "Gin (生态最大，市场份额高)", Value: "gin"},
		})
		if err != nil {
			return opts, err
		}
		opts.Backend = val
	}

	// 步骤 2：ORM 选择
	if opts.ORM == "" {
		val, err := t.RunStep("请选择 ORM 框架", []Choice{
			{Label: "Ent (默认，类型安全，代码生成)", Value: "ent"},
			{Label: "sqlc (类型安全，SQL 优先)", Value: "sqlc"},
			{Label: "GORM (学习曲线低，灵活)", Value: "gorm"},
		})
		if err != nil {
			return opts, err
		}
		opts.ORM = val
	}

	// 步骤 3：数据库选择
	if opts.Database == "" {
		val, err := t.RunStep("请选择数据库", []Choice{
			{Label: "PostgreSQL (默认，类型丰富，开源免费)", Value: "postgres"},
			{Label: "MySQL (生态成熟，使用广泛)", Value: "mysql"},
			{Label: "SQLite (轻量，适合开发测试)", Value: "sqlite"},
		})
		if err != nil {
			return opts, err
		}
		opts.Database = val
	}

	// 步骤 4：前端框架选择
	if opts.Frontend == "" {
		val, err := t.RunStep("请选择前端框架", []Choice{
			{Label: "React 19 + TypeScript + Vite (默认)", Value: "react"},
			{Label: "Vue 3 + TypeScript + Vite", Value: "vue"},
			{Label: "Svelte 5 + TypeScript + Vite", Value: "svelte"},
		})
		if err != nil {
			return opts, err
		}
		opts.Frontend = val
	}

	// 步骤 5：功能开关（多选）
	if !t.opts.Initial.EnableJWT && !t.opts.Initial.EnableDocker && !t.opts.Initial.EnableCI {
		features, err := t.RunMultiSelect("请选择需要启用的功能（空格选择，回车确认）", []Choice{
			{Label: "JWT 认证", Value: "jwt"},
			{Label: "Docker 配置", Value: "docker"},
			{Label: "GitHub Actions CI/CD", Value: "ci"},
		})
		if err != nil {
			return opts, err
		}
		// 解析多选结果
		for _, f := range features {
			switch f {
			case "jwt":
				opts.EnableJWT = true
			case "docker":
				opts.EnableDocker = true
			case "ci":
				opts.EnableCI = true
			}
		}
	}

	return opts, nil
}

// RunStep 启动单个配置步骤的选择界面。
// title 为步骤标题，choices 为可选项列表。
// 返回用户选中项的 Value。
func (t *TUI) RunStep(title string, choices []Choice) (string, error) {
	m := &selectModel{
		title:   title,
		choices: choices,
		cursor:  0,
		theme:   t.opts.Theme,
	}
	p := tea.NewProgram(m, tea.WithInput(t.opts.In), tea.WithOutput(t.opts.Out))
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("TUI 运行失败: %w", err)
	}
	m, ok := finalModel.(*selectModel)
	if !ok {
		return "", fmt.Errorf("TUI 返回类型错误")
	}
	if m.quit {
		return "", fmt.Errorf("用户取消操作")
	}
	if m.cursor < 0 || m.cursor >= len(choices) {
		return "", fmt.Errorf("无效的选择")
	}
	return choices[m.cursor].Value, nil
}

// RunMultiSelect 启动多选步骤界面。
// 返回用户选中的所有项的 Value 列表。
func (t *TUI) RunMultiSelect(title string, choices []Choice) ([]string, error) {
	m := &multiSelectModel{
		title:    title,
		choices:  choices,
		selected: make(map[int]bool),
		theme:    t.opts.Theme,
	}
	p := tea.NewProgram(m, tea.WithInput(t.opts.In), tea.WithOutput(t.opts.Out))
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI 运行失败: %w", err)
	}
	mm, ok := finalModel.(*multiSelectModel)
	if !ok {
		return nil, fmt.Errorf("TUI 返回类型错误")
	}
	if mm.quit {
		return nil, fmt.Errorf("用户取消操作")
	}
	// 收集选中的项
	var result []string
	for i, c := range choices {
		if mm.selected[i] {
			result = append(result, c.Value)
		}
	}
	return result, nil
}

// selectModel 是单选步骤的 Bubble Tea 模型。
type selectModel struct {
	title   string   // 步骤标题
	choices []Choice // 可选项
	cursor  int      // 当前光标位置
	theme   *Theme   // 主题
	quit    bool     // 是否退出
	done    bool     // 是否完成
}

// Init 初始化模型。
func (m *selectModel) Init() tea.Cmd {
	return nil
}

// Update 处理键盘事件。
func (m *selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			// 上移光标
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// 下移光标
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			// 确认选择
			m.done = true
			return m, tea.Quit
		case "esc", "ctrl+c":
			// 取消
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View 渲染界面。
func (m *selectModel) View() string {
	var b strings.Builder
	// 标题样式
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")
	// 渲染选项
	for i, c := range m.choices {
		// 光标与选中标记
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.Cursor + " "
		}
		// 选中项高亮
		if i == m.cursor {
			selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
			b.WriteString(selectedStyle.Render(cursor + c.Label))
		} else {
			b.WriteString(cursor + c.Label)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	// 提示信息
	hintStyle := lipgloss.NewStyle().Faint(true)
	b.WriteString(hintStyle.Render("↑/↓ 选择 · Enter 确认 · Esc 取消"))
	return b.String()
}

// multiSelectModel 是多选步骤的 Bubble Tea 模型。
type multiSelectModel struct {
	title    string        // 步骤标题
	choices  []Choice      // 可选项
	cursor   int           // 当前光标位置
	selected map[int]bool  // 选中状态
	theme    *Theme        // 主题
	quit     bool          // 是否退出
}

// Init 初始化模型。
func (m *multiSelectModel) Init() tea.Cmd {
	return nil
}

// Update 处理键盘事件。
func (m *multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case " ":
			// 空格切换选中
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			// 确认
			return m, tea.Quit
		case "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// View 渲染界面。
func (m *multiSelectModel) View() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")
	for i, c := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = m.theme.Cursor + " "
		}
		// 选中标记
		checked := " "
		if m.selected[i] {
			checked = m.theme.SelectedPrefix
		}
		line := cursor + "[" + checked + "] " + c.Label
		if i == m.cursor {
			selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
			b.WriteString(selectedStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Faint(true)
	b.WriteString(hintStyle.Render("↑/↓ 选择 · 空格 切换 · Enter 确认 · Esc 取消"))
	return b.String()
}
