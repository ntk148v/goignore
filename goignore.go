// Copyright 2021 Kien Nguyen-Tuan <kiennt2609@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#BD2109", Dark: "#BD2109"}).
				Render
)

type item struct {
	title       string
	description string
	path        string
}

func (i item) Title() string       { return i.title }
func (i item) Path() string        { return i.path }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

type listKeyMap struct {
	updateTemplates key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		updateTemplates: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "update gitignore templates"),
		),
	}
}

type model struct {
	list         list.Model
	spinner      spinner.Model
	keys         *listKeyMap
	delegateKeys *delegateKeyMap
}

func newModel() model {
	var (
		delegateKeys = newDelegateKeyMap()
		listKeys     = newListKeyMap()
	)

	// Get gitignore repository
	items, err := updateTemplateList()
	if err != nil {
		panic(err)
	}

	// Setup list
	delegate := newItemDelegate(delegateKeys)
	ignoreList := list.New(items, delegate, 0, 0)
	ignoreList.Title = "Ignore templates"
	ignoreList.Styles.Title = titleStyle
	ignoreList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.updateTemplates,
		}
	}
	ignoreList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.updateTemplates,
		}
	}

	// Setup spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	return model{
		list:         ignoreList,
		spinner:      s,
		keys:         listKeys,
		delegateKeys: delegateKeys,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.list.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		// Don't match any of the keys below if we're actively filtering.
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.updateTemplates):
			cmd := m.list.NewStatusMessage(m.spinner.View() + statusMessageStyle("Pull newest templates"))
			if err := pullTemplateUpdates(); err != nil {
				cmd = m.list.NewStatusMessage(errorMessageStyle(err.Error()))
			}
			return m, tea.Batch(cmd)
		}
	}

	// This will also call our delegate's update function.
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(m.list.View())
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if err := initTemplates(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
