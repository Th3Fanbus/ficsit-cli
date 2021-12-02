package scenes

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
	"github.com/satisfactorymodding/ficsit-cli/tea/components"
	"github.com/satisfactorymodding/ficsit-cli/tea/utils"
)

var _ tea.Model = (*modsList)(nil)

type modsList struct {
	root   components.RootModel
	list   list.Model
	parent tea.Model
	items  chan []list.Item
}

func NewMods(root components.RootModel, parent tea.Model) tea.Model {
	// TODO Color mods that are installed in current profile
	l := list.NewModel([]list.Item{}, utils.ItemDelegate{}, root.Size().Width, root.Size().Height-root.Height())
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetSpinner(spinner.MiniDot)
	l.Title = "Mods"
	l.Styles = utils.ListStyles
	l.SetSize(l.Width(), l.Height())
	l.KeyMap.Quit.SetHelp("q", "back")

	m := &modsList{
		root:   root,
		list:   l,
		parent: parent,
		items:  make(chan []list.Item),
	}

	go func() {
		items := make([]list.Item, 0)
		allMods := make([]ficsit.ModsGetModsModsMod, 0)
		offset := 0
		for {
			mods, err := ficsit.Mods(context.TODO(), root.GetAPIClient(), ficsit.ModFilter{
				Limit:    100,
				Offset:   offset,
				Order_by: ficsit.ModFieldsLastVersionDate,
				Order:    ficsit.OrderDesc,
			})

			if err != nil {
				panic(err) // TODO Handle Error
			}

			if len(mods.GetMods.Mods) == 0 {
				break
			}

			allMods = append(allMods, mods.GetMods.Mods...)

			for i := 0; i < len(mods.GetMods.Mods); i++ {
				currentOffset := offset
				currentI := i
				items = append(items, utils.SimpleItem{
					Title: mods.GetMods.Mods[i].Name,
					Activate: func(msg tea.Msg, currentModel tea.Model) (tea.Model, tea.Cmd) {
						mod := allMods[currentOffset+currentI]
						return NewModMenu(root, currentModel, utils.Mod{
							Name:      mod.Name,
							ID:        mod.Id,
							Reference: mod.Mod_reference,
						}), nil
					},
				})
			}

			offset += len(mods.GetMods.Mods)
		}

		m.items <- items
	}()

	return m
}

func (m modsList) Init() tea.Cmd {
	return utils.Ticker()
}

func (m modsList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Info().Msg(spew.Sdump(msg))
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case KeyControlC:
			return m, tea.Quit
		case "q":
			if m.parent != nil {
				m.parent.Update(m.root.Size())
				return m.parent, nil
			}
			return m, tea.Quit
		case KeyEnter:
			i, ok := m.list.SelectedItem().(utils.SimpleItem)
			if ok {
				if i.Activate != nil {
					newModel, cmd := i.Activate(msg, m)
					if newModel != nil || cmd != nil {
						if newModel == nil {
							newModel = m
						}
						return newModel, cmd
					}
					return m, nil
				}
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := lipgloss.NewStyle().Margin(m.root.Height(), 2, 0).GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
		m.root.SetSize(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case utils.TickMsg:
		select {
		case items := <-m.items:
			m.list.StopSpinner()
			cmd := m.list.SetItems(items)
			// Done to refresh keymap
			m.list.SetFilteringEnabled(m.list.FilteringEnabled())
			return m, cmd
		default:
			start := m.list.StartSpinner()
			return m, tea.Batch(utils.Ticker(), start)
		}
	}

	return m, nil
}

func (m modsList) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.root.View(), m.list.View())
}
