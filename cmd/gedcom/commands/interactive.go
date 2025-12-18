package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/lesfleursdelanuitdev/gedcom-go/cmd/gedcom/internal"
	"github.com/lesfleursdelanuitdev/gedcom-go/internal/parser"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom"
	"github.com/lesfleursdelanuitdev/gedcom-go/pkg/gedcom/query"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive [input.ged]",
	Short: "Interactive mode",
	Long:  "Start interactive mode to query GEDCOM data. Parse a file once, then perform multiple queries.",
	Args:  cobra.ExactArgs(1),
	RunE:  runInteractive,
}

// InteractiveState holds the state for interactive mode
type InteractiveState struct {
	tree  *gedcom.GedcomTree
	graph *query.Graph
	query *query.QueryBuilder
}

var state *InteractiveState

func init() {
	interactiveCmd.Flags().Bool("no-graph", false, "Don't build graph (faster startup, limited queries)")
}

func runInteractive(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	noGraph, _ := cmd.Flags().GetBool("no-graph")

	// Load config
	config, err := internal.LoadConfig("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize color
	internal.InitColor(config.Output.Color)

	// Check if file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		internal.PrintError("✗ File not found: %s\n", inputFile)
		return fmt.Errorf("file not found: %s", inputFile)
	}

	// Parse file
	internal.PrintInfo("ℹ Loading GEDCOM file: %s\n", inputFile)

	p := parser.NewHierarchicalParser()
	tree, err := p.Parse(inputFile)
	if err != nil {
		internal.PrintError("✗ Parse failed: %v\n", err)
		return err
	}

	// Get statistics
	individuals := tree.GetAllIndividuals()
	families := tree.GetAllFamilies()

	internal.PrintSuccess("✓ Loaded successfully\n")
	internal.PrintInfo("  Individuals: %d\n", len(individuals))
	internal.PrintInfo("  Families: %d\n", len(families))

	// Build graph if requested
	var graph *query.Graph
	var qb *query.QueryBuilder
	if !noGraph {
		internal.PrintInfo("ℹ Building graph...\n")
		graph, err = query.BuildGraph(tree)
		if err != nil {
			internal.PrintError("✗ Graph build failed: %v\n", err)
			return err
		}
		internal.PrintSuccess("✓ Graph built successfully\n")

		// Create query builder
		qb, err = query.NewQuery(tree)
		if err != nil {
			internal.PrintError("✗ Query builder failed: %v\n", err)
			return err
		}
	} else {
		internal.PrintInfo("ℹ Graph building skipped (limited queries available)\n")
	}

	// Initialize state
	state = &InteractiveState{
		tree:  tree,
		graph: graph,
		query: qb,
	}

	// Start interactive loop
	internal.PrintInfo("\n")
	internal.PrintSuccess("✓ Interactive mode ready\n")
	internal.PrintInfo("  Type 'help' for available commands\n")
	internal.PrintInfo("  Type 'exit' or 'quit' to exit\n\n")

	startREPL()

	return nil
}

func startREPL() {
	// Check if we have a TTY (terminal)
	// Try to use go-prompt, but fall back to simple input if it fails
	defer func() {
		if r := recover(); r != nil {
			// If go-prompt fails (no TTY), use simple input
			internal.PrintInfo("Note: Using simple input mode (no TTY detected)\n")
			startSimpleREPL()
		}
	}()

	// Check if stdin is a terminal
	fileInfo, err := os.Stdin.Stat()
	if err != nil || (fileInfo.Mode()&os.ModeCharDevice) == 0 {
		// Not a terminal, use simple input
		startSimpleREPL()
		return
	}

	// Use go-prompt for enhanced experience
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("gedcom> "),
		prompt.OptionTitle("GEDCOM Interactive Mode"),
		prompt.OptionPrefixTextColor(prompt.Cyan),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)
	p.Run()
}

func startSimpleREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("gedcom> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "" {
			continue
		}
		executor(line)
	}
	if err := scanner.Err(); err != nil {
		internal.PrintError("Error reading input: %v\n", err)
	}
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	parts := strings.Fields(in)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "exit", "quit", "q":
		internal.PrintInfo("Goodbye!\n")
		os.Exit(0)

	case "help", "h":
		printHelp()

	case "stats", "statistics":
		showStats()

	case "individual", "indi", "i":
		if len(args) == 0 {
			internal.PrintError("Usage: individual <xref>\n")
			return
		}
		showIndividual(args[0])

	case "family", "fam", "f":
		if len(args) == 0 {
			internal.PrintError("Usage: family <xref>\n")
			return
		}
		showFamily(args[0])

	case "search":
		if len(args) == 0 {
			internal.PrintError("Usage: search <name>\n")
			return
		}
		searchByName(strings.Join(args, " "))

	case "parents":
		if len(args) == 0 {
			internal.PrintError("Usage: parents <xref>\n")
			return
		}
		showParents(args[0])

	case "children":
		if len(args) == 0 {
			internal.PrintError("Usage: children <xref>\n")
			return
		}
		showChildren(args[0])

	case "siblings":
		if len(args) == 0 {
			internal.PrintError("Usage: siblings <xref>\n")
			return
		}
		showSiblings(args[0])

	case "spouses":
		if len(args) == 0 {
			internal.PrintError("Usage: spouses <xref>\n")
			return
		}
		showSpouses(args[0])

	case "ancestors":
		if len(args) == 0 {
			internal.PrintError("Usage: ancestors <xref> [max-generations]\n")
			return
		}
		maxGen := -1
		if len(args) > 1 {
			fmt.Sscanf(args[1], "%d", &maxGen)
		}
		showAncestors(args[0], maxGen)

	case "descendants":
		if len(args) == 0 {
			internal.PrintError("Usage: descendants <xref> [max-generations]\n")
			return
		}
		maxGen := -1
		if len(args) > 1 {
			fmt.Sscanf(args[1], "%d", &maxGen)
		}
		showDescendants(args[0], maxGen)

	case "relationship", "rel":
		if len(args) < 2 {
			internal.PrintError("Usage: relationship <xref1> <xref2>\n")
			return
		}
		showRelationship(args[0], args[1])

	case "path":
		if len(args) < 2 {
			internal.PrintError("Usage: path <xref1> <xref2>\n")
			return
		}
		showPath(args[0], args[1])

	default:
		internal.PrintError("Unknown command: %s\n", command)
		internal.PrintInfo("Type 'help' for available commands\n")
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "help", Description: "Show help"},
		{Text: "exit", Description: "Exit interactive mode"},
		{Text: "quit", Description: "Exit interactive mode"},
		{Text: "stats", Description: "Show statistics"},
		{Text: "individual", Description: "Show individual details"},
		{Text: "family", Description: "Show family details"},
		{Text: "search", Description: "Search by name"},
		{Text: "parents", Description: "Show parents"},
		{Text: "children", Description: "Show children"},
		{Text: "siblings", Description: "Show siblings"},
		{Text: "spouses", Description: "Show spouses"},
		{Text: "ancestors", Description: "Show ancestors"},
		{Text: "descendants", Description: "Show descendants"},
		{Text: "relationship", Description: "Calculate relationship"},
		{Text: "path", Description: "Find path between individuals"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func printHelp() {
	internal.PrintInfo("\nAvailable Commands:\n\n")
	internal.PrintInfo("  help, h                    Show this help\n")
	internal.PrintInfo("  exit, quit, q              Exit interactive mode\n")
	internal.PrintInfo("  stats                      Show file statistics\n\n")
	internal.PrintInfo("Individual Commands:\n")
	internal.PrintInfo("  individual <xref>          Show individual details\n")
	internal.PrintInfo("  family <xref>              Show family details\n")
	internal.PrintInfo("  search <name>              Search individuals by name\n\n")
	internal.PrintInfo("Relationship Commands:\n")
	internal.PrintInfo("  parents <xref>             Show parents\n")
	internal.PrintInfo("  children <xref>            Show children\n")
	internal.PrintInfo("  siblings <xref>            Show siblings\n")
	internal.PrintInfo("  spouses <xref>             Show spouses\n")
	internal.PrintInfo("  ancestors <xref> [n]       Show ancestors (optional max generations)\n")
	internal.PrintInfo("  descendants <xref> [n]     Show descendants (optional max generations)\n")
	internal.PrintInfo("  relationship <x1> <x2>    Calculate relationship between two individuals\n")
	internal.PrintInfo("  path <x1> <x2>            Find path between two individuals\n\n")
}

func showStats() {
	if state == nil || state.tree == nil {
		internal.PrintError("No data loaded\n")
		return
	}

	individuals := state.tree.GetAllIndividuals()
	families := state.tree.GetAllFamilies()
	notes := state.tree.GetAllNotes()
	sources := state.tree.GetAllSources()

	internal.PrintInfo("\nStatistics:\n")
	internal.PrintInfo("  Individuals: %d\n", len(individuals))
	internal.PrintInfo("  Families: %d\n", len(families))
	internal.PrintInfo("  Notes: %d\n", len(notes))
	internal.PrintInfo("  Sources: %d\n", len(sources))

	if state.graph != nil {
		internal.PrintInfo("  Graph nodes: %d\n", state.graph.NodeCount())
		internal.PrintInfo("  Graph edges: %d\n", state.graph.EdgeCount())
	}
	internal.PrintInfo("\n")
}

func showIndividual(xref string) {
	if state == nil || state.tree == nil {
		internal.PrintError("No data loaded\n")
		return
	}

	record := state.tree.GetIndividual(xref)
	if record == nil {
		internal.PrintError("Individual not found: %s\n", xref)
		return
	}

	indi, ok := record.(*gedcom.IndividualRecord)
	if !ok {
		internal.PrintError("Record is not an individual: %s\n", xref)
		return
	}

	internal.PrintInfo("\nIndividual: %s\n", indi.XrefID())
	internal.PrintInfo("  Name: %s\n", indi.GetName())
	internal.PrintInfo("  Sex: %s\n", indi.GetSex())
	internal.PrintInfo("  Birth: %s\n", indi.GetBirthDate())
	internal.PrintInfo("  Death: %s\n", indi.GetDeathDate())
	internal.PrintInfo("\n")
}

func showFamily(xref string) {
	if state == nil || state.tree == nil {
		internal.PrintError("No data loaded\n")
		return
	}

	record := state.tree.GetFamily(xref)
	if record == nil {
		internal.PrintError("Family not found: %s\n", xref)
		return
	}

	fam, ok := record.(*gedcom.FamilyRecord)
	if !ok {
		internal.PrintError("Record is not a family: %s\n", xref)
		return
	}

	internal.PrintInfo("\nFamily: %s\n", fam.XrefID())
	internal.PrintInfo("  Husband: %s\n", fam.GetValue("HUSB"))
	internal.PrintInfo("  Wife: %s\n", fam.GetValue("WIFE"))
	children := fam.GetValues("CHIL")
	internal.PrintInfo("  Children: %d\n", len(children))
	for _, child := range children {
		internal.PrintInfo("    - %s\n", child)
	}
	internal.PrintInfo("\n")
}

func searchByName(name string) {
	if state == nil || state.tree == nil {
		internal.PrintError("No data loaded\n")
		return
	}

	individuals := state.tree.GetAllIndividuals()
	nameLower := strings.ToLower(name)
	found := 0

	internal.PrintInfo("\nSearch results for '%s':\n", name)
	for xref, record := range individuals {
		indi, ok := record.(*gedcom.IndividualRecord)
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(indi.GetName()), nameLower) {
			internal.PrintInfo("  %s: %s\n", xref, indi.GetName())
			found++
			if found >= 20 {
				internal.PrintInfo("  ... (showing first 20 results)\n")
				break
			}
		}
	}

	if found == 0 {
		internal.PrintWarning("No matches found\n")
	}
	internal.PrintInfo("\n")
}

func showParents(xref string) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	parents, err := state.query.Individual(xref).Parents()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nParents of %s:\n", xref)
	if len(parents) == 0 {
		internal.PrintInfo("  No parents found\n")
	} else {
		for _, parent := range parents {
			internal.PrintInfo("  %s: %s\n", parent.XrefID(), parent.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showChildren(xref string) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	children, err := state.query.Individual(xref).Children()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nChildren of %s:\n", xref)
	if len(children) == 0 {
		internal.PrintInfo("  No children found\n")
	} else {
		for _, child := range children {
			internal.PrintInfo("  %s: %s\n", child.XrefID(), child.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showSiblings(xref string) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	siblings, err := state.query.Individual(xref).Siblings()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nSiblings of %s:\n", xref)
	if len(siblings) == 0 {
		internal.PrintInfo("  No siblings found\n")
	} else {
		for _, sibling := range siblings {
			internal.PrintInfo("  %s: %s\n", sibling.XrefID(), sibling.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showSpouses(xref string) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	spouses, err := state.query.Individual(xref).Spouses()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nSpouses of %s:\n", xref)
	if len(spouses) == 0 {
		internal.PrintInfo("  No spouses found\n")
	} else {
		for _, spouse := range spouses {
			internal.PrintInfo("  %s: %s\n", spouse.XrefID(), spouse.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showAncestors(xref string, maxGen int) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	ancestorQuery := state.query.Individual(xref).Ancestors()
	if maxGen > 0 {
		ancestorQuery = ancestorQuery.MaxGenerations(maxGen)
	}

	ancestors, err := ancestorQuery.Execute()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nAncestors of %s", xref)
	if maxGen > 0 {
		internal.PrintInfo(" (max %d generations)", maxGen)
	}
	internal.PrintInfo(":\n")
	if len(ancestors) == 0 {
		internal.PrintInfo("  No ancestors found\n")
	} else {
		for _, ancestor := range ancestors {
			internal.PrintInfo("  %s: %s\n", ancestor.XrefID(), ancestor.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showDescendants(xref string, maxGen int) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	descendantQuery := state.query.Individual(xref).Descendants()
	if maxGen > 0 {
		descendantQuery = descendantQuery.MaxGenerations(maxGen)
	}

	descendants, err := descendantQuery.Execute()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nDescendants of %s", xref)
	if maxGen > 0 {
		internal.PrintInfo(" (max %d generations)", maxGen)
	}
	internal.PrintInfo(":\n")
	if len(descendants) == 0 {
		internal.PrintInfo("  No descendants found\n")
	} else {
		for _, descendant := range descendants {
			internal.PrintInfo("  %s: %s\n", descendant.XrefID(), descendant.GetName())
		}
	}
	internal.PrintInfo("\n")
}

func showRelationship(xref1, xref2 string) {
	if state.query == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	result, err := state.query.Individual(xref1).RelationshipTo(xref2).Execute()
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	internal.PrintInfo("\nRelationship from %s to %s:\n", xref1, xref2)
	internal.PrintInfo("  Type: %s\n", result.RelationshipType)
	internal.PrintInfo("  Degree: %d\n", result.Degree)
	internal.PrintInfo("  Removal: %d\n", result.Removal)
	internal.PrintInfo("  Is Direct: %v\n", result.IsDirect)
	internal.PrintInfo("  Is Collateral: %v\n", result.IsCollateral)
	internal.PrintInfo("\n")
}

func showPath(xref1, xref2 string) {
	if state.graph == nil {
		internal.PrintError("Graph not built. Use --no-graph=false\n")
		return
	}

	path, err := state.graph.ShortestPath(xref1, xref2)
	if err != nil {
		internal.PrintError("Error: %v\n", err)
		return
	}

	if path == nil {
		internal.PrintWarning("No path found between %s and %s\n", xref1, xref2)
		return
	}

	internal.PrintInfo("\nPath from %s to %s:\n", xref1, xref2)
	internal.PrintInfo("  Type: %s\n", path.Type)
	internal.PrintInfo("  Length: %d\n", len(path.Nodes))
	for i, node := range path.Nodes {
		if i > 0 {
			internal.PrintInfo(" -> ")
		}
		internal.PrintInfo("%s", node.ID())
	}
	internal.PrintInfo("\n\n")
}

// GetInteractiveCommand returns the interactive command
func GetInteractiveCommand() *cobra.Command {
	return interactiveCmd
}
