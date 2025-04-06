package terms

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTermManager_AddTerms(t *testing.T) {
	tm := New()
	terms := []*Term{
		{Term: "term1", Description: "desc1", Keep: true},
		{Term: "term2", Description: "desc2", Keep: false},
	}

	tm.AddTerms(terms)
	require.Equal(t, terms, tm.GetAllTerms())
}

func TestTermManager_GetUntranslatableTerms(t *testing.T) {
	tm := New()
	terms := []*Term{
		{Term: "term1", Keep: true},
		{Term: "term2", Keep: false},
		{Term: "term3", Keep: true},
	}

	tm.AddTerms(terms)
	result := tm.GetUntranslatableTerms()
	require.Equal(t, []string{"term1", "term3"}, result)
}

func TestTermManager_ProcessTermsInteractive_AcceptAll(t *testing.T) {
	tm := New()
	terms := []*Term{
		{Term: "term1", Description: "desc1"},
		{Term: "term2", Description: "desc2"},
	}
	tm.AddTerms(terms)

	// Mock user input
	tm.SetInput(strings.NewReader("1\n"))

	err := tm.ProcessTermsInteractive()
	require.NoError(t, err)

	for _, term := range tm.GetAllTerms() {
		require.True(t, term.Keep)
	}
}

func TestTermManager_ProcessTermsInteractive_RejectAll(t *testing.T) {
	tm := New()
	terms := []*Term{
		{Term: "term1", Description: "desc1"},
		{Term: "term2", Description: "desc2"},
	}
	tm.AddTerms(terms)

	// Mock user input
	tm.SetInput(strings.NewReader("2\n"))

	err := tm.ProcessTermsInteractive()
	require.NoError(t, err)

	for _, term := range tm.GetAllTerms() {
		require.False(t, term.Keep)
	}
}

func TestTermManager_ProcessTermsInteractive_InvalidChoice(t *testing.T) {
	tm := New()
	terms := []*Term{
		{Term: "term1", Description: "desc1"},
	}
	tm.AddTerms(terms)

	// Mock invalid input
	tm.SetInput(strings.NewReader("5\n"))

	err := tm.ProcessTermsInteractive()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid choice")
}
