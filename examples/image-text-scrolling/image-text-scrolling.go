package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/bakedsnake/goberzurg"
)

const (
	crust     = "#181926"
	mantle    = "#1e2030"
	base      = "#24273a"
	surface0  = "#363a4f"
	surface1  = "#494d64"
	surface2  = "#5b6078"
	overlay0  = "#6e738d"
	overlay1  = "#8087a2"
	overlay2  = "#939ab7"
	subtext0  = "#a5adcb"
	subtext1  = "#b8c0e0"
	text      = "#cad3f5"
	lavender  = "#b7bdf8"
	blue      = "#8aadf4"
	sapphire  = "#7dc4e4"
	sky       = "#91d7e3"
	teal      = "#8bd5ca"
	green     = "#a6da95"
	yellow    = "#eed49f"
	peach     = "#f5a97f"
	maroon    = "#ee99a0"
	red       = "#ed8796"
	mauve     = "#c6a0f6"
	pink      = "#f5bde6"
	flamingo  = "#f0c6c6"
	rosewater = "#f4dbd6"
)

const (
	imgW       = 15
	imgH       = 25
	cardHeight = imgH + 4
	colImg     = 2
	colText    = colImg + imgW + 2
)

type card struct {
	imagePath string
	title     string
	body      string
}

type model struct {
	renderer  *goberzurg.Renderer
	cards     []card
	lines     []string
	width     int
	height    int
	yOff      int
	placedAt  map[int]int
	scrollSeq int
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mauve)).
			Background(lipgloss.Color(crust)).
			Bold(true)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(subtext0)).
			Background(lipgloss.Color(crust))

	initStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(text)).
			Background(lipgloss.Color(crust))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(teal)).
			Background(lipgloss.Color(crust))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(overlay1)).
			Background(lipgloss.Color(crust))

	sepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(overlay0)).
			Background(lipgloss.Color(crust)).
			Italic(true)
)

func (m *model) buildContent(w int) {
	var b strings.Builder
	seps := []string{
		"━━━ ⋆⋅☆⋅⋆ ━━━    these are my favorite cards",
		"━━━ ⋆⋅☆⋅⋆ ━━━    collected over many years",
		"━━━ ⋆⋅☆⋅⋆ ━━━    each one has a story",
		"━━━ ⋆⋅☆⋅⋆ ━━━    some rare, some common",
		"━━━ ⋆⋅☆⋅⋆ ━━━    all precious to me",
		"━━━ ⋆⋅☆⋅⋆ ━━━    the collection grows",
		"━━━ ⋆⋅☆⋅⋆ ━━━    and there's always room for more",
	}
	pad := strings.Repeat(" ", colText)
	for i, c := range m.cards {
		b.WriteString(titleStyle.Width(w).Render(pad + "  " + c.title + "  "))
		b.WriteByte('\n')

		bodyLines := strings.Split(c.body, "\n")
		for row := 0; row < imgH; row++ {
			var line string
			if row < len(bodyLines) {
				line = bodyLines[row]
			}
			b.WriteString(bodyStyle.Width(w).Render(pad + line))
			b.WriteByte('\n')
		}

		b.WriteByte('\n')
		if i < len(seps) {
			b.WriteString(sepStyle.Width(w).Render(pad + seps[i]))
			b.WriteByte('\n')
			b.WriteByte('\n')
		}
	}
	m.lines = strings.Split(b.String(), "\n")
	if len(m.lines) > 0 && m.lines[len(m.lines)-1] == "" {
		m.lines = m.lines[:len(m.lines)-1]
	}
}

func (m model) Init() tea.Cmd { return nil }

type refreshImagesMsg struct {
	seq int
}

func debounceImages(seq int) tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return refreshImagesMsg{seq: seq}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.buildContent(msg.Width)
		m.renderer.Clear()
		m.placedAt = make(map[int]int)
		m.placeImages()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		prevOff := m.yOff
		switch msg.String() {
		case "up", "k":
			m.yOff = max(0, m.yOff-1)
		case "down", "j":
			maxOff := max(0, len(m.lines)-m.height+1)
			m.yOff = min(maxOff, m.yOff+1)
		case "pgup":
			m.yOff = max(0, m.yOff-(m.height-1))
		case "pgdown":
			maxOff := max(0, len(m.lines)-m.height+1)
			m.yOff = min(maxOff, m.yOff+(m.height-1))
		case "home":
			m.yOff = 0
		case "end":
			maxOff := max(0, len(m.lines)-m.height+1)
			m.yOff = maxOff
		}

		if m.yOff == prevOff {
			return m, nil
		}
		m.renderer.Clear()
		m.placedAt = make(map[int]int)
		m.scrollSeq++
		return m, debounceImages(m.scrollSeq)
	case tea.MouseMsg:
		prevOff := m.yOff
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.yOff = max(0, m.yOff-3)
		case tea.MouseButtonWheelDown:
			maxOff := max(0, len(m.lines)-m.height+1)
			m.yOff = min(maxOff, m.yOff+3)
		}
		if m.yOff == prevOff {
			return m, nil
		}
		m.renderer.Clear()
		m.placedAt = make(map[int]int)
		m.scrollSeq++
		return m, debounceImages(m.scrollSeq)
	case refreshImagesMsg:
		if msg.seq != m.scrollSeq {
			return m, nil
		}
		m.renderer.Clear()
		m.placedAt = make(map[int]int)
		m.placeImages()
	}
	return m, nil
}

func (m *model) placeImages() {
	for i := range m.cards {
		imgLine := i*cardHeight + 1
		newTop := imgLine - m.yOff + 1

		if newTop+imgH <= 1 || newTop > m.height {
			continue
		}

		m.renderer.Display(m.cards[i].imagePath,
			goberzurg.WithPos(colImg, newTop),
			goberzurg.WithSize(imgW, imgH),
		)
		m.placedAt[i] = newTop
	}
}

func (m model) View() string {
	if len(m.lines) == 0 {
		return initStyle.Render("loading...")
	}

	header := headerStyle.Width(m.width).Render("  cards  ↑/↓ scroll  · q quit  ") + "\n"

	top := min(m.yOff, len(m.lines))
	bottom := min(top+m.height-1, len(m.lines))
	visible := m.lines[top:bottom]

	return header + strings.Join(visible, "\n")
}

func main() {
	r := goberzurg.New()
	defer r.Close()

	cards := []card{
		{
			imagePath: "yugi-card.jpg",
			title:     "Dark Magician",
			body:      "The ultimate wizard in terms of attack and defense.\nATK: 2500 / DEF: 2000\nA spellcaster who commands the forces of darkness.\nHis signature spell is Dark Magic Attack.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Blue-Eyes White Dragon",
			body:      "This legendary dragon is a powerful engine of destruction.\nATK: 3000 / DEF: 2500\nVirtually invincible, very few have faced\nthis awesome creature and lived to tell the tale.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Red-Eyes Black Dragon",
			body:      "A ferocious dragon with a deadly attack.\nATK: 2400 / DEF: 2000\nThis jet-black dragon is a fearsome sight to behold.\nIt channels flames hotter than the sun itself.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Exodia the Forbidden One",
			body:      "If you gather all five pieces, you win the duel instantly.\nLeft Leg / Right Leg / Left Arm / Right Arm\nEach piece sealed in a different dimension.\nOnce united, no force can stand against it.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Dark Magician Girl",
			body:      "Gains 300 ATK for every Dark Magician in the graveyard.\nATK: 2000 / DEF: 1700\nThe apprentice of the Dark Magician.\nHer magical power grows with her confidence.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Slifer the Sky Dragon",
			body:      "One of the three Egyptian God Cards.\nIts power is infinite.\nATK: ? / DEF: ?\nIts attack equals the number of cards in your hand times 1000.\nThe sky itself rends when this beast appears.",
		},
		{
			imagePath: "yugi-card.jpg",
			title:     "Obelisk the Tormentor",
			body:      "Another Egyptian God Card.\nFist of fate descends upon thee.\nATK: 4000 / DEF: 4000\nRequires three tributes to summon.\nIts power is beyond mortal comprehension.",
		},
	}

	m := model{
		renderer: r,
		cards:    cards,
		placedAt: make(map[int]int),
	}

	tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
}
