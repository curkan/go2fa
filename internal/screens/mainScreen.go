package screens

import tea "github.com/charmbracelet/bubbletea"

type rootScreenModel struct {
    model  tea.Model    // this will hold the current screen model
}

func RootScreen() rootScreenModel {
    var rootModel tea.Model

	screen_one := ListMethodsScreen()
	rootModel = &screen_one

    return rootScreenModel{
        model: rootModel,
    }
}

func (m rootScreenModel) Init() tea.Cmd {
    return m.model.Init()    // rest methods are just wrappers for the model's methods
}

func (m rootScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m.model.Update(msg)
}

func (m rootScreenModel) View() string {
    return m.model.View()
}

// this is the switcher which will switch between screens
func (m rootScreenModel) SwitchScreen(model tea.Model) (tea.Model, tea.Cmd) {
    m.model = model
    return m.model, m.model.Init()    // must return .Init() to initialize the screen (and here the magic happens)
}
