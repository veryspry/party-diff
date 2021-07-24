package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	content       string
	ready         bool
	leftViewport  viewport.Model
	spinner       spinner.Model
}

const (
	// You generally won't need this unless you're processing stuff with some
	// pretty complicated ANSI escape sequences. Turn it on if you notice
	// flickering.
	//
	// Also note that high performance rendering only works for programs that
	// use the full size of the terminal. We're enabling that below with
	// tea.EnterAltScreen().
	useHighPerformanceRenderer = false

	headerHeight = 3
	footerHeight = 3
)

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func initViewPort(width int, height int, content string) viewport.Model {
	v := viewport.Model{Width: width, Height: height}
	v.YPosition = headerHeight
	v.HighPerformanceRendering = useHighPerformanceRenderer
	v.SetContent(content)
	return v
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		leftCmd  tea.Cmd
		cmds     []tea.Cmd
	)
	
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.content = fmt.Sprintf("%d %d", msg.Height, msg.Width)

		verticalMargins := headerHeight + footerHeight
		height := msg.Height - verticalMargins
		sideWidth := msg.Width / 2

		if !m.ready {
			// Since this program is using the full size of the viewport we need
			// to wait until we've received the window dimensions before we
			// can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.leftViewport = initViewPort(sideWidth, height, m.content)
			m.ready = true
		} else {
			height := msg.Height - verticalMargins

			c := ""

			for i := 0; i < height; i++ {
				c += "|\n"
			}

			m.leftViewport.Width = msg.Width
			m.leftViewport.Height = height
			m.leftViewport.SetContent(c)
		}


		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.leftViewport))
		}
	}

	// Because we're using the viewport's default update function (with pager-
	// style navigation) it's important that the viewport's update function:
	//
	// * Receives messages from the Bubble Tea runtime
	// * Returns commands to the Bubble Tea runtime
	//
	m.leftViewport, leftCmd = m.leftViewport.Update(msg)
	if useHighPerformanceRenderer {
		cmds = append(cmds, leftCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return fmt.Sprintf("\n\n %s Initializing", m.spinner.View())
	}

	header := "|-------------------------------------|"
	footer := "|-------------------------------------|"

	// Send the UI for rendering
	return fmt.Sprintf("%s\n%s\n%s", header, m.leftViewport.View(), footer)
}

func main() {
	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	content := ""
	// for i := 0; i < 13; i++ {
	// 	content += fmt.Sprintf("%d\n", i)
	// }
	var initialModel = model{
		content: content,
		spinner: s,
	}

	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// --------------------------------------------------------
// Example git diff
// --------------------------------------------------------

// diff --git a/api/database/queries.js b/api/database/queries.js
// index c9cc2af..873304d 100644
// --- a/api/database/queries.js
// +++ b/api/database/queries.js
// @@ -194,7 +194,17 @@ const mergeInNewUserApplications = async ({ email, selectedAppIds }) => {
// 	     .insert(newUserAppRecords);
//  };
//
// +const testTimeZoneThings = async () => {
// 	+  const res = await database('applications').first();
// +  // .where('id', 3);
// +  const d = new Date(res.created_at);
// +
// +  console.log(res, res.created_at, res.updated_at, d.toUTCString());
// +  return res;
// +};
// +
//  module.exports = {
// 	+  testTimeZoneThings,
//    database,
//    getAllUsers,
//    findUserById,
// diff --git a/api/index.js b/api/index.js
// index c84b2ab..684b141 100644
// --- a/api/index.js
// +++ b/api/index.js
// @@ -100,6 +100,14 @@ function routeHandlerWithError({ handler, errorMessage }) {
// 	   };
//  }
//
// +app.get('/timezone', routeHandlerWithError({
// 	+  handler: async (req, res) => {
// 	+    const query = await queries.testTimeZoneThings();
// +    res.status(200).json({ message: 'succes' });
// +  },
// +  errorMessage: 'error running timestamp test query',
// +}));
// +
//  /**
//   * Health endpoint to ensure that the server is running as expected
//   */
