package main

import (
	"fmt"
	"os"
	"bufio"
	"strconv"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/waigani/diffparser"
	"github.com/sourcegraph/go-diff/diff"

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
		// m.content = fmt.Sprintf("%d %d", msg.Height, msg.Width)

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
			// m.leftViewport.SetContent(c)
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
	// 1. Receives messages from the Bubble Tea runtime
	// 2. Returns commands to the Bubble Tea runtime
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

	// Send to the UI for rendering
	return fmt.Sprintf("%s\n%s\n%s", header, m.leftViewport.View(), footer)
}

func removedLine(content string) string {
	return lipgloss.NewStyle().
		// red text
		Foreground(lipgloss.Color("#e23868")).
		Render(content)
}

func newLine(content string) string {
	return lipgloss.NewStyle().
		// blue text
		Foreground(lipgloss.Color("#04B575")).
		Render(content)
}

func parseStdIn() string {
	// git diff output info
	// https://stackoverflow.com/questions/2529441/how-to-read-the-output-from-git-diff

	/**
	 * deleted - show on left
	 * added - show on right
	 * modified - show on both sides
	 */

	scanner := bufio.NewScanner(os.Stdin)
	content := ""

	inputStr := ""

	diffBytes := make([]byte, 0)
	
	for {
		scanner.Scan()
		text := scanner.Text()
		b := scanner.Bytes()
		err := scanner.Err()

		if err != nil {
			panic(err)
		}

		if len(text) == 0 || len(b) == 0 {
			break
		}

		diffBytes = append(diffBytes, b...)	

		inputStr += text
		inputStr += "\n"
	}

	d, err := diff.ParseMultiFileDiff(diffBytes)
	fmt.Println("new stuff start")
	fmt.Println(string(diffBytes))
	fmt.Println()
	fmt.Println("new stuff end")

	if err != nil {
		panic(err)
	}

		
	diff, err := diffparser.Parse(inputStr)

	if err != nil {
		panic(err)
	}

	for _, file := range diff.Files {
		// content += file.DiffHeader
		// content += "\n"
				
		for _, hunk := range file.Hunks {
			lines := hunk.WholeRange
			o := hunk.OrigRange
			n := hunk.NewRange
			fmt.Println(lines.Start, lines.Length)
			fmt.Println(o.Start, o.Length)
			fmt.Println(n.Start, n.Length)
			fmt.Println("--------")

			content += strconv.Itoa(lines.Start)
			content += ", "
			content += strconv.Itoa(lines.Length)
			content += "\n"

				
			for _, line := range lines.Lines {

				content += strconv.Itoa(line.Position)
				switch line.Mode {
				case 0:
					// deleted
					content += removedLine(line.Content)
					content += "\n"
				case 1:
					// modified
				case 2:
					// created
					content += newLine(line.Content)
					content += "\n"
					
				}
			}
		}
	}

	return content
}

func main() {
	content := parseStdIn()

	s := spinner.NewModel()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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
