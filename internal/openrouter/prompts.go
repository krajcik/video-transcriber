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
Translate the following text from %s to %s. Make the translation as concise as possible by removing:
- Unnecessary words and phrases
- Redundant explanations
- Entire sentences that do not convey key information

Preserve the following terms untranslated:
%s

Text to translate:
%s

Instructions:
1. Remove filler and redundant content.
2. Delete sentences that do not add value.
3. Keep technical terms and essential facts.
4. Make the translation as short as possible while preserving core meaning.
5. Return only the translated text, without any comments or formatting outside the translation.
`
)
