package tts

import ()

var punctuations = map[string]string{
	".":  "",
	",":  "",
	"。":  "",
	"，":  "",
	"~":  "",
	"～":  "",
	"！":  "",
	"!":  "",
	"？":  "",
	"?":  "",
	";":  "",
	"；":  "",
	":":  "",
	"：":  "",
	"……": "",
}

func SplitSentence(content string, sentenceLen int) []string {
	var sentences []string
	prePos := 0
	for pos, c := range content {
		char := string(c)
		_, ok := punctuations[char]
		if ok {
			endPos := pos + len(char)
			newSen := content[prePos:endPos]
			prePos = endPos
			sentences = addSentence(sentences, newSen, sentenceLen)
		}
	}

	if prePos < len(content)-1 {
		newSen := content[prePos:]
		sentences = addSentence(sentences, newSen, sentenceLen)
	}
	return sentences

}

func addSentence(sentences []string, newSentence string, sentenceLen int) []string {
	if len(sentences) == 0 {
		sentences = append(sentences, newSentence)
	} else {
		lastIndex := len(sentences) - 1
		lastSen := sentences[lastIndex]
		if len(lastSen)+len(newSentence) > sentenceLen {
			sentences = append(sentences, newSentence)
		} else {
			sentences[lastIndex] = lastSen + newSentence
		}
	}
	return sentences
}
