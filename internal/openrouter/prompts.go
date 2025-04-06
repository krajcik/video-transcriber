package openrouter

const (
	analyzeTermsPrompt = `
Analyze the following text and identify special terms 
that should not be translated from English to Russian. For each term, provide 
a description and context (sentences where the term appears).

Text:
%s

Return your answer strictly in the following JSON format:
{
  "terms": [
    {
      "term": "grip",
      "description": "traction between tires and road surface",
      "context": [
        "The car had excellent grip on the wet track.",
        "Maintaining grip is crucial during high-speed cornering."
      ]
    },
    {
      "term": "rotation",
      "description": "car rotating around its vertical axis",
      "context": [
        "The driver initiated rotation by applying the correct amount of steering input."
      ]
    }
  ]
}
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
