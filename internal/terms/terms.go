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

// ProcessTermsInteractive interactively processes terms with the user
func (tm *TermManager) ProcessTermsInteractive() error {
	if len(tm.terms) == 0 {
		fmt.Println("No terms to process.")
		return nil
	}

	fmt.Printf("Found %d terms that may not need translation:\n\n", len(tm.terms))
	for i, term := range tm.terms {
		fmt.Printf("%d. %s - %s\n", i+1, term.Term, term.Description)
		for _, ctx := range term.Context {
			fmt.Printf("   Context: %s\n", ctx)
		}
		fmt.Println()
	}

	fmt.Println("Choose an action:")
	fmt.Println("1. Accept all terms as is")
	fmt.Println("2. Reject all terms")
	fmt.Println("3. Process terms interactively")
	fmt.Println("4. Edit terms in text editor")

	var choice string
	if tm.input != nil {
		// Use test input
		fmt.Print("> ")
		if _, err := fmt.Fscanln(tm.input, &choice); err != nil {
			return fmt.Errorf("error reading choice: %w", err)
		}
	} else {
		// Use standard input
		fmt.Print("> ")
		if _, err := fmt.Scanln(&choice); err != nil {
			return fmt.Errorf("error reading choice: %w", err)
		}
	}

	switch choice {
	case "1":
		// accept all terms
		for _, term := range tm.terms {
			term.Keep = true
		}
	case "2":
		// reject all terms
		for _, term := range tm.terms {
			term.Keep = false
		}
	case "3":
		// process interactively
		for i, term := range tm.terms {
			fmt.Printf("\nTerm: %s\n", term.Term)
			fmt.Printf("Description: %s\n", term.Description)
			for _, ctx := range term.Context {
				fmt.Printf("Context: %s\n", ctx)
			}
			fmt.Print("Keep untranslated? [Y/n/e/s]: ")

			var response string
			if tm.input != nil {
				if _, err := fmt.Fscanln(tm.input, &response); err != nil {
					return fmt.Errorf("error reading response: %w", err)
				}
			} else {
				if _, err := fmt.Scanln(&response); err != nil {
					return fmt.Errorf("error reading response: %w", err)
				}
			}
			response = strings.ToLower(response)

			if response == "n" {
				term.Keep = false
			} else if response == "e" {
				// edit term
				fmt.Printf("New term [%s]: ", term.Term)
				var newTerm string
				if tm.input != nil {
					if _, err := fmt.Fscanln(tm.input, &newTerm); err != nil {
						return fmt.Errorf("error reading term: %w", err)
					}
				} else {
					if _, err := fmt.Scanln(&newTerm); err != nil {
						return fmt.Errorf("error reading term: %w", err)
					}
				}
				if newTerm != "" {
					term.Term = newTerm
				}

				fmt.Printf("New description [%s]: ", term.Description)
				var newDesc string
				if tm.input != nil {
					if _, err := fmt.Fscanln(tm.input, &newDesc); err != nil {
						return fmt.Errorf("error reading description: %w", err)
					}
				} else {
					if _, err := fmt.Scanln(&newDesc); err != nil {
						return fmt.Errorf("error reading description: %w", err)
					}
				}
				if newDesc != "" {
					term.Description = newDesc
				}
				term.Keep = true
			} else if response == "s" {
				// skip
				continue
			} else {
				// default is Yes
				term.Keep = true
			}

			// show progress
			fmt.Printf("Processed %d/%d terms\n", i+1, len(tm.terms))
		}
	case "4":
		// edit in text editor
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
	// prepare JSON data for editing
	jsonData := struct {
		Terms []*Term `json:"terms"`
	}{
		Terms: tm.terms,
	}

	// set all terms to keep by default
	for _, term := range tm.terms {
		term.Keep = true
	}

	// marshal to JSON
	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling terms: %w", err)
	}

	// create temporary file
	tmpFile, err := os.CreateTemp("", "terms-*.json")
	if err != nil {
		return fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// write JSON to file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("error writing temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %w", err)
	}

	// open in editor
	editor := getDefaultEditor()
	cmd := exec.Command(editor, filepath.Clean(tmpFile.Name())) // #nosec G204
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Opening terms in %s. Save and exit the editor when finished.\n", editor)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running editor: %w", err)
	}

	// read edited file
	editedData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("error reading edited file: %w", err)
	}

	// unmarshal edited data
	var editedJSON struct {
		Terms []*Term `json:"terms"`
	}
	if err := json.Unmarshal(editedData, &editedJSON); err != nil {
		return fmt.Errorf("error parsing edited file: %w", err)
	}

	// update terms
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
		// try to use environment variables for Linux
		if editor := os.Getenv("EDITOR"); editor != "" {
			return editor
		}
		if editor := os.Getenv("VISUAL"); editor != "" {
			return editor
		}
		// default to nano as it's commonly available
		return "nano"
	}
}
