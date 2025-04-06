package terms

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Term represents a term that should not be translated
type Term struct {
	Term        string   `json:"term"`
	Description string   `json:"description"`
	Context     []string `json:"context,omitempty"`
	Keep        bool     `json:"keep_untranslated"`
}

// TermManager manages the terms analysis and user interaction
type TermManager struct {
	terms []*Term
	input *strings.Reader // for testing
}

// New creates a new term manager
func New() *TermManager {
	return &TermManager{
		terms: []*Term{},
	}
}

// SetInput sets the input reader for testing
func (tm *TermManager) SetInput(input *strings.Reader) {
	tm.input = input
}

// AddTerms adds terms to the manager
func (tm *TermManager) AddTerms(terms []*Term) {
	tm.terms = terms
}

func (tm *TermManager) printTermsSummary() {
	fmt.Printf("Found %d terms that may not need translation:\n\n", len(tm.terms))
	for i, term := range tm.terms {
		fmt.Printf("%d. %s - %s\n", i+1, term.Term, term.Description)
		for _, ctx := range term.Context {
			fmt.Printf("   Context: %s\n", ctx)
		}
		fmt.Println()
	}
}

func (tm *TermManager) printMenu() {
	fmt.Println("Choose an action:")
	fmt.Println("1. Accept all terms as is")
	fmt.Println("2. Reject all terms")
	fmt.Println("3. Process terms interactively")
	fmt.Println("4. Edit terms in text editor")
}

func (tm *TermManager) readUserChoice() (string, error) {
	if tm.input != nil {
		fmt.Print("> ")
		var choice string
		if _, err := fmt.Fscanln(tm.input, &choice); err != nil {
			return "", fmt.Errorf("error reading choice: %w", err)
		}
		return choice, nil
	}

	fmt.Print("> ")
	var choice string
	if _, err := fmt.Scanln(&choice); err != nil || choice == "" {
		return "1", nil // default to accept all
	}
	return choice, nil
}

func (tm *TermManager) acceptAllTerms() {
	for _, term := range tm.terms {
		term.Keep = true
	}
}

func (tm *TermManager) rejectAllTerms() {
	for _, term := range tm.terms {
		term.Keep = false
	}
}

func (tm *TermManager) processSingleTerm(term *Term, index, total int) error {
	fmt.Printf("\nTerm: %s\n", term.Term)
	fmt.Printf("Description: %s\n", term.Description)
	for _, ctx := range term.Context {
		fmt.Printf("Context: %s\n", ctx)
	}
	fmt.Print("Keep untranslated? [Y/n/e/s]: ")

	response, err := tm.readUserResponse()
	if err != nil {
		return err
	}

	switch strings.ToLower(response) {
	case "n":
		term.Keep = false
	case "e":
		if err := tm.editTerm(term); err != nil {
			return err
		}
	case "s":
		return nil // skip
	default:
		term.Keep = true
	}

	fmt.Printf("Processed %d/%d terms\n", index+1, total)
	return nil
}

func (tm *TermManager) readUserResponse() (string, error) {
	if tm.input != nil {
		var response string
		if _, err := fmt.Fscanln(tm.input, &response); err != nil {
			return "", fmt.Errorf("error reading response: %w", err)
		}
		return response, nil
	}

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}
	return response, nil
}

func (tm *TermManager) editTerm(term *Term) error {
	fmt.Printf("New term [%s]: ", term.Term)
	newTerm, err := tm.readUserInput()
	if err != nil {
		return fmt.Errorf("error reading term: %w", err)
	}
	if newTerm != "" {
		term.Term = newTerm
	}

	fmt.Printf("New description [%s]: ", term.Description)
	newDesc, err := tm.readUserInput()
	if err != nil {
		return fmt.Errorf("error reading description: %w", err)
	}
	if newDesc != "" {
		term.Description = newDesc
	}
	term.Keep = true
	return nil
}

func (tm *TermManager) readUserInput() (string, error) {
	if tm.input != nil {
		var input string
		if _, err := fmt.Fscanln(tm.input, &input); err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
		return input, nil
	}

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return input, nil
}

// ProcessTermsInteractive interactively processes terms with the user
func (tm *TermManager) ProcessTermsInteractive() error {
	if len(tm.terms) == 0 {
		fmt.Println("No terms to process.")
		return nil
	}

	tm.printTermsSummary()
	tm.printMenu()

	choice, err := tm.readUserChoice()
	if err != nil {
		return err
	}

	switch choice {
	case "1":
		tm.acceptAllTerms()
	case "2":
		tm.rejectAllTerms()
	case "3":
		for i, term := range tm.terms {
			if err := tm.processSingleTerm(term, i, len(tm.terms)); err != nil {
				return err
			}
		}
	case "4":
		if err := tm.editInTextEditor(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	return nil
}

// editInTextEditor opens the terms in a text editor for batch editing
func (tm *TermManager) editInTextEditor() error {
	jsonData := struct {
		Terms []*Term `json:"terms"`
	}{
		Terms: tm.terms,
	}

	for _, term := range tm.terms {
		term.Keep = true
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling terms: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "terms-*.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("error writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %w", err)
	}

	editor := getDefaultEditor()
	cmd := exec.Command(editor, filepath.Clean(tmpFile.Name())) // #nosec G204
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening terms in %s. Save and exit the editor when finished.\n", editor)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running editor: %w", err)
	}

	editedData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("error reading edited file: %w", err)
	}

	var editedJSON struct {
		Terms []*Term `json:"terms"`
	}
	if err := json.Unmarshal(editedData, &editedJSON); err != nil {
		return fmt.Errorf("error parsing edited file: %w", err)
	}

	tm.terms = editedJSON.Terms
	return nil
}

// GetUntranslatableTerms returns a list of terms that should not be translated
func (tm *TermManager) GetUntranslatableTerms() []string {
	var results []string
	for _, term := range tm.terms {
		if term.Keep {
			results = append(results, term.Term)
		}
	}
	return results
}

// GetAllTerms returns all terms
func (tm *TermManager) GetAllTerms() []*Term {
	return tm.terms
}

// getDefaultEditor returns the default text editor based on OS
func getDefaultEditor() string {
	switch runtime.GOOS {
	case "windows":
		return "notepad"
	case "darwin":
		return "open -a TextEdit"
	default:
		if editor := os.Getenv("EDITOR"); editor != "" {
			return editor
		}
		if editor := os.Getenv("VISUAL"); editor != "" {
			return editor
		}
		return "nano"
	}
}
