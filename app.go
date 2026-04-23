package main

import (
	"fmt"
	"strings"
)

type PageID int

const (
	PageMain PageID = iota
	PageStatus
	PageCommit
	PageBranch
	PageLog
	PageStash
	PageDiff
	PageRemote
	PageSettings
)

type DialogType int

const (
	DialogNone DialogType = iota
	DialogConfirm
	DialogInput
	DialogInfo
	DialogMenu
)

type DialogState struct {
	Type         DialogType
	Title        string
	Message      string
	Buttons      []string
	Selected     int
	InputValue   string
	InputCursor  int
	OnConfirm    func()
	OnInput      func(string)
	MenuItems    []MenuItem
	MenuSelected int
	OnMenuSelect func(int)
}

type AppState struct {
	CurrentPage PageID
	PageStack   []PageID

	MenuSelected int
	MenuScroll   int

	StatusSelected int
	StatusScroll   int
	StatusTab      int

	CommitSelected    int
	CommitScroll      int
	CommitTab         int
	CommitMessage     string
	CommitMsgCursor   int
	CommitInputActive bool

	BranchSelected int
	BranchScroll   int
	BranchTab      int

	LogSelected     int
	LogScroll       int
	LogDetailScroll int
	LogShowDetail   bool
	LogSearchMode   bool
	LogSearchQuery  string
	LogSearchCursor int

	StashSelected int
	StashScroll   int

	DiffSelected int
	DiffScroll   int
	DiffTab      int

	RemoteSelected int
	RemoteScroll   int

	SettingsSelected int

	Dialog DialogState

	Notification string
	NotifType    int
	NotifTicks   int

	ThemeIdx int
}

type App struct {
	term  *Terminal
	ui    *UI
	git   *Git
	state AppState
	alive bool
}

func NewApp(t *Terminal, g *Git) *App {
	themes := []Theme{CatppuccinTheme, DarkTheme}
	return &App{
		term:  t,
		ui:    NewUI(t, themes[0]),
		git:   g,
		alive: true,
		state: AppState{
			CurrentPage: PageMain,
			ThemeIdx:    0,
		},
	}
}

func (a *App) Run() error {
	a.render()
	for a.alive {
		key := a.term.ReadKey()
		a.handleKey(key)
		if a.alive {
			a.render()
		}
	}
	return nil
}

func (a *App) Stop() {
	a.alive = false
}

func (a *App) pushPage(page PageID) {
	a.state.PageStack = append(a.state.PageStack, a.state.CurrentPage)
	a.state.CurrentPage = page
	a.resetPageState(page)
}

func (a *App) popPage() {
	if len(a.state.PageStack) > 0 {
		a.state.CurrentPage = a.state.PageStack[len(a.state.PageStack)-1]
		a.state.PageStack = a.state.PageStack[:len(a.state.PageStack)-1]
	} else {
		a.alive = false
	}
}

func (a *App) resetPageState(page PageID) {
	switch page {
	case PageStatus:
		a.state.StatusSelected = 0
		a.state.StatusScroll = 0
		a.state.StatusTab = 0
	case PageCommit:
		a.state.CommitSelected = 0
		a.state.CommitScroll = 0
		a.state.CommitTab = 0
		a.state.CommitMessage = ""
		a.state.CommitMsgCursor = 0
		a.state.CommitInputActive = false
	case PageBranch:
		a.state.BranchSelected = 0
		a.state.BranchScroll = 0
		a.state.BranchTab = 0
	case PageLog:
		a.state.LogSelected = 0
		a.state.LogScroll = 0
		a.state.LogDetailScroll = 0
		a.state.LogShowDetail = false
		a.state.LogSearchMode = false
		a.state.LogSearchQuery = ""
	case PageStash:
		a.state.StashSelected = 0
		a.state.StashScroll = 0
	case PageDiff:
		a.state.DiffSelected = 0
		a.state.DiffScroll = 0
		a.state.DiffTab = 0
	case PageRemote:
		a.state.RemoteSelected = 0
		a.state.RemoteScroll = 0
	case PageSettings:
		a.state.SettingsSelected = 0
	}
}

func (a *App) notify(msg string, notifType int) {
	a.state.Notification = msg
	a.state.NotifType = notifType
	a.state.NotifTicks = 8
}

func (a *App) showDialog(dt DialogType, title, message string, buttons []string) {
	a.state.Dialog = DialogState{
		Type:     dt,
		Title:    title,
		Message:  message,
		Buttons:  buttons,
		Selected: 0,
	}
}

func (a *App) showConfirm(title, message string, onConfirm func()) {
	a.state.Dialog = DialogState{
		Type:      DialogConfirm,
		Title:     title,
		Message:   message,
		Buttons:   []string{L("dialog.yes"), L("dialog.no")},
		Selected:  1,
		OnConfirm: onConfirm,
	}
}

func (a *App) showInput(title, prompt string, onInput func(string)) {
	a.state.Dialog = DialogState{
		Type:        DialogInput,
		Title:       title,
		Message:     prompt,
		InputValue:  "",
		InputCursor: 0,
		OnInput:     onInput,
	}
}

func (a *App) showInfo(title, message string) {
	a.state.Dialog = DialogState{
		Type:     DialogInfo,
		Title:    title,
		Message:  message,
		Buttons:  []string{L("dialog.ok")},
		Selected: 0,
	}
}

func (a *App) showMenuDialog(title string, items []MenuItem, onSelect func(int)) {
	a.state.Dialog = DialogState{
		Type:         DialogMenu,
		Title:        title,
		MenuItems:    items,
		MenuSelected: 0,
		OnMenuSelect: onSelect,
	}
}

func (a *App) closeDialog() {
	a.state.Dialog = DialogState{Type: DialogNone}
}

func (a *App) handleKey(key Key) {
	if a.state.NotifTicks > 0 {
		a.state.NotifTicks--
		if a.state.NotifTicks == 0 {
			a.state.Notification = ""
		}
	}
	if a.state.Dialog.Type != DialogNone {
		a.handleDialogKey(key)
		return
	}
	switch a.state.CurrentPage {
	case PageMain:
		a.handleMainKey(key)
	case PageStatus:
		a.handleStatusKey(key)
	case PageCommit:
		a.handleCommitKey(key)
	case PageBranch:
		a.handleBranchKey(key)
	case PageLog:
		a.handleLogKey(key)
	case PageStash:
		a.handleStashKey(key)
	case PageDiff:
		a.handleDiffKey(key)
	case PageRemote:
		a.handleRemoteKey(key)
	case PageSettings:
		a.handleSettingsKey(key)
	}
}

func (a *App) handleDialogKey(key Key) {
	d := &a.state.Dialog
	switch d.Type {
	case DialogConfirm:
		switch key.Type {
		case KeyLeft:
			if d.Selected > 0 {
				d.Selected--
			}
		case KeyRight:
			if d.Selected < len(d.Buttons)-1 {
				d.Selected++
			}
		case KeyEnter:
			if d.Selected == 0 && d.OnConfirm != nil {
				fn := d.OnConfirm
				a.closeDialog()
				fn()
			} else {
				a.closeDialog()
			}
		case KeyEscape:
			a.closeDialog()
		case KeyRune:
			if key.Rune == 'y' || key.Rune == 'Y' {
				if d.OnConfirm != nil {
					fn := d.OnConfirm
					a.closeDialog()
					fn()
				}
			} else if key.Rune == 'n' || key.Rune == 'N' {
				a.closeDialog()
			}
		}
	case DialogInput:
		switch key.Type {
		case KeyEscape:
			a.closeDialog()
		case KeyEnter:
			if d.OnInput != nil {
				fn := d.OnInput
				val := d.InputValue
				a.closeDialog()
				fn(val)
			}
		case KeyBackspace:
			if d.InputCursor > 0 {
				runes := []rune(d.InputValue)
				d.InputValue = string(runes[:d.InputCursor-1]) + string(runes[d.InputCursor:])
				d.InputCursor--
			}
		case KeyDelete:
			runes := []rune(d.InputValue)
			if d.InputCursor < len(runes) {
				d.InputValue = string(runes[:d.InputCursor]) + string(runes[d.InputCursor+1:])
			}
		case KeyLeft:
			if d.InputCursor > 0 {
				d.InputCursor--
			}
		case KeyRight:
			if d.InputCursor < len([]rune(d.InputValue)) {
				d.InputCursor++
			}
		case KeyHome, KeyCtrlA:
			d.InputCursor = 0
		case KeyEnd, KeyCtrlE:
			d.InputCursor = len([]rune(d.InputValue))
		case KeyCtrlU:
			runes := []rune(d.InputValue)
			d.InputValue = string(runes[d.InputCursor:])
			d.InputCursor = 0
		case KeyCtrlK:
			runes := []rune(d.InputValue)
			d.InputValue = string(runes[:d.InputCursor])
		case KeyCtrlW:
			runes := []rune(d.InputValue)
			if d.InputCursor > 0 {
				start := d.InputCursor - 1
				for start > 0 && runes[start] == ' ' {
					start--
				}
				for start > 0 && runes[start-1] != ' ' {
					start--
				}
				d.InputValue = string(runes[:start]) + string(runes[d.InputCursor:])
				d.InputCursor = start
			}
		case KeyRune:
			runes := []rune(d.InputValue)
			newRunes := make([]rune, 0, len(runes)+1)
			newRunes = append(newRunes, runes[:d.InputCursor]...)
			newRunes = append(newRunes, key.Rune)
			newRunes = append(newRunes, runes[d.InputCursor:]...)
			d.InputValue = string(newRunes)
			d.InputCursor++
		}
	case DialogInfo:
		switch key.Type {
		case KeyEnter, KeyEscape:
			a.closeDialog()
		}
	case DialogMenu:
		switch key.Type {
		case KeyUp:
			if d.MenuSelected > 0 {
				d.MenuSelected--
			}
		case KeyDown:
			if d.MenuSelected < len(d.MenuItems)-1 {
				d.MenuSelected++
			}
		case KeyEnter:
			if d.OnMenuSelect != nil {
				fn := d.OnMenuSelect
				idx := d.MenuSelected
				a.closeDialog()
				fn(idx)
			}
		case KeyEscape:
			a.closeDialog()
		case KeyRune:
			if key.Rune >= '1' && key.Rune <= '9' {
				idx := int(key.Rune - '1')
				if idx < len(d.MenuItems) && d.OnMenuSelect != nil {
					fn := d.OnMenuSelect
					a.closeDialog()
					fn(idx)
				}
			}
		}
	}
}

func (a *App) handleMainKey(key Key) {
	menuItems := a.getMainMenuItems()
	switch key.Type {
	case KeyUp:
		if a.state.MenuSelected > 0 {
			a.state.MenuSelected--
		}
	case KeyDown:
		if a.state.MenuSelected < len(menuItems)-1 {
			a.state.MenuSelected++
		}
	case KeyEnter:
		a.activateMainItem(a.state.MenuSelected)
	case KeyEscape:
		a.showConfirm(L("app.quit"), L("app.quit.confirm"), func() {
			a.alive = false
		})
	case KeyRune:
		switch key.Rune {
		case 'q':
			a.showConfirm(L("app.quit"), L("app.quit.confirm"), func() {
				a.alive = false
			})
		case '1':
			a.activateMainItem(0)
		case '2':
			a.activateMainItem(1)
		case '3':
			a.activateMainItem(2)
		case '4':
			a.activateMainItem(3)
		case '5':
			a.activateMainItem(4)
		case '6':
			a.activateMainItem(5)
		case '7':
			a.activateMainItem(6)
		case '8':
			a.activateMainItem(7)
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) getMainMenuItems() []MenuItem {
	return []MenuItem{
		{Label: L("menu.status"), Hotkey: "1"},
		{Label: L("menu.commit"), Hotkey: "2"},
		{Label: L("menu.branch"), Hotkey: "3"},
		{Label: L("menu.log"), Hotkey: "4"},
		{Label: L("menu.stash"), Hotkey: "5"},
		{Label: L("menu.diff"), Hotkey: "6"},
		{Label: L("menu.remote"), Hotkey: "7"},
		{Label: L("menu.settings"), Hotkey: "8"},
	}
}

func (a *App) activateMainItem(idx int) {
	items := a.getMainMenuItems()
	if idx >= len(items) {
		return
	}
	switch idx {
	case 0:
		a.pushPage(PageStatus)
	case 1:
		a.pushPage(PageCommit)
	case 2:
		a.pushPage(PageBranch)
	case 3:
		a.pushPage(PageLog)
	case 4:
		a.pushPage(PageStash)
	case 5:
		a.pushPage(PageDiff)
	case 6:
		a.pushPage(PageRemote)
	case 7:
		a.pushPage(PageSettings)
	}
}

func (a *App) handleStatusKey(key Key) {
	switch key.Type {
	case KeyEscape:
		a.popPage()
		return
	case KeyUp:
		if a.state.StatusSelected > 0 {
			a.state.StatusSelected--
		}
	case KeyDown:
		a.state.StatusSelected++
	case KeyTab:
		a.state.StatusTab = (a.state.StatusTab + 1) % 3
		a.state.StatusSelected = 0
		a.state.StatusScroll = 0
	case KeyEnter:
		a.handleStatusAction()
	case KeyRune:
		switch key.Rune {
		case 'q', 'h':
			a.popPage()
			return
		case 's':
			a.state.StatusTab = 0
			a.state.StatusSelected = 0
		case 'u':
			a.state.StatusTab = 1
			a.state.StatusSelected = 0
		case 'n':
			a.state.StatusTab = 2
			a.state.StatusSelected = 0
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleStatusAction() {
	staged, unstaged, untracked, _ := a.git.Status()
	var files []FileStatus
	switch a.state.StatusTab {
	case 0:
		files = staged
	case 1:
		files = unstaged
	case 2:
		files = untracked
	}
	if a.state.StatusSelected >= len(files) {
		return
	}
	file := files[a.state.StatusSelected]

	items := []MenuItem{}
	if a.state.StatusTab == 0 {
		items = append(items,
			MenuItem{Label: L("action.unstage"), Hotkey: "u"},
			MenuItem{Label: L("action.view.diff"), Hotkey: "d"},
		)
	} else {
		items = append(items,
			MenuItem{Label: L("action.stage"), Hotkey: "s"},
			MenuItem{Label: L("action.view.diff"), Hotkey: "d"},
		)
		if a.state.StatusTab == 2 {
			items = append(items,
				MenuItem{Label: L("action.gitignore"), Hotkey: "i"},
			)
		}
	}

	a.showMenuDialog(file.Path, items, func(idx int) {
		switch idx {
		case 0:
			if a.state.StatusTab == 0 {
				err := a.git.Unstage(file.Path)
				if err != nil {
					a.notify(L("error.git")+": "+err.Error(), 2)
				} else {
					a.notify("✅ "+L("action.unstage")+" "+file.Path, 0)
				}
			} else {
				err := a.git.Stage(file.Path)
				if err != nil {
					a.notify(L("error.git")+": "+err.Error(), 2)
				} else {
					a.notify("✅ "+L("action.stage")+" "+file.Path, 0)
				}
			}
		case 1:
			a.pushPage(PageDiff)
		case 2:
			err := a.git.AddToGitignore(file.Path)
			if err != nil {
				a.showInfo(L("error.git"), err.Error())
			} else {
				a.notify("✅ "+L("action.gitignore")+" "+file.Path, 0)
			}
		}
	})
}

func (a *App) handleCommitKey(key Key) {
	if a.state.CommitInputActive {
		switch key.Type {
		case KeyEscape:
			a.state.CommitInputActive = false
		case KeyEnter:
			a.state.CommitInputActive = false
			a.doCommit()
		case KeyBackspace:
			if a.state.CommitMsgCursor > 0 {
				runes := []rune(a.state.CommitMessage)
				a.state.CommitMessage = string(runes[:a.state.CommitMsgCursor-1]) + string(runes[a.state.CommitMsgCursor:])
				a.state.CommitMsgCursor--
			}
		case KeyLeft:
			if a.state.CommitMsgCursor > 0 {
				a.state.CommitMsgCursor--
			}
		case KeyRight:
			if a.state.CommitMsgCursor < len([]rune(a.state.CommitMessage)) {
				a.state.CommitMsgCursor++
			}
		case KeyHome, KeyCtrlA:
			a.state.CommitMsgCursor = 0
		case KeyEnd, KeyCtrlE:
			a.state.CommitMsgCursor = len([]rune(a.state.CommitMessage))
		case KeyRune:
			runes := []rune(a.state.CommitMessage)
			newRunes := make([]rune, 0, len(runes)+1)
			newRunes = append(newRunes, runes[:a.state.CommitMsgCursor]...)
			newRunes = append(newRunes, key.Rune)
			newRunes = append(newRunes, runes[a.state.CommitMsgCursor:]...)
			a.state.CommitMessage = string(newRunes)
			a.state.CommitMsgCursor++
		}
		return
	}

	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.CommitSelected > 0 {
			a.state.CommitSelected--
		}
	case KeyDown:
		a.state.CommitSelected++
	case KeyTab:
		a.state.CommitTab = (a.state.CommitTab + 1) % 3
		a.state.CommitSelected = 0
		a.state.CommitScroll = 0
	case KeyEnter:
		a.handleCommitAction()
	case KeyRune:
		switch key.Rune {
		case 'e':
			a.state.CommitInputActive = true
		case 'c':
			a.doCommit()
		case 'A':
			a.doAmendCommit()
		case 'a':
			a.git.StageAll()
			a.notify("✅ "+L("commit.stage.all"), 0)
		case 'u':
			a.git.UnstageAll()
			a.notify("✅ "+L("commit.unstage.all"), 0)
		case 's':
			a.handleCommitStage()
		case 'q':
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleCommitStage() {
	staged, unstaged, untracked, _ := a.git.Status()
	var files []FileStatus
	switch a.state.CommitTab {
	case 0:
		files = staged
	case 1:
		files = unstaged
	case 2:
		files = untracked
	}
	if a.state.CommitSelected >= len(files) {
		return
	}
	file := files[a.state.CommitSelected]
	if a.state.CommitTab == 0 {
		a.git.Unstage(file.Path)
		a.notify("✅ "+L("action.unstage")+" "+file.Path, 0)
	} else {
		a.git.Stage(file.Path)
		a.notify("✅ "+L("action.stage")+" "+file.Path, 0)
	}
}

func (a *App) doCommit() {
	msg := strings.TrimSpace(a.state.CommitMessage)
	if msg == "" {
		a.showInfo(L("commit.empty"), L("commit.empty"))
		return
	}
	a.showConfirm(L("commit.confirm"), msg, func() {
		err := a.git.Commit(msg)
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.state.CommitMessage = ""
			a.state.CommitMsgCursor = 0
			a.notify(L("commit.success"), 0)
		}
	})
}

func (a *App) doAmendCommit() {
	msg := strings.TrimSpace(a.state.CommitMessage)
	a.showConfirm(L("commit.amend.confirm"), L("commit.amend.confirm"), func() {
		var err error
		if msg != "" {
			err = a.git.AmendCommit(msg)
		} else {
			err = a.git.AmendCommit("")
		}
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.state.CommitMessage = ""
			a.state.CommitMsgCursor = 0
			a.notify("✅ "+L("commit.amend")+" "+L("commit.success"), 0)
		}
	})
}

func (a *App) handleCommitAction() {
	staged, unstaged, untracked, _ := a.git.Status()
	var files []FileStatus
	switch a.state.CommitTab {
	case 0:
		files = staged
	case 1:
		files = unstaged
	case 2:
		files = untracked
	}
	if a.state.CommitSelected >= len(files) {
		return
	}
	file := files[a.state.CommitSelected]

	items := []MenuItem{}
	if a.state.CommitTab == 0 {
		items = append(items,
			MenuItem{Label: L("action.unstage"), Hotkey: "u"},
			MenuItem{Label: L("action.view.diff"), Hotkey: "d"},
		)
	} else {
		items = append(items,
			MenuItem{Label: L("action.stage"), Hotkey: "s"},
			MenuItem{Label: L("action.view.diff"), Hotkey: "d"},
		)
	}

	a.showMenuDialog(file.Path, items, func(idx int) {
		switch idx {
		case 0:
			if a.state.CommitTab == 0 {
				a.git.Unstage(file.Path)
				a.notify("✅ "+L("action.unstage")+" "+file.Path, 0)
			} else {
				a.git.Stage(file.Path)
				a.notify("✅ "+L("action.stage")+" "+file.Path, 0)
			}
		case 1:
			a.pushPage(PageDiff)
		}
	})
}

func (a *App) handleBranchKey(key Key) {
	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.BranchSelected > 0 {
			a.state.BranchSelected--
		}
	case KeyDown:
		branches, _ := a.git.Branches()
		if a.state.BranchSelected < len(branches)-1 {
			a.state.BranchSelected++
		}
	case KeyTab:
		a.state.BranchTab = (a.state.BranchTab + 1) % 2
		a.state.BranchSelected = 0
		a.state.BranchScroll = 0
	case KeyEnter:
		a.handleBranchAction()
	case KeyRune:
		switch key.Rune {
		case 'n':
			a.showInput(L("branch.create"), L("branch.create.prompt"), func(name string) {
				name = strings.TrimSpace(name)
				if name == "" {
					return
				}
				err := a.git.CreateBranch(name)
				if err != nil {
					a.showInfo(L("error.git"), err.Error())
				} else {
					a.notify(Lf("branch.success.create", name), 0)
				}
			})
		case 's':
			a.doBranchSwitch()
		case 'd':
			a.doBranchDelete()
		case 'm':
			a.doBranchMerge()
		case 'r':
			a.doBranchRebase()
		case 'q':
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleBranchAction() {
	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		} else if a.state.BranchTab == 1 && b.IsRemote {
			filtered = append(filtered, b)
		}
	}
	if a.state.BranchSelected >= len(filtered) {
		return
	}
	branch := filtered[a.state.BranchSelected]
	if branch.Current {
		return
	}

	items := []MenuItem{
		{Label: L("branch.switch"), Hotkey: "s"},
		{Label: L("branch.merge"), Hotkey: "m"},
	}
	if !branch.IsRemote {
		items = append(items,
			MenuItem{Label: L("branch.delete"), Hotkey: "d"},
			MenuItem{Label: L("branch.rebase"), Hotkey: "r"},
		)
	}

	a.showMenuDialog(branch.Name, items, func(idx int) {
		switch idx {
		case 0:
			a.doBranchSwitchTo(branch.Name)
		case 1:
			a.doBranchMergeTo(branch.Name)
		case 2:
			a.doBranchDeleteName(branch.Name)
		case 3:
			a.doBranchRebaseTo(branch.Name)
		}
	})
}

func (a *App) doBranchSwitch() {
	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		} else if a.state.BranchTab == 1 && b.IsRemote {
			filtered = append(filtered, b)
		}
	}
	if a.state.BranchSelected >= len(filtered) {
		return
	}
	branch := filtered[a.state.BranchSelected]
	a.doBranchSwitchTo(branch.Name)
}

func (a *App) doBranchSwitchTo(name string) {
	a.showConfirm(L("branch.switch.confirm"), Lf("branch.switch.confirm", name), func() {
		err := a.git.SwitchBranch(name)
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.notify(Lf("branch.success.switch", name), 0)
		}
	})
}

func (a *App) doBranchDelete() {
	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		}
	}
	if a.state.BranchSelected >= len(filtered) {
		return
	}
	branch := filtered[a.state.BranchSelected]
	a.doBranchDeleteName(branch.Name)
}

func (a *App) doBranchDeleteName(name string) {
	a.showConfirm(L("branch.delete.confirm"), Lf("branch.delete.confirm", name), func() {
		err := a.git.DeleteBranch(name, false)
		if err != nil {
			a.showConfirm(L("branch.delete.force.confirm"), Lf("branch.delete.force.confirm", name), func() {
				err2 := a.git.DeleteBranch(name, true)
				if err2 != nil {
					a.showInfo(L("error.git"), err2.Error())
				} else {
					a.notify(Lf("branch.success.delete", name), 0)
				}
			})
		} else {
			a.notify(Lf("branch.success.delete", name), 0)
		}
	})
}

func (a *App) doBranchMerge() {
	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		} else if a.state.BranchTab == 1 && b.IsRemote {
			filtered = append(filtered, b)
		}
	}
	if a.state.BranchSelected >= len(filtered) {
		return
	}
	branch := filtered[a.state.BranchSelected]
	a.doBranchMergeTo(branch.Name)
}

func (a *App) doBranchMergeTo(name string) {
	a.showConfirm(L("branch.merge.confirm"), Lf("branch.merge.confirm", name), func() {
		err := a.git.MergeBranch(name)
		if err != nil {
			a.showInfo(L("branch.conflict"), err.Error())
		} else {
			a.notify(L("branch.success.merge"), 0)
		}
	})
}

func (a *App) doBranchRebase() {
	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		} else if a.state.BranchTab == 1 && b.IsRemote {
			filtered = append(filtered, b)
		}
	}
	if a.state.BranchSelected >= len(filtered) {
		return
	}
	branch := filtered[a.state.BranchSelected]
	a.doBranchRebaseTo(branch.Name)
}

func (a *App) doBranchRebaseTo(name string) {
	a.showConfirm(L("branch.rebase.confirm"), Lf("branch.rebase.confirm", name), func() {
		err := a.git.RebaseBranch(name)
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.notify(L("branch.success.rebase"), 0)
		}
	})
}

func (a *App) handleLogKey(key Key) {
	if a.state.LogSearchMode {
		switch key.Type {
		case KeyEscape:
			a.state.LogSearchMode = false
			a.state.LogSearchQuery = ""
		case KeyEnter:
			a.state.LogSearchMode = false
		case KeyBackspace:
			if a.state.LogSearchCursor > 0 {
				runes := []rune(a.state.LogSearchQuery)
				a.state.LogSearchQuery = string(runes[:a.state.LogSearchCursor-1]) + string(runes[a.state.LogSearchCursor:])
				a.state.LogSearchCursor--
			}
		case KeyRune:
			runes := []rune(a.state.LogSearchQuery)
			newRunes := make([]rune, 0, len(runes)+1)
			newRunes = append(newRunes, runes[:a.state.LogSearchCursor]...)
			newRunes = append(newRunes, key.Rune)
			newRunes = append(newRunes, runes[a.state.LogSearchCursor:]...)
			a.state.LogSearchQuery = string(newRunes)
			a.state.LogSearchCursor++
		}
		return
	}

	switch key.Type {
	case KeyEscape:
		if a.state.LogShowDetail {
			a.state.LogShowDetail = false
		} else {
			a.popPage()
		}
	case KeyUp:
		if a.state.LogSelected > 0 {
			a.state.LogSelected--
		}
	case KeyDown:
		a.state.LogSelected++
	case KeyEnter:
		if a.state.LogShowDetail {
			a.state.LogShowDetail = false
		} else {
			a.handleLogAction()
		}
	case KeyPageUp:
		a.state.LogSelected -= 10
		if a.state.LogSelected < 0 {
			a.state.LogSelected = 0
		}
	case KeyPageDown:
		a.state.LogSelected += 10
	case KeyRune:
		switch key.Rune {
		case 'd':
			a.state.LogShowDetail = true
			a.state.LogDetailScroll = 0
		case '/':
			a.state.LogSearchMode = true
			a.state.LogSearchQuery = ""
			a.state.LogSearchCursor = 0
		case 'q':
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleLogAction() {
	commits, _ := a.git.Log(100)
	if a.state.LogSelected >= len(commits) {
		return
	}
	commit := commits[a.state.LogSelected]

	items := []MenuItem{
		{Label: L("log.detail"), Hotkey: "d"},
		{Label: L("log.checkout"), Hotkey: "o"},
		{Label: L("log.cherry"), Hotkey: "c"},
		{Label: L("log.revert"), Hotkey: "r"},
		{Label: L("log.copy.hash"), Hotkey: "y"},
	}

	a.showMenuDialog(commit.Short+" "+commit.Message, items, func(idx int) {
		switch idx {
		case 0:
			a.state.LogShowDetail = true
			a.state.LogDetailScroll = 0
		case 1:
			a.showConfirm(L("log.checkout"), commit.Short, func() {
				a.git.CheckoutCommit(commit.Hash)
				a.notify("✅ "+L("log.checkout")+" "+commit.Short, 0)
			})
		case 2:
			a.showConfirm(L("log.cherry"), commit.Short, func() {
				err := a.git.CherryPick(commit.Hash)
				if err != nil {
					a.showInfo(L("error.git"), err.Error())
				} else {
					a.notify("✅ "+L("log.cherry")+" 完成", 0)
				}
			})
		case 3:
			a.showConfirm(L("log.revert"), commit.Short, func() {
				err := a.git.RevertCommit(commit.Hash)
				if err != nil {
					a.showInfo(L("error.git"), err.Error())
				} else {
					a.notify("✅ "+L("log.revert")+" 完成", 0)
				}
			})
		case 4:
			a.notify("📋 "+L("log.copy.hash")+": "+commit.Hash, 0)
		}
	})
}

func (a *App) handleStashKey(key Key) {
	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.StashSelected > 0 {
			a.state.StashSelected--
		}
	case KeyDown:
		stashes, _ := a.git.StashList()
		if a.state.StashSelected < len(stashes)-1 {
			a.state.StashSelected++
		}
	case KeyEnter:
		a.handleStashAction()
	case KeyRune:
		switch key.Rune {
		case 's':
			a.showInput(L("stash.save"), L("stash.save.prompt"), func(msg string) {
				err := a.git.StashSave(msg)
				if err != nil {
					a.showInfo(L("error.git"), err.Error())
				} else {
					a.notify(L("stash.success.save"), 0)
				}
			})
		case 'p':
			a.doStashPop()
		case 'a':
			a.doStashApply()
		case 'd':
			a.doStashDrop()
		case 'c':
			a.showConfirm(L("stash.clear.confirm"), L("stash.clear.confirm"), func() {
				a.git.StashClear()
				a.notify(L("stash.success.clear"), 0)
			})
		case 'q':
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleStashAction() {
	stashes, _ := a.git.StashList()
	if a.state.StashSelected >= len(stashes) {
		return
	}
	stash := stashes[a.state.StashSelected]

	items := []MenuItem{
		{Label: L("stash.pop"), Hotkey: "p"},
		{Label: L("stash.apply"), Hotkey: "a"},
		{Label: L("stash.drop"), Hotkey: "d"},
	}

	a.showMenuDialog(stash.Message, items, func(idx int) {
		switch idx {
		case 0:
			a.doStashPopIdx(stash.Index)
		case 1:
			a.doStashApplyIdx(stash.Index)
		case 2:
			a.doStashDropIdx(stash.Index)
		}
	})
}

func (a *App) doStashPop() {
	stashes, _ := a.git.StashList()
	if len(stashes) == 0 {
		a.showInfo(L("stash.empty"), L("stash.empty"))
		return
	}
	a.doStashPopIdx(stashes[a.state.StashSelected].Index)
}

func (a *App) doStashPopIdx(idx int) {
	err := a.git.StashPop(idx)
	if err != nil {
		a.showInfo(L("error.git"), err.Error())
	} else {
		a.notify(L("stash.success.pop"), 0)
	}
}

func (a *App) doStashApply() {
	stashes, _ := a.git.StashList()
	if len(stashes) == 0 {
		a.showInfo(L("stash.empty"), L("stash.empty"))
		return
	}
	a.doStashApplyIdx(stashes[a.state.StashSelected].Index)
}

func (a *App) doStashApplyIdx(idx int) {
	err := a.git.StashApply(idx)
	if err != nil {
		a.showInfo(L("error.git"), err.Error())
	} else {
		a.notify(L("stash.success.apply"), 0)
	}
}

func (a *App) doStashDrop() {
	stashes, _ := a.git.StashList()
	if len(stashes) == 0 {
		return
	}
	a.doStashDropIdx(stashes[a.state.StashSelected].Index)
}

func (a *App) doStashDropIdx(idx int) {
	a.showConfirm(L("stash.drop.confirm"), L("stash.drop.confirm"), func() {
		err := a.git.StashDrop(idx)
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.notify(L("stash.success.drop"), 0)
		}
	})
}

func (a *App) handleDiffKey(key Key) {
	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.DiffScroll > 0 {
			a.state.DiffScroll--
		}
	case KeyDown:
		a.state.DiffScroll++
	case KeyPageUp:
		a.state.DiffScroll -= 20
		if a.state.DiffScroll < 0 {
			a.state.DiffScroll = 0
		}
	case KeyPageDown:
		a.state.DiffScroll += 20
	case KeyTab:
		a.state.DiffTab = (a.state.DiffTab + 1) % 2
		a.state.DiffScroll = 0
	case KeyRune:
		if key.Rune == 'q' {
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleRemoteKey(key Key) {
	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.RemoteSelected > 0 {
			a.state.RemoteSelected--
		}
	case KeyDown:
		remotes, _ := a.git.Remotes()
		if a.state.RemoteSelected < len(remotes) {
			a.state.RemoteSelected++
		}
	case KeyEnter:
		a.handleRemoteAction()
	case KeyRune:
		switch key.Rune {
		case 'f':
			a.doFetch()
		case 'p':
			a.doPull()
		case 'u':
			a.doPush(false)
		case 'q':
			a.popPage()
		}
	case KeyF5:
		a.term.GetSize()
	}
}

func (a *App) handleRemoteAction() {
	remotes, _ := a.git.Remotes()
	if a.state.RemoteSelected >= len(remotes) {
		return
	}
	remote := remotes[a.state.RemoteSelected]

	items := []MenuItem{
		{Label: L("remote.fetch"), Hotkey: "f"},
		{Label: L("remote.pull"), Hotkey: "p"},
		{Label: L("remote.push"), Hotkey: "u"},
		{Label: L("remote.push.force"), Hotkey: "F"},
		{Label: L("remote.remove"), Hotkey: "d"},
	}

	a.showMenuDialog(remote, items, func(idx int) {
		switch idx {
		case 0:
			a.doFetchRemote(remote)
		case 1:
			a.doPullRemote(remote)
		case 2:
			a.doPushRemote(remote, false)
		case 3:
			a.doPushRemote(remote, true)
		case 4:
			a.showConfirm(L("remote.remove"), remote, func() {
				a.git.RemoveRemote(remote)
				a.notify("✅ "+L("remote.remove")+" "+remote, 0)
			})
		}
	})
}

func (a *App) doFetch() {
	err := a.git.Fetch("")
	if err != nil {
		a.showInfo(L("error.git"), err.Error())
	} else {
		a.notify(L("remote.success.fetch"), 0)
	}
}

func (a *App) doFetchRemote(remote string) {
	err := a.git.Fetch(remote)
	if err != nil {
		a.showInfo(L("error.git"), err.Error())
	} else {
		a.notify(L("remote.success.fetch"), 0)
	}
}

func (a *App) doPull() {
	a.showConfirm(L("remote.pull.confirm"), L("remote.pull.confirm"), func() {
		err := a.git.Pull("", "")
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.notify(L("remote.success.pull"), 0)
		}
	})
}

func (a *App) doPullRemote(remote string) {
	a.showConfirm(L("remote.pull.confirm"), remote, func() {
		err := a.git.Pull(remote, "")
		if err != nil {
			a.showInfo(L("error.git"), err.Error())
		} else {
			a.notify(L("remote.success.pull"), 0)
		}
	})
}

func (a *App) doPush(force bool) {
	if force {
		a.showConfirm(L("remote.push.force.confirm"), L("remote.push.force.confirm"), func() {
			err := a.git.Push("", "", true)
			if err != nil {
				a.showInfo(L("error.git"), err.Error())
			} else {
				a.notify(L("remote.success.push"), 0)
			}
		})
	} else {
		a.showConfirm(L("remote.push.confirm"), L("remote.push.confirm"), func() {
			err := a.git.Push("", "", false)
			if err != nil {
				a.showInfo(L("error.git"), err.Error())
			} else {
				a.notify(L("remote.success.push"), 0)
			}
		})
	}
}

func (a *App) doPushRemote(remote string, force bool) {
	branch := a.git.CurrentBranch()
	if force {
		a.showConfirm(L("remote.push.force.confirm"), remote+" "+branch, func() {
			err := a.git.Push(remote, branch, true)
			if err != nil {
				a.showInfo(L("error.git"), err.Error())
			} else {
				a.notify(L("remote.success.push"), 0)
			}
		})
	} else {
		a.showConfirm(L("remote.push.confirm"), remote+" "+branch, func() {
			err := a.git.Push(remote, branch, false)
			if err != nil {
				a.showInfo(L("error.git"), err.Error())
			} else {
				a.notify(L("remote.success.push"), 0)
			}
		})
	}
}

func (a *App) handleSettingsKey(key Key) {
	settingsItems := a.getSettingsItems()
	switch key.Type {
	case KeyEscape:
		a.popPage()
	case KeyUp:
		if a.state.SettingsSelected > 0 {
			a.state.SettingsSelected--
		}
	case KeyDown:
		if a.state.SettingsSelected < len(settingsItems)-1 {
			a.state.SettingsSelected++
		}
	case KeyEnter:
		a.handleSettingsAction()
	case KeyRune:
		if key.Rune == 'q' {
			a.popPage()
		}
	}
}

func (a *App) getSettingsItems() []MenuItem {
	themes := []string{L("settings.theme.catppuccin"), L("settings.theme.dark")}
	currentTheme := themes[a.state.ThemeIdx]

	return []MenuItem{
		{Label: fmt.Sprintf("%s: %s", L("settings.theme"), currentTheme), Hotkey: "1"},
		{Label: fmt.Sprintf("%s: %s", L("settings.username"), a.git.Config("user.name")), Hotkey: "2"},
		{Label: fmt.Sprintf("%s: %s", L("settings.email"), a.git.Config("user.email")), Hotkey: "3"},
		{Label: L("settings.about"), Hotkey: "a"},
	}
}

func (a *App) handleSettingsAction() {
	switch a.state.SettingsSelected {
	case 0:
		a.state.ThemeIdx = (a.state.ThemeIdx + 1) % 2
		themes := []Theme{CatppuccinTheme, DarkTheme}
		a.ui.theme = themes[a.state.ThemeIdx]
	case 1:
		current := a.git.Config("user.name")
		a.showInput(L("settings.username"), current, func(val string) {
			a.git.SetConfig("user.name", val)
			a.notify("✅ "+L("settings.username")+" 已更新", 0)
		})
	case 2:
		current := a.git.Config("user.email")
		a.showInput(L("settings.email"), current, func(val string) {
			a.git.SetConfig("user.email", val)
			a.notify("✅ "+L("settings.email")+" 已更新", 0)
		})
	case 3:
		a.showInfo(L("settings.about"), L("settings.about.text"))
	}
}

func (a *App) render() {
	a.term.GetSize()

	w := a.term.Width()
	h := a.term.Height()

	a.term.MoveCursor(1, 1)
	a.term.Write("\x1b[J")

	a.renderHeader(Rect{0, 0, w, 1})
	a.renderStatusBar(Rect{0, h - 1, w, 1})

	contentRect := Rect{0, 1, w, h - 2}

	if a.state.Notification != "" {
		a.renderNotification(Rect{0, h - 2, w, 1})
		contentRect.H--
	}

	if a.state.Dialog.Type != DialogNone {
		a.renderPage(contentRect)
		a.renderDialog()
	} else {
		a.renderPage(contentRect)
	}

	a.term.Write("\x1b[0m")
}

func (a *App) renderHeader(r Rect) {
	branch := a.git.CurrentBranch()
	hash := a.git.ShortHash()
	title := L("app.title")
	headerText := fmt.Sprintf(" %s  🌿 %s  📎 %s ", title, branch, hash)
	a.ui.DrawHeader(r, headerText)
}

func (a *App) renderStatusBar(r Rect) {
	var sections []string
	page := a.state.CurrentPage

	switch page {
	case PageMain:
		sections = []string{
			L("statusbar.nav"),
			L("statusbar.ok"),
			L("statusbar.select"),
			L("statusbar.quit"),
		}
	case PageStatus:
		tabs := []string{L("status.staged"), L("status.unstaged"), L("status.untracked")}
		sections = []string{
			L("statusbar.tab"),
			tabs[a.state.StatusTab],
			L("statusbar.ok"),
			L("statusbar.back"),
		}
	case PageCommit:
		tabs := []string{L("status.staged"), L("status.unstaged"), L("status.untracked")}
		sections = []string{
			L("statusbar.tab"),
			tabs[a.state.CommitTab],
			L("statusbar.edit"),
			L("statusbar.commit"),
			L("statusbar.stage"),
			L("statusbar.stageall"),
			L("statusbar.unstageall"),
			L("statusbar.amend"),
		}
	case PageBranch:
		tabs := []string{L("branch.local"), L("branch.remote")}
		sections = []string{
			L("statusbar.tab"),
			tabs[a.state.BranchTab],
			L("statusbar.new"),
			L("statusbar.switch"),
			L("statusbar.delete"),
			L("statusbar.merge"),
			L("statusbar.rebase"),
		}
	case PageLog:
		sections = []string{
			L("statusbar.detail"),
			L("statusbar.search"),
			L("statusbar.ok"),
			L("statusbar.pgupdn"),
		}
	case PageStash:
		sections = []string{
			L("statusbar.save"),
			L("statusbar.pop"),
			L("statusbar.apply"),
			L("statusbar.drop"),
			L("statusbar.clear"),
		}
	case PageDiff:
		tabs := []string{L("diff.staged"), L("diff.unstaged")}
		sections = []string{
			L("statusbar.tab"),
			tabs[a.state.DiffTab],
			L("statusbar.scroll"),
			L("statusbar.pgupdn"),
		}
	case PageRemote:
		sections = []string{
			L("statusbar.fetch"),
			L("statusbar.pull"),
			L("statusbar.push"),
			L("statusbar.ok"),
		}
	case PageSettings:
		sections = []string{
			L("statusbar.ok"),
			L("statusbar.back"),
		}
	}

	a.ui.DrawStatusBar(r, sections)
}

func (a *App) renderNotification(r Rect) {
	var color Color
	switch a.state.NotifType {
	case 0:
		color = a.ui.theme.Success
	case 1:
		color = a.ui.theme.Warning
	case 2:
		color = a.ui.theme.Error
	default:
		color = a.ui.theme.Text
	}
	text := string(color) + string(Bold) + " " + a.state.Notification + " " + string(Reset)
	a.ui.SetPixel(r.Y, r.X, string(a.ui.theme.AccentBg)+padRight(text, r.W)+string(Reset))
}

func (a *App) renderPage(r Rect) {
	switch a.state.CurrentPage {
	case PageMain:
		a.renderMainPage(r)
	case PageStatus:
		a.renderStatusPage(r)
	case PageCommit:
		a.renderCommitPage(r)
	case PageBranch:
		a.renderBranchPage(r)
	case PageLog:
		a.renderLogPage(r)
	case PageStash:
		a.renderStashPage(r)
	case PageDiff:
		a.renderDiffPage(r)
	case PageRemote:
		a.renderRemotePage(r)
	case PageSettings:
		a.renderSettingsPage(r)
	}
}

func (a *App) renderMainPage(r Rect) {
	menuW := 34
	menuH := min(len(a.getMainMenuItems())+4, r.H-2)
	menuX := 1
	menuY := 1

	menuRect := Rect{menuX, menuY, menuW, menuH}
	a.ui.DrawBox(menuRect, L("app.title"))

	items := a.getMainMenuItems()
	maxVisible := menuH - 2
	scroll := a.state.MenuScroll
	if a.state.MenuSelected < scroll {
		scroll = a.state.MenuSelected
	}
	if a.state.MenuSelected >= scroll+maxVisible {
		scroll = a.state.MenuSelected - maxVisible + 1
	}
	a.state.MenuScroll = scroll

	menuInner := Rect{menuX, menuY, menuW, menuH}
	a.ui.DrawMenu(menuInner, items, a.state.MenuSelected, scroll)

	a.renderGitInfo(r, menuX, menuY, menuW, menuH)
}

func (a *App) renderGitInfo(r Rect, menuX, menuY, menuW, menuH int) {
	infoX := menuX + menuW + 1
	infoW := r.W - infoX - 1
	if infoW < 20 {
		return
	}

	infoY := menuY
	infoH := menuH

	infoRect := Rect{infoX, infoY, infoW, infoH}
	a.ui.DrawBox(infoRect, "📋 "+L("status.title"))

	branch := a.git.CurrentBranch()
	hash := a.git.ShortHash()
	ahead, behind := a.git.AheadBehind()

	row := infoY + 1
	a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Accent)+"🌿 "+L("status.branch")+": "+string(Reset)+string(a.ui.theme.Text)+truncate(branch, infoW-20)+string(Reset))

	row++
	a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Accent)+"📎 "+L("log.hash")+": "+string(Reset)+string(a.ui.theme.Text)+hash+string(Reset))

	row++
	if ahead > 0 || behind > 0 {
		syncText := ""
		if ahead > 0 {
			syncText += string(a.ui.theme.Success) + fmt.Sprintf("↑%d", ahead) + string(Reset)
		}
		if behind > 0 {
			syncText += string(a.ui.theme.Warning) + fmt.Sprintf("↓%d", behind) + string(Reset)
		}
		a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Accent)+"🔄 Sync: "+string(Reset)+syncText)
	} else {
		a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Accent)+"🔄 Sync: "+string(Reset)+string(a.ui.theme.Success)+"✓ OK"+string(Reset))
	}

	staged, unstaged, untracked, _ := a.git.Status()
	row += 2
	a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Success)+fmt.Sprintf("✅ "+L("status.staged")+": %d", len(staged))+string(Reset))
	row++
	a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Warning)+fmt.Sprintf("⚠️  "+L("status.unstaged")+": %d", len(unstaged))+string(Reset))
	row++
	a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Error)+fmt.Sprintf("❓ "+L("status.untracked")+": %d", len(untracked))+string(Reset))

	if len(staged) == 0 && len(unstaged) == 0 && len(untracked) == 0 {
		row += 2
		a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Success)+L("status.clean")+string(Reset))
	}

	remotes, _ := a.git.Remotes()
	row += 2
	if len(remotes) > 0 {
		a.ui.SetPixel(row, infoX+2, string(a.ui.theme.Accent)+"🌐 "+L("remote.title")+": "+string(Reset))
		for i, remote := range remotes {
			url := a.git.RemoteURL(remote)
			if i > 0 {
				row++
			}
			maxURLW := infoW - 8
			if maxURLW < 10 {
				maxURLW = 10
			}
			a.ui.SetPixel(row, infoX+4, string(a.ui.theme.TextDim)+remote+": "+truncate(url, maxURLW-strWidth(remote)-2)+string(Reset))
		}
	} else {
		a.ui.SetPixel(row, infoX+2, string(a.ui.theme.TextDim)+L("status.no.remote")+string(Reset))
	}
}

func (a *App) renderStatusPage(r Rect) {
	a.ui.DrawBox(r, L("status.title"))

	tabRect := Rect{r.X + 1, r.Y + 1, r.W - 2, 1}
	tabs := []string{L("status.staged"), L("status.unstaged"), L("status.untracked")}
	a.ui.DrawTabs(tabRect, tabs, a.state.StatusTab)

	listRect := Rect{r.X, r.Y + 2, r.W, r.H - 2}

	staged, unstaged, untracked, _ := a.git.Status()
	var files []FileStatus
	switch a.state.StatusTab {
	case 0:
		files = staged
	case 1:
		files = unstaged
	case 2:
		files = untracked
	}

	if len(files) == 0 {
		emptyMsg := ""
		switch a.state.StatusTab {
		case 0:
			emptyMsg = L("status.staged") + " - 0"
		case 1:
			emptyMsg = L("status.unstaged") + " - 0"
		case 2:
			emptyMsg = L("status.untracked") + " - 0"
		}
		a.ui.SetPixel(r.Y+r.H/2, r.X+r.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	var items []ListItem
	for _, f := range files {
		tagColor := a.ui.theme.TextDim
		switch {
		case f.Modified:
			tagColor = a.ui.theme.Warning
		case f.Added:
			tagColor = a.ui.theme.Success
		case f.Deleted:
			tagColor = a.ui.theme.Error
		case f.Untracked:
			tagColor = a.ui.theme.TextDim
		}
		items = append(items, ListItem{
			Columns:  []string{f.Path},
			Tag:      f.Status,
			TagColor: tagColor,
		})
	}

	if a.state.StatusSelected >= len(items) {
		a.state.StatusSelected = len(items) - 1
	}
	if a.state.StatusSelected < 0 {
		a.state.StatusSelected = 0
	}

	maxVisible := listRect.H - 2
	scroll := a.state.StatusScroll
	if a.state.StatusSelected < scroll {
		scroll = a.state.StatusSelected
	}
	if a.state.StatusSelected >= scroll+maxVisible {
		scroll = a.state.StatusSelected - maxVisible + 1
	}
	a.state.StatusScroll = scroll

	a.ui.DrawList(listRect, items, a.state.StatusSelected, scroll, []int{listRect.W - 14})
}

func (a *App) renderCommitPage(r Rect) {
	splitY := r.Y + r.H*2/3
	listRect := Rect{r.X, r.Y, r.W, splitY - r.Y}
	inputRect := Rect{r.X, splitY, r.W, r.H - (splitY - r.Y)}

	a.ui.DrawBox(listRect, L("commit.title"))

	tabRect := Rect{listRect.X + 1, listRect.Y + 1, listRect.W - 2, 1}
	tabs := []string{L("status.staged"), L("status.unstaged"), L("status.untracked")}
	a.ui.DrawTabs(tabRect, tabs, a.state.CommitTab)

	fileListRect := Rect{listRect.X, listRect.Y + 2, listRect.W, listRect.H - 2}

	staged, unstaged, untracked, _ := a.git.Status()
	var files []FileStatus
	switch a.state.CommitTab {
	case 0:
		files = staged
	case 1:
		files = unstaged
	case 2:
		files = untracked
	}

	if len(files) > 0 {
		var items []ListItem
		for _, f := range files {
			tagColor := a.ui.theme.TextDim
			switch {
			case f.Modified:
				tagColor = a.ui.theme.Warning
			case f.Added:
				tagColor = a.ui.theme.Success
			case f.Deleted:
				tagColor = a.ui.theme.Error
			case f.Untracked:
				tagColor = a.ui.theme.TextDim
			}
			items = append(items, ListItem{
				Columns:  []string{f.Path},
				Tag:      f.Status,
				TagColor: tagColor,
			})
		}
		if a.state.CommitSelected >= len(items) {
			a.state.CommitSelected = len(items) - 1
		}
		if a.state.CommitSelected < 0 {
			a.state.CommitSelected = 0
		}
		maxVisible := fileListRect.H - 2
		scroll := a.state.CommitScroll
		if a.state.CommitSelected < scroll {
			scroll = a.state.CommitSelected
		}
		if a.state.CommitSelected >= scroll+maxVisible {
			scroll = a.state.CommitSelected - maxVisible + 1
		}
		a.state.CommitScroll = scroll
		a.ui.DrawList(fileListRect, items, a.state.CommitSelected, scroll, []int{fileListRect.W - 14})
	} else {
		emptyMsg := "0 files"
		a.ui.SetPixel(fileListRect.Y+fileListRect.H/2, fileListRect.X+fileListRect.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
	}

	inputState := InputState{
		Value:  a.state.CommitMessage,
		Cursor: a.state.CommitMsgCursor,
		Active: a.state.CommitInputActive,
		Prompt: L("commit.message"),
	}
	a.ui.DrawInput(inputRect, inputState)
}

func (a *App) renderBranchPage(r Rect) {
	a.ui.DrawBox(r, L("branch.title"))

	tabRect := Rect{r.X + 1, r.Y + 1, r.W - 2, 1}
	tabs := []string{L("branch.local"), L("branch.remote")}
	a.ui.DrawTabs(tabRect, tabs, a.state.BranchTab)

	listRect := Rect{r.X, r.Y + 2, r.W, r.H - 2}

	branches, _ := a.git.Branches()
	var filtered []BranchInfo
	for _, b := range branches {
		if a.state.BranchTab == 0 && !b.IsRemote {
			filtered = append(filtered, b)
		} else if a.state.BranchTab == 1 && b.IsRemote {
			filtered = append(filtered, b)
		}
	}

	if len(filtered) == 0 {
		emptyMsg := "0 " + L("branch.local")
		a.ui.SetPixel(r.Y+r.H/2, r.X+r.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	var items []ListItem
	for _, b := range filtered {
		name := b.Name
		if b.IsRemote {
			name = strings.TrimPrefix(name, "remotes/")
		}
		tag := ""
		tagColor := a.ui.theme.TextDim
		if b.Current {
			tag = "*"
			tagColor = a.ui.theme.Success
		} else if b.IsRemote {
			tag = "R"
			tagColor = a.ui.theme.Accent
		}
		items = append(items, ListItem{
			Columns:  []string{name},
			Tag:      tag,
			TagColor: tagColor,
		})
	}

	if a.state.BranchSelected >= len(items) {
		a.state.BranchSelected = len(items) - 1
	}
	if a.state.BranchSelected < 0 {
		a.state.BranchSelected = 0
	}

	maxVisible := listRect.H - 2
	scroll := a.state.BranchScroll
	if a.state.BranchSelected < scroll {
		scroll = a.state.BranchSelected
	}
	if a.state.BranchSelected >= scroll+maxVisible {
		scroll = a.state.BranchSelected - maxVisible + 1
	}
	a.state.BranchScroll = scroll

	colW := listRect.W - 6
	if colW < 10 {
		colW = 10
	}
	a.ui.DrawList(listRect, items, a.state.BranchSelected, scroll, []int{colW})
}

func (a *App) renderLogPage(r Rect) {
	if a.state.LogShowDetail {
		a.renderLogDetail(r)
		return
	}

	a.ui.DrawBox(r, L("log.title"))

	if a.state.LogSearchMode {
		searchRect := Rect{r.X + 1, r.Y + 1, r.W - 2, 1}
		a.ui.SetPixel(searchRect.Y, searchRect.X, string(a.ui.theme.Accent)+"🔍 /"+a.state.LogSearchQuery+string(Reverse)+" "+string(Reset))
	}

	var commits []CommitInfo
	var err error
	if a.state.LogSearchQuery != "" {
		commits, err = a.git.SearchLog(a.state.LogSearchQuery, 100)
	} else {
		commits, err = a.git.Log(100)
	}
	if err != nil || len(commits) == 0 {
		emptyMsg := L("log.no.commits")
		a.ui.SetPixel(r.Y+r.H/2, r.X+r.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	if a.state.LogSelected >= len(commits) {
		a.state.LogSelected = len(commits) - 1
	}
	if a.state.LogSelected < 0 {
		a.state.LogSelected = 0
	}

	listRect := Rect{r.X, r.Y + 2, r.W, r.H - 2}

	var items []ListItem
	for _, c := range commits {
		items = append(items, ListItem{
			Columns:  []string{c.Short, c.Author, c.Date, c.Message},
			Tag:      "",
			TagColor: a.ui.theme.TextDim,
		})
	}

	maxVisible := listRect.H - 2
	scroll := a.state.LogScroll
	if a.state.LogSelected < scroll {
		scroll = a.state.LogSelected
	}
	if a.state.LogSelected >= scroll+maxVisible {
		scroll = a.state.LogSelected - maxVisible + 1
	}
	a.state.LogScroll = scroll

	hashW := 8
	authorW := 12
	dateW := 16
	msgW := listRect.W - hashW - authorW - dateW - 8
	if msgW < 10 {
		msgW = 10
	}

	a.ui.DrawList(listRect, items, a.state.LogSelected, scroll, []int{hashW, authorW, dateW, msgW})
}

func (a *App) renderLogDetail(r Rect) {
	commits, _ := a.git.Log(100)
	if a.state.LogSelected >= len(commits) {
		return
	}
	commit := commits[a.state.LogSelected]

	detail, _ := a.git.ShowCommit(commit.Hash)
	if detail == "" {
		return
	}

	a.ui.DrawBox(r, L("log.detail")+" - "+commit.Short)

	lines := strings.Split(detail, "\n")
	maxVisible := r.H - 2
	scroll := a.state.LogDetailScroll
	if scroll < 0 {
		scroll = 0
	}
	if scroll > len(lines)-maxVisible {
		scroll = len(lines) - maxVisible
	}
	if scroll < 0 {
		scroll = 0
	}
	a.state.LogDetailScroll = scroll

	for i := 0; i < maxVisible; i++ {
		lineIdx := i + scroll
		row := r.Y + 1 + i
		if row >= r.Y+r.H-1 || lineIdx >= len(lines) {
			break
		}
		line := lines[lineIdx]
		var color Color
		if strings.HasPrefix(line, "commit ") {
			color = a.ui.theme.Accent
		} else if strings.HasPrefix(line, "Author: ") {
			color = a.ui.theme.Success
		} else if strings.HasPrefix(line, "Date: ") {
			color = a.ui.theme.Warning
		} else if strings.HasPrefix(line, "    ") {
			color = a.ui.theme.Text
		} else {
			color = a.ui.theme.TextDim
		}
		a.ui.SetPixel(row, r.X+2, string(color)+truncate(line, r.W-4)+string(Reset))
	}
}

func (a *App) renderStashPage(r Rect) {
	a.ui.DrawBox(r, L("stash.title"))

	stashes, _ := a.git.StashList()
	if len(stashes) == 0 {
		emptyMsg := L("stash.empty")
		a.ui.SetPixel(r.Y+r.H/2, r.X+r.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	listRect := Rect{r.X, r.Y + 1, r.W, r.H - 1}

	var items []ListItem
	for _, s := range stashes {
		items = append(items, ListItem{
			Columns:  []string{fmt.Sprintf("stash@{%d}", s.Index), s.Message},
			Tag:      fmt.Sprintf("#%d", s.Index),
			TagColor: a.ui.theme.Accent,
		})
	}

	if a.state.StashSelected >= len(items) {
		a.state.StashSelected = len(items) - 1
	}
	if a.state.StashSelected < 0 {
		a.state.StashSelected = 0
	}

	maxVisible := listRect.H - 2
	scroll := a.state.StashScroll
	if a.state.StashSelected < scroll {
		scroll = a.state.StashSelected
	}
	if a.state.StashSelected >= scroll+maxVisible {
		scroll = a.state.StashSelected - maxVisible + 1
	}
	a.state.StashScroll = scroll

	a.ui.DrawList(listRect, items, a.state.StashSelected, scroll, []int{14, listRect.W - 24})
}

func (a *App) renderDiffPage(r Rect) {
	tabRect := Rect{r.X, r.Y, r.W, 1}
	tabs := []string{L("diff.staged"), L("diff.unstaged")}
	a.ui.DrawTabs(tabRect, tabs, a.state.DiffTab)

	diffRect := Rect{r.X, r.Y + 1, r.W, r.H - 1}

	var diffText string
	if a.state.DiffTab == 0 {
		diffText, _ = a.git.DiffStaged()
	} else {
		diffText, _ = a.git.Diff()
	}

	if diffText == "" {
		emptyMsg := L("diff.no.diff")
		a.ui.SetPixel(diffRect.Y+diffRect.H/2, diffRect.X+diffRect.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	lines := strings.Split(diffText, "\n")
	maxVisible := diffRect.H
	scroll := a.state.DiffScroll
	if scroll < 0 {
		scroll = 0
	}
	if scroll > len(lines)-maxVisible {
		scroll = len(lines) - maxVisible
	}
	if scroll < 0 {
		scroll = 0
	}
	a.state.DiffScroll = scroll

	a.ui.DrawDiff(diffRect, lines, scroll)
}

func (a *App) renderRemotePage(r Rect) {
	a.ui.DrawBox(r, L("remote.title"))

	remotes, _ := a.git.Remotes()
	if len(remotes) == 0 {
		emptyMsg := L("remote.no.remote")
		a.ui.SetPixel(r.Y+r.H/2, r.X+r.W/2-strWidth(emptyMsg)/2, string(a.ui.theme.TextDim)+emptyMsg+string(Reset))
		return
	}

	listRect := Rect{r.X, r.Y + 1, r.W, r.H - 1}

	var items []ListItem
	for _, remote := range remotes {
		url := a.git.RemoteURL(remote)
		items = append(items, ListItem{
			Columns:  []string{remote, url},
			Tag:      "🌐",
			TagColor: a.ui.theme.Accent,
		})
	}

	if a.state.RemoteSelected >= len(items) {
		a.state.RemoteSelected = len(items) - 1
	}
	if a.state.RemoteSelected < 0 {
		a.state.RemoteSelected = 0
	}

	maxVisible := listRect.H - 2
	scroll := a.state.RemoteScroll
	if a.state.RemoteSelected < scroll {
		scroll = a.state.RemoteSelected
	}
	if a.state.RemoteSelected >= scroll+maxVisible {
		scroll = a.state.RemoteSelected - maxVisible + 1
	}
	a.state.RemoteScroll = scroll

	a.ui.DrawList(listRect, items, a.state.RemoteSelected, scroll, []int{12, listRect.W - 20})
}

func (a *App) renderSettingsPage(r Rect) {
	a.ui.DrawBox(r, L("settings.title"))

	items := a.getSettingsItems()
	menuRect := Rect{r.X + 2, r.Y + 2, r.W - 4, min(len(items)+4, r.H-4)}
	a.ui.DrawBox(menuRect, "")

	scroll := 0
	a.ui.DrawMenu(menuRect, items, a.state.SettingsSelected, scroll)
}

func (a *App) renderDialog() {
	w := a.term.Width()
	h := a.term.Height()
	d := a.state.Dialog

	switch d.Type {
	case DialogConfirm, DialogInfo:
		lines := strings.Split(d.Message, "\n")
		dialogW := 50
		dialogH := len(lines) + 6
		if dialogW > w-4 {
			dialogW = w - 4
		}
		if dialogH > h-4 {
			dialogH = h - 4
		}
		dialogX := (w - dialogW) / 2
		dialogY := (h - dialogH) / 2

		dialogRect := Rect{dialogX, dialogY, dialogW, dialogH}
		a.ui.DrawDialog(dialogRect, d.Title, d.Message, d.Buttons, d.Selected)

	case DialogInput:
		dialogW := 60
		dialogH := 5
		if dialogW > w-4 {
			dialogW = w - 4
		}
		dialogX := (w - dialogW) / 2
		dialogY := (h - dialogH) / 2

		inputRect := Rect{dialogX, dialogY, dialogW, dialogH}
		a.ui.DrawInput(inputRect, InputState{
			Value:  d.InputValue,
			Cursor: d.InputCursor,
			Active: true,
			Prompt: d.Title,
		})

	case DialogMenu:
		dialogW := 44
		dialogH := len(d.MenuItems) + 4
		if dialogW > w-4 {
			dialogW = w - 4
		}
		if dialogH > h-4 {
			dialogH = h - 4
		}
		dialogX := (w - dialogW) / 2
		dialogY := (h - dialogH) / 2

		menuRect := Rect{dialogX, dialogY, dialogW, dialogH}
		a.ui.DrawBox(menuRect, d.Title)
		a.ui.DrawMenu(menuRect, d.MenuItems, d.MenuSelected, 0)
	}
}
