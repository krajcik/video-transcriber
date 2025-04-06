package openrouter

const (
	analyzeTermsPrompt = `
Analyze the text and extract MAX 15 most important special terms that MUST remain untranslated. Prioritize technical terms, proper nouns, and domain-specific jargon.

Rules:
1. Include only terms critical for accurate translation
2. Group similar terms (e.g. plural forms)
3. Avoid obvious common nouns
4. Never exceed 15 terms

Format response as:
{
  "terms": [
    {
      "term": "original_term",
      "note": "short context hint", 
      "category": "technical|name|acronym|unit"
    }
  ]
}

Example:
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

Text:
%s
`

	translateTextPrompt = `
Translate the following text from English to Russian in a very concise way, aggressively removing:
- Unnecessary words and phrases
- Redundant explanations
- Entire sentences that don't carry key information
Keep only the most important facts and actions. Preserve the following terms untranslated:
%s

Text to translate:
%s

Rules:
1. Remove filler words and fluff
2. Delete whole sentences if they don't add value
3. Keep technical terms and key facts
4. Make the text as short as possible while keeping core meaning
5. Return only the cleaned translation without comments
`
)
