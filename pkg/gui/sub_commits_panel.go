package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands"
)

// list panel functions

func (gui *Gui) getSelectedSubCommit() *commands.Commit {
	selectedLine := gui.State.Panels.SubCommits.SelectedLineIdx
	commits := gui.State.SubCommits
	if selectedLine == -1 || len(commits) == 0 {
		return nil
	}

	return commits[selectedLine]
}

func (gui *Gui) handleSubCommitSelect() error {
	commit := gui.getSelectedSubCommit()
	var task updateTask
	if commit == nil {
		task = gui.createRenderStringTask("No commits")
	} else {
		cmd := gui.OSCommand.ExecutableFromString(
			gui.GitCommand.ShowCmdStr(commit.Sha, gui.State.Modes.Filtering.Path),
		)

		task = gui.createRunPtyTask(cmd)
	}

	return gui.refreshMain(refreshMainOpts{
		main: &viewUpdateOpts{
			title: "Commit",
			task:  task,
		},
	})
}

func (gui *Gui) handleCheckoutSubCommit(g *gocui.Gui, v *gocui.View) error {
	commit := gui.getSelectedSubCommit()
	if commit == nil {
		return nil
	}

	err := gui.ask(askOpts{
		returnToView:       gui.getCommitsView(),
		returnFocusOnClose: true,
		title:              gui.Tr.SLocalize("checkoutCommit"),
		prompt:             gui.Tr.SLocalize("SureCheckoutThisCommit"),
		handleConfirm: func() error {
			return gui.handleCheckoutRef(commit.Sha, handleCheckoutRefOptions{})
		},
	})
	if err != nil {
		return err
	}

	gui.State.Panels.SubCommits.SelectedLineIdx = 0

	return nil
}

func (gui *Gui) handleCreateSubCommitResetMenu() error {
	commit := gui.getSelectedSubCommit()

	return gui.createResetMenu(commit.Sha)
}

func (gui *Gui) handleViewSubCommitFiles() error {
	commit := gui.getSelectedSubCommit()
	if commit == nil {
		return nil
	}

	return gui.switchToCommitFilesContext(commit.Sha, false, gui.Contexts.SubCommits.Context, "branches")
}

func (gui *Gui) switchToSubCommitsContext(refName string) error {
	// need to populate my sub commits
	builder := commands.NewCommitListBuilder(gui.Log, gui.GitCommand, gui.OSCommand, gui.Tr, gui.State.Modes.CherryPicking.CherryPickedCommits)

	commits, err := builder.GetCommits(
		commands.GetCommitsOptions{
			Limit:                gui.State.Panels.Commits.LimitCommits,
			FilterPath:           gui.State.Modes.Filtering.Path,
			IncludeRebaseCommits: false,
			RefName:              refName,
		},
	)
	if err != nil {
		return err
	}

	gui.State.SubCommits = commits
	gui.State.Panels.SubCommits.refName = refName
	gui.State.Panels.SubCommits.SelectedLineIdx = 0
	gui.Contexts.SubCommits.Context.SetParentContext(gui.currentSideContext())

	return gui.switchContext(gui.Contexts.SubCommits.Context)
}

func (gui *Gui) handleSwitchToSubCommits() error {
	currentContext := gui.currentSideContext()
	if currentContext == nil {
		return nil
	}

	gui.Log.Warn(currentContext.GetKey())

	return gui.switchToSubCommitsContext(currentContext.GetSelectedItemId())
}
