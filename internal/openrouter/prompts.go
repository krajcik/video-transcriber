package openrouter

const (
	analyzeTermsPrompt = `
Extract up to 15 key terms from the text that must remain untranslated. Focus on technical terms, proper nouns, and domain-specific jargon.

Instructions:
1. Only include terms essential for correct translation.
2. Group similar forms (e.g., singular/plural).
3. Exclude common nouns and generic words.
4. Never exceed 15 terms.
5. Always return a valid JSON array, even if empty.
6. Do not add any comments or explanations outside the JSON.

Response format:
{
  "terms": [
    {
      "term": "original_term",
      "note": "short context hint",
      "category": "technical|name|acronym|unit"
    }
  ]
}

Examples:
{
  "terms": [
    {
      "term": "ABS",
      "note": "anti-lock braking system",
      "category": "acronym"
    },
    {
      "term": "Nürburgring",
      "note": "famous race track",
      "category": "name"
    }
  ]
}

Edge-case (no terms found):
{
  "terms": []
}

Text:
%s
`

	translateTextPrompt = `
Translate the following text from %s to %s and format the result as a well-structured markdown article.

Requirements:
1. Remove unnecessary words, redundant explanations and non-essential sentences
2. Keep technical terms and key facts
3. Preserve these terms untranslated:
%s
4. Format the translation as markdown with:
	  - Clear section headings (##, ###)
	  - Bullet points for lists
	  - Bold/italic for emphasis
	  - Code blocks for technical terms
5. Make the translation concise but readable

Example markdown structure:
## Main Topic
Key points about the topic.

### Subsection
- Important detail 1
- Important detail 2

Use 'Term' for technical terms.

Text to translate:
%s

Return only the formatted markdown without additional comments.
`
)
