/* TODO:
- validatie verbeteren (wachtwoord regels, verplichte velden)
- error handling bij file opslaan
- meer hints/feedback tijdens het invullen (bv. groen vinkje bij valid veld)
- voorbeeldwaarden in de velden (placeholder)
- optie om terug te gaan en aanpassingen te maken na bevestiging
- betere scheiding tussen form logica en UI (MVVM patroon)
- extra tabblad met overzicht van alle instellingen voor bevestiging
- mogelijkheid om alleen Terraform of Ansible bestanden te genereren
*/

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// wordt overschreven via ldflags bij build
var version = "v1.0.0"

// min terminal size
const (
	minWidth  = 80
	minHeight = 30
)

// mainframe style
var (
	green    = lipgloss.Color("#00FF41")
	amber    = lipgloss.Color("#FFB000")
	dim      = lipgloss.Color("#33FF57")
	bgCol    = lipgloss.Color("#0a0a0a")
	borderCl = lipgloss.Color("#00CC00")
	redCl    = lipgloss.Color("#FF0000")

	tabNames = []string{"AZURE", "WACHTWOORDEN", "WORDPRESS", "DATABASE", "COMPONENTEN", "LUANTI", "SSH & OPTIES"}
)

// vars

type TerraformVars struct {
	SubscriptionID    string `json:"subscription_id"`
	ResourceGroupName string `json:"resource_group_name"`
	PublicIPDNSLabel  string `json:"public_ip_dns_label"`
	LuantiDNSLabel    string `json:"luanti_dns_label"`
}

type AnsibleVars struct {
	MysqlRootPassword     string `json:"mysql_root_password"`
	DBWpPassword          string `json:"db_wp_password"`
	WpAdminPassword       string `json:"wp_admin_password"`
	AnsibleBecomePassword string `json:"ansible_become_password"`

	WpDBName string `json:"wp_db_name"`
	WpDBUser string `json:"wp_db_user"`

	WpAdminUser    string `json:"wp_admin_user"`
	WpAdminEmail   string `json:"wp_admin_email"`
	WpTitle        string `json:"wp_title"`
	WpLocale       string `json:"wp_locale"`
	CertbotStg     bool   `json:"certbot_staging"`
	ExternalDomain string `json:"external_domain,omitempty"`

	EnableVaultwarden     bool   `json:"enable_vaultwarden"`
	VaultwardenAdminToken string `json:"vaultwarden_admin_token,omitempty"`
	EnableTechSnake       bool   `json:"enable_tech_snake"`

	LuantiServerName   string `json:"luanti_server_name,omitempty"`
	LuantiSshHostAlias string `json:"luanti_ssh_host_alias,omitempty"`

	SSHHostAlias string `json:"ssh_host_alias"`
	SSHKey       string `json:"ssh_key"`
}

// file handling

func findRoot() string {
	dir, _ := os.Getwd()
	for {
		// zoek naar pyproject.toml — uniek voor de projectroot
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			cwd, _ := os.Getwd()
			return filepath.Dir(cwd)
		}
		dir = parent
	}
}

func loadJSON[T any](path string, fallback string) T {
	var result T
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = os.ReadFile(fallback)
		if err != nil {
			return result
		}
	}
	_ = json.Unmarshal(data, &result)
	return result
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0600)
}

// ASCII art (past in 76 kolommen)

func asciiLogo() string {
	return ` ██████╗ ██████╗ ██████╗ ██████╗  █████╗  ██████╗██╗  ██╗████████╗███████╗
██╔═══██╗██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██║  ██║╚══██╔══╝██╔════╝
██║   ██║██████╔╝██║  ██║██████╔╝███████║██║     ███████║   ██║   ███████╗
██║   ██║██╔═══╝ ██║  ██║██╔══██╗██╔══██║██║     ██╔══██║   ██║   ╚════██║
╚██████╔╝██║     ██████╔╝██║  ██║██║  ██║╚██████╗██║  ██║   ██║   ███████║
 ╚═════╝ ╚═╝     ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝`
}

func subtitle() string {
	return fmt.Sprintf("C O N F I G U R A T I E   G E N E R A T O R   %s\n           Groep 99  ─  SELab Opdracht 5", version)
}

// huh theme

func mainframeTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Title = t.Focused.Title.Foreground(amber).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(dim)
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(green)
	t.Focused.TextInput.Text = t.Focused.TextInput.Text.Foreground(green)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(borderCl)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(amber)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(green)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(dim)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(bgCol).Background(green).Bold(true)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(dim).Background(bgCol)
	t.Focused.Base = t.Focused.Base.BorderForeground(borderCl)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(amber).Bold(true)

	t.Blurred = t.Focused
	t.Blurred.TextInput.Text = t.Blurred.TextInput.Text.Foreground(dim)
	t.Blurred.Title = t.Blurred.Title.Foreground(dim)
	t.Blurred.Base = t.Blurred.Base.BorderForeground(lipgloss.Color("#008800"))

	return t
}

// bubbletea model

type appState int

const (
	stateResizeWait appState = iota
	stateForm
	stateDone
)

type model struct {
	state       appState
	width       int
	height      int
	form        *huh.Form
	root        string
	tf          *TerraformVars
	ans         *AnsibleVars
	confirmSave *bool
	result      string
	quitting    bool
}

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.state == stateResizeWait && m.width >= minWidth && m.height >= minHeight {
			m.state = stateForm
		}
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))) && m.state == stateDone:
			m.quitting = true
			return m, tea.Quit
		}
	}

	if m.state == stateForm {
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}

		if m.form.State == huh.StateCompleted {
			if *m.confirmSave {
				m.saveFiles()
			} else {
				m.result = "  Geannuleerd - er zijn geen bestanden weggeschreven."
			}
			m.state = stateDone
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

func (m *model) saveFiles() {
	tfPath := filepath.Join(m.root, "terraform.tfvars.json")
	ansPath := filepath.Join(m.root, "ansible_vars.json")

	var lines []string

	if err := writeJSON(tfPath, m.tf); err != nil {
		lines = append(lines, fmt.Sprintf("  ✗ %s: %s", tfPath, err.Error()))
	} else {
		lines = append(lines, fmt.Sprintf("  ✓ %s", tfPath))
	}

	if err := writeJSON(ansPath, m.ans); err != nil {
		lines = append(lines, fmt.Sprintf("  \u2717 %s: %s", ansPath, err.Error()))
	} else {
		lines = append(lines, fmt.Sprintf("  \u2713 %s", ansPath))
	}

	lines = append(lines, "")
	lines = append(lines, "  Klaar! Start deployment met:")
	lines = append(lines, "")
	lines = append(lines, "    make all")

	m.result = strings.Join(lines, "\n")
}

// views

func (m model) View() string {
	if m.quitting {
		return ""
	}

	w := m.width
	h := m.height
	if w == 0 || h == 0 {
		return ""
	}

	if w < minWidth || h < minHeight {
		return m.viewResizeWarning()
	}

	switch m.state {
	case stateForm:
		return m.viewForm()
	case stateDone:
		return m.viewDone()
	default:
		return m.viewResizeWarning()
	}
}

func (m model) viewResizeWarning() string {
	w := m.width
	h := m.height

	icon := lipgloss.NewStyle().
		Foreground(redCl).
		Bold(true).
		Render("▓▓▓  ⚠  ▓▓▓")

	title := lipgloss.NewStyle().
		Foreground(amber).
		Bold(true).
		Render("TERMINAL TE KLEIN")

	current := lipgloss.NewStyle().Foreground(redCl).Bold(true).
		Render(fmt.Sprintf("%d × %d", w, h))
	required := lipgloss.NewStyle().Foreground(green).Bold(true).
		Render(fmt.Sprintf("%d × %d", minWidth, minHeight))

	sizeInfo := fmt.Sprintf("Huidig : %s\nVereist: %s", current, required)

	hint := lipgloss.NewStyle().Foreground(dim).
		Render("Vergroot je terminal venster...")

	content := lipgloss.JoinVertical(lipgloss.Center,
		icon,
		"",
		title,
		"",
		sizeInfo,
		"",
		hint,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(redCl).
		Padding(2, 6).
		Align(lipgloss.Center).
		Render(content)

	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func (m model) renderHeader(innerW int) string {
	logo := lipgloss.NewStyle().
		Foreground(green).
		Bold(true).
		Align(lipgloss.Center).
		Width(innerW).
		Render(asciiLogo())

	sub := lipgloss.NewStyle().
		Foreground(green).
		Align(lipgloss.Center).
		Width(innerW).
		Render(subtitle())

	return lipgloss.JoinVertical(lipgloss.Center, logo, sub)
}

func (m model) renderSep(innerW int) string {
	return lipgloss.NewStyle().
		Foreground(borderCl).
		Render(strings.Repeat("─", innerW))
}

// onderrand met versie links en hint-tekst in het midden
func buildBottomBorder(width int, left string, hint string) string {
	db := lipgloss.DoubleBorder()
	borderStyle := lipgloss.NewStyle().Foreground(borderCl)

	leftWidth := lipgloss.Width(left)
	hintWidth := lipgloss.Width(hint)
	usedWidth := leftWidth + hintWidth
	totalBorderChars := width - usedWidth
	if totalBorderChars < 2 {
		return borderStyle.Render(db.BottomLeft + strings.Repeat(db.Bottom, width) + db.BottomRight)
	}

	gapLeft := totalBorderChars / 2
	gapRight := totalBorderChars - gapLeft

	return borderStyle.Render(db.BottomLeft) +
		left +
		borderStyle.Render(strings.Repeat(db.Bottom, gapLeft)) +
		hint +
		borderStyle.Render(strings.Repeat(db.Bottom, gapRight)+db.BottomRight)
}

// bepaalt de actieve groep-index adhv de form-output
func (m model) currentGroup() int {
	v := m.form.View()
	// Zoek de groep-titels die als Note in elke groep staan.
	markers := []string{
		"AZURE / TERRAFORM",
		"WACHTWOORDEN",
		"WORDPRESS",
		"DATABASE",
		"OPTIONELE COMPONENTEN",
		"LUANTI / VOXELIBRE",
		"SSH & OPTIES",
	}
	for i, mk := range markers {
		if strings.Contains(v, mk) {
			return i
		}
	}
	return 0
}

// een horizontale tab-balk met de actieve sectie
func (m model) renderTabBar(innerW int) string {
	active := m.currentGroup()

	var tabs []string
	for i, name := range tabNames {
		var style lipgloss.Style
		if i == active {
			style = lipgloss.NewStyle().
				Foreground(bgCol).
				Background(amber).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(dim).
				Background(lipgloss.Color("#0a0a0a")).
				Padding(0, 1)
		}
		tabs = append(tabs, style.Render(name))
	}

	bar := strings.Join(tabs, lipgloss.NewStyle().Foreground(borderCl).Render(" │ "))

	return lipgloss.NewStyle().
		Width(innerW).
		Align(lipgloss.Center).
		Render(bar)
}

func (m model) viewForm() string {
	w := m.width
	h := m.height
	innerW := w - 4

	header := m.renderHeader(innerW)
	sep := m.renderSep(innerW)

	tfPath := filepath.Join(m.root, "terraform.tfvars.json")
	ansPath := filepath.Join(m.root, "ansible_vars.json")
	filesInfo := lipgloss.NewStyle().
		Foreground(dim).
		Padding(0, 1).
		Render(fmt.Sprintf("▌ Bron: %s  │  %s", tfPath, ansPath))

	tabBar := m.renderTabBar(innerW)

	formView := lipgloss.NewStyle().
		Width(innerW).
		Render(m.form.View())

	content := lipgloss.JoinVertical(lipgloss.Left,
		header,
		sep,
		filesInfo,
		sep,
		tabBar,
		sep,
		formView,
	)

	frame := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderCl).
		BorderBottom(false).
		Width(w-2).
		Height(h-3).
		Padding(0, 1)

	hint := lipgloss.NewStyle().
		Foreground(green).
		Bold(true).
		Render(" tab/enter ▸ volgende  │  shift+tab ◂ vorige  │  ctrl+c ✕ stop ")

	versionLabel := lipgloss.NewStyle().
		Foreground(dim).
		Render(" " + version + " ")

	bottomBorder := buildBottomBorder(w-2, versionLabel, hint)

	return frame.Render(content) + "\n" + bottomBorder
}

func (m model) viewDone() string {
	w := m.width
	h := m.height
	innerW := w - 4

	header := m.renderHeader(innerW)
	sep := m.renderSep(innerW)

	resultBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(green).
		Foreground(green).
		Bold(true).
		Padding(1, 2).
		Width(innerW - 4).
		Render(m.result)

	content := lipgloss.JoinVertical(lipgloss.Center,
		header,
		sep,
		"",
		lipgloss.NewStyle().Foreground(amber).Bold(true).Padding(0, 1).Render("▌ BESTANDEN OPGESLAGEN"),
		"",
		resultBox,
		"",
	)

	frame := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderCl).
		BorderBottom(false).
		Width(w-2).
		Height(h-3).
		Padding(0, 1)

	hint := lipgloss.NewStyle().
		Foreground(green).
		Bold(true).
		Render(" q / esc ✕ sluiten ")

	versionLabel := lipgloss.NewStyle().
		Foreground(dim).
		Render(" " + version + " ")

	bottomBorder := buildBottomBorder(w-2, versionLabel, hint)

	return frame.Render(content) + "\n" + bottomBorder
}

// MAIN garbage

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("config-starter %s\n", version)
		os.Exit(0)
	}

	root := findRoot()

	tfPath := filepath.Join(root, "terraform.tfvars.json")
	tfExample := filepath.Join(root, "terraform.tfvars.json.example")
	ansPath := filepath.Join(root, "ansible_vars.json")
	ansExample := filepath.Join(root, "ansible_vars.json.example")

	tf := loadJSON[TerraformVars](tfPath, tfExample)
	ans := loadJSON[AnsibleVars](ansPath, ansExample)

	// defaults
	if ans.WpDBName == "" {
		ans.WpDBName = "wordpress"
	}
	if ans.WpLocale == "" {
		ans.WpLocale = "nl_BE"
	}
	if ans.SSHKey == "" {
		ans.SSHKey = "~/.ssh/id_ed25519_hogent"
	}
	if ans.SSHHostAlias == "" {
		ans.SSHHostAlias = "<jouw-initialen>-wordpressapp"
	}
	if ans.VaultwardenAdminToken == "" {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		ans.VaultwardenAdminToken = hex.EncodeToString(b)
	}
	if ans.LuantiSshHostAlias == "" {
		ans.LuantiSshHostAlias = "luanti"
	}
	if ans.LuantiServerName == "" {
		ans.LuantiServerName = "VoxeLibre Server - Groep 99"
	}

	if tf.ResourceGroupName == "" {
		tf.ResourceGroupName = "SELab-Wordpress"
	}

	confirmSave := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("█ AZURE / TERRAFORM").
				Description("Infrastructuur instellingen voor Azure provisioning."),
			huh.NewInput().
				Title("Subscription ID").
				Description("Azure abonnements-ID").
				Value(&tf.SubscriptionID).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("verplicht veld")
					}
					return nil
				}),
			huh.NewInput().
				Title("Resource Group").
				Description("Naam van de Azure resourcegroep").
				Value(&tf.ResourceGroupName),
			huh.NewInput().
				Title("DNS Label (Docker host)").
				Description("Publiek IP DNS label -> <label>.swedencentral.cloudapp.azure.com").
				Value(&tf.PublicIPDNSLabel),
			huh.NewInput().
				Title("DNS Label (Luanti VM)").
				Description("DNS label voor de Luanti/VoxeLibre VM").
				Value(&tf.LuantiDNSLabel),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ WACHTWOORDEN").
				Description("Database en WordPress admin wachtwoorden."),
			huh.NewInput().
				Title("MariaDB Root Wachtwoord").
				Description("Root wachtwoord voor de MariaDB container").
				Value(&ans.MysqlRootPassword).
				EchoMode(huh.EchoModePassword),
			huh.NewInput().
				Title("DB WordPress Wachtwoord").
				Description("WordPress applicatie DB gebruiker").
				Value(&ans.DBWpPassword).
				EchoMode(huh.EchoModePassword),
			huh.NewInput().
				Title("WordPress Admin Wachtwoord").
				Description("WordPress admin paneel").
				Value(&ans.WpAdminPassword).
				EchoMode(huh.EchoModePassword),
			huh.NewInput().
				Title("Ansible Become Wachtwoord").
				Description("sudo wachtwoord op de VM").
				Value(&ans.AnsibleBecomePassword).
				EchoMode(huh.EchoModePassword),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ WORDPRESS").
				Description("Site-instellingen en admin account.\nDomein wordt automatisch ingesteld via Azure FQDN (DNS label)."),
			huh.NewInput().
				Title("Admin Gebruiker").
				Value(&ans.WpAdminUser),
			huh.NewInput().
				Title("Admin E-mail").
				Value(&ans.WpAdminEmail),
			huh.NewInput().
				Title("Site Titel").
				Value(&ans.WpTitle),
			huh.NewInput().
				Title("Locale").
				Description("bv. nl_BE, nl_NL, en_US").
				Value(&ans.WpLocale),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ DATABASE").
				Description("MariaDB draait als Docker container.\nHost is automatisch localhost."),
			huh.NewInput().
				Title("Database Naam").
				Value(&ans.WpDBName),
			huh.NewInput().
				Title("WordPress DB Gebruiker").
				Value(&ans.WpDBUser),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ OPTIONELE COMPONENTEN").
				Description("Extra self-hosted tools die naast WordPress draaien.\nPortainer wordt altijd geïnstalleerd."),
			huh.NewConfirm().
				Title("Vaultwarden").
				Description("wachtwoordkluis (docker container /secrets)").
				Value(&ans.EnableVaultwarden),
			huh.NewInput().
				Title("Vaultwarden Admin Token").
				Description("token voor /secrets/admin panel (automatisch gegenereerd, leeg = admin uit)").
				Value(&ans.VaultwardenAdminToken),
			huh.NewConfirm().
				Title("Tech Snake").
				Description("snake game (WASM op /snake)").
				Value(&ans.EnableTechSnake),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ LUANTI / VOXELIBRE").
				Description("Minetest/Luanti dedicated server met VoxeLibre game.\nDraait op een aparte ARM64 VM."),
			huh.NewInput().
				Title("Server Naam").
				Description("Naam die spelers zien in de serverlijst").
				Value(&ans.LuantiServerName),
			huh.NewInput().
				Title("Luanti SSH Host Alias").
				Description("Naam in ~/.ssh/config voor de Luanti VM").
				Value(&ans.LuantiSshHostAlias),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("█ SSH & OPTIES").
				Description("SSH configuratie en deployment opties."),
			huh.NewInput().
				Title("SSH Host Alias (Docker host)").
				Description("Naam in ~/.ssh/config").
				Value(&ans.SSHHostAlias),
			huh.NewInput().
				Title("SSH Sleutel").
				Description("Pad naar privé-sleutel").
				Value(&ans.SSHKey),
			huh.NewConfirm().
				Title("Certbot Staging").
				Description("Staging server (hogere rate limits, ongeldig cert)?").
				Value(&ans.CertbotStg),
			huh.NewInput().
				Title("Extern Domein").
				Description("Basisdomain (bijv. groep99.be) -- wordt {dns-label}.groep99.be").
				Value(&ans.ExternalDomain),
			huh.NewConfirm().
				Title("Configuratie opslaan?").
				Description("Bestanden worden aangemaakt in de projectroot.").
				Affirmative("Opslaan").
				Negative("Annuleren").
				Value(&confirmSave),
		),
	).WithTheme(mainframeTheme())

	m := model{
		state:       stateResizeWait,
		form:        form,
		root:        root,
		tf:          &tf,
		ans:         &ans,
		confirmSave: &confirmSave,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fout: %v\n", err)
		os.Exit(1)
	}
}
