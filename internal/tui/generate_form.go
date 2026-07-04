package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/example/go-scaffold/internal/ai"
)

// GenerateFormOptions 定义 generate 表单的初始选项。
type GenerateFormOptions struct {
	InitialDescription string // 命令行预填充的描述
}

// GenerateFormResult 是 generate 表单的返回结果。
type GenerateFormResult struct {
	Input  *ai.GenerateInput // 生成输入
	IsDryRun bool             // 是否预览模式
	Cancel  bool              // 是否取消
}

// generateFormModel 是 generate 代码生成表单的 Bubble Tea 模型。
// 包含描述、类型、实体、字段四个输入区域和预览/生成/取消三个按钮。
type generateFormModel struct {
	description string // 描述输入内容
	codeTypeIdx int    // 当前选中的代码类型索引
	entity      string // 实体名称输入内容
	fields      string // 字段输入内容（name:type 逗号分隔）

	focusIdx int  // 当前焦点区域索引（0=描述,1=类型,2=实体,3=字段,4=预览,5=生成,6=取消）
	quit     bool // 是否退出
	result   *GenerateFormResult // 返回结果

	codeTypes []string // 代码类型选项
}

// codeTypeOptions 是代码类型的可选项，对应接口文档 2.3.3。
var codeTypeOptions = []string{"auto", "handler", "model", "service", "route", "test"}

// RunGenerateForm 启动 generate 代码生成表单界面。
// initialDescription 为命令行预填充的描述（可为空）。
// 返回用户填写的结果。
func RunGenerateForm(initialDescription string) (*GenerateFormResult, error) {
	m := &generateFormModel{
		description: initialDescription,
		codeTypeIdx: 0,
		codeTypes:   codeTypeOptions,
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI 表单运行失败: %w", err)
	}
	m, ok := finalModel.(*generateFormModel)
	if !ok {
		return nil, fmt.Errorf("TUI 返回类型错误")
	}
	if m.quit || m.result == nil {
		return &GenerateFormResult{Cancel: true}, nil
	}
	return m.result, nil
}

// Init 初始化模型。
func (m *generateFormModel) Init() tea.Cmd {
	return nil
}

// Update 处理键盘事件。
func (m *generateFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "esc":
			// 在输入区域 Esc 取消，在按钮区域 Esc 也取消
			m.quit = true
			return m, tea.Quit
		case "tab":
			// Tab 切换焦点到下一个区域
			m.focusIdx = (m.focusIdx + 1) % 7
		case "shift+tab":
			// Shift+Tab 切换到上一个区域
			m.focusIdx = (m.focusIdx - 1 + 7) % 7
		case "up":
			// 在类型选择区域上键切换选项
			if m.focusIdx == 1 {
				m.codeTypeIdx = (m.codeTypeIdx - 1 + len(m.codeTypes)) % len(m.codeTypes)
			}
		case "down":
			// 在类型选择区域下键切换选项
			if m.focusIdx == 1 {
				m.codeTypeIdx = (m.codeTypeIdx + 1) % len(m.codeTypes)
			}
		case "left":
			// 在类型选择区域左键切换
			if m.focusIdx == 1 {
				m.codeTypeIdx = (m.codeTypeIdx - 1 + len(m.codeTypes)) % len(m.codeTypes)
			}
		case "right":
			// 在类型选择区域右键切换
			if m.focusIdx == 1 {
				m.codeTypeIdx = (m.codeTypeIdx + 1) % len(m.codeTypes)
			}
		case "enter":
			// Enter 根据当前焦点执行操作
			switch m.focusIdx {
			case 4: // 预览按钮
				m.result = m.buildResult(true)
				return m, tea.Quit
			case 5: // 生成按钮
				m.result = m.buildResult(false)
				return m, tea.Quit
			case 6: // 取消按钮
				m.quit = true
				return m, tea.Quit
			default:
				// 在输入区域 Enter 跳到下一个区域
				m.focusIdx = (m.focusIdx + 1) % 7
			}
		case "backspace":
			// 退格删除输入
			switch m.focusIdx {
			case 0:
				if len(m.description) > 0 {
					m.description = m.description[:len(m.description)-1]
				}
			case 2:
				if len(m.entity) > 0 {
					m.entity = m.entity[:len(m.entity)-1]
				}
			case 3:
				if len(m.fields) > 0 {
					m.fields = m.fields[:len(m.fields)-1]
				}
			}
		default:
			// 处理普通字符输入
			if len(msg.String()) == 1 {
				ch := msg.String()[0]
				// 仅接受可打印字符
				if ch >= 32 && ch < 127 {
					switch m.focusIdx {
					case 0:
						m.description += string(ch)
					case 2:
						m.entity += string(ch)
					case 3:
						m.fields += string(ch)
					}
				}
			}
		}
	}
	return m, nil
}

// buildResult 根据表单输入构建返回结果。
func (m *generateFormModel) buildResult(isDryRun bool) *GenerateFormResult {
	input := &ai.GenerateInput{
		Description: m.description,
		CodeType:    m.codeTypes[m.codeTypeIdx],
		Entity:      m.entity,
		Fields:      parseFields(m.fields),
	}
	return &GenerateFormResult{
		Input:     input,
		IsDryRun:  isDryRun,
		Cancel:    false,
	}
}

// parseFields 解析字段输入字符串为 Field 列表。
// 格式为 name:type 逗号分隔，如 id:int,name:string。
func parseFields(fieldsStr string) []ai.Field {
	if fieldsStr == "" {
		return nil
	}
	var fields []ai.Field
	parts := strings.Split(fieldsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, ":", 2)
		if len(kv) == 2 {
			fields = append(fields, ai.Field{
				Name: strings.TrimSpace(kv[0]),
				Type: strings.TrimSpace(kv[1]),
			})
		}
	}
	return fields
}

// View 渲染表单界面。
func (m *generateFormModel) View() string {
	var b strings.Builder

	// 标题
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).PaddingBottom(1)
	b.WriteString(titleStyle.Render("AI 代码生成"))
	b.WriteString("\n\n")

	// 描述输入框
	b.WriteString(m.renderLabel("描述", 0))
	b.WriteString("\n")
	b.WriteString(m.renderInputBox(m.description, 0, 50))
	b.WriteString("\n\n")

	// 类型下拉选择
	b.WriteString(m.renderLabel("类型", 1))
	b.WriteString("\n")
	b.WriteString(m.renderTypeSelector())
	b.WriteString("\n\n")

	// 实体输入框
	b.WriteString(m.renderLabel("实体", 2))
	b.WriteString("\n")
	b.WriteString(m.renderInputBox(m.entity, 2, 50))
	b.WriteString("\n\n")

	// 字段输入框
	b.WriteString(m.renderLabel("字段", 3))
	b.WriteString("\n")
	b.WriteString(m.renderInputBox(m.fields, 3, 50))
	b.WriteString("\n")

	// 字段预览
	if preview := m.renderFieldsPreview(); preview != "" {
		b.WriteString("\n")
		b.WriteString(preview)
	}
	b.WriteString("\n\n")

	// 按钮组
	b.WriteString(m.renderButtons())
	b.WriteString("\n\n")

	// 提示
	hintStyle := lipgloss.NewStyle().Faint(true)
	b.WriteString(hintStyle.Render("Tab 切换焦点 · ↑/↓ 选择类型 · Enter 确认 · Esc 取消"))
	return b.String()
}

// renderLabel 渲染字段标签，焦点区域高亮。
func (m *generateFormModel) renderLabel(label string, idx int) string {
	if m.focusIdx == idx {
		style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
		return style.Render("▸ " + label + ":")
	}
	return "  " + label + ":"
}

// renderInputBox 渲染输入框，焦点区域显示光标。
func (m *generateFormModel) renderInputBox(value string, idx, width int) string {
	// 补全显示宽度
	display := value
	if len(display) < width-2 {
		display = display + strings.Repeat(" ", width-2-len(display))
	} else if len(display) > width-2 {
		display = display[:width-2]
	}
	// 焦点区域显示光标
	if m.focusIdx == idx {
		boxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		return boxStyle.Render("[" + display + "_]")
	}
	return "[" + display + " ]"
}

// renderTypeSelector 渲染类型下拉选择器。
func (m *generateFormModel) renderTypeSelector() string {
	current := m.codeTypes[m.codeTypeIdx]
	if m.focusIdx == 1 {
		// 焦点时显示所有选项
		var opts []string
		for i, t := range m.codeTypes {
			if i == m.codeTypeIdx {
				opts = append(opts, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("["+t+" ▼]"))
			} else {
				opts = append(opts, " "+t+"  ")
			}
		}
		return strings.Join(opts, " ")
	}
	return "[" + current + " ▼]"
}

// renderFieldsPreview 渲染字段解析预览。
func (m *generateFormModel) renderFieldsPreview() string {
	fields := parseFields(m.fields)
	if len(fields) == 0 {
		return ""
	}
	var b strings.Builder
	previewStyle := lipgloss.NewStyle().Faint(true)
	b.WriteString(previewStyle.Render("  字段预览:"))
	for _, f := range fields {
		b.WriteString(previewStyle.Render(fmt.Sprintf("\n    %s → %s", f.Name, mapGoType(f.Type))))
	}
	return b.String()
}

// renderButtons 渲染底部按钮组。
func (m *generateFormModel) renderButtons() string {
	buttons := []struct {
		label string
		idx   int
	}{
		{"[预览]", 4},
		{"[生成]", 5},
		{"[取消]", 6},
	}
	var parts []string
	for _, btn := range buttons {
		if m.focusIdx == btn.idx {
			style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Background(lipgloss.Color("238"))
			parts = append(parts, style.Render(" "+btn.label+" "))
		} else {
			parts = append(parts, " "+btn.label+" ")
		}
	}
	return strings.Join(parts, "  ")
}

// mapGoType 将简写类型映射为 Go 类型，用于预览显示。
func mapGoType(short string) string {
	switch short {
	case "int":
		return "int"
	case "int64":
		return "int64"
	case "string":
		return "string"
	case "bool":
		return "bool"
	case "time":
		return "time.Time"
	case "float":
		return "float64"
	default:
		return short
	}
}
