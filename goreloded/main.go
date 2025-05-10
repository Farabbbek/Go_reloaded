package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type TextProcessor struct{}

func (tp *TextProcessor) Process(text string) string {
	text = tp.processMultipleTypes(text)
	text = tp.processConversion(text)
	text = tp.processConversioncount(text)
	text = tp.processHexNumbers(text)
	text = tp.processBinaryNumbers(text)

	text = tp.processCaseModifiers(text)
	text = tp.processMultiModifiers(text)
	text = tp.processNestedModifiers(text)
	text = tp.processPunctuation(text)
	text = tp.processQuotes(text)
	text = tp.processArticles(text)
	text = tp.processStandaloneModifiers(text)

	return text
}

func (tp *TextProcessor) processConversion(text string) string {
	re := regexp.MustCompile(`([0-9A-Fa-f]+)\s*\((\w+)\)\s*\((\w+)\)`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		num := parts[1]
		firstOp := parts[2]
		secondOp := parts[3]

		// First conversion
		var val int64
		var err error
		if firstOp == "bin" {
			val, err = strconv.ParseInt(num, 2, 64)
		} else if firstOp == "hex" {
			val, err = strconv.ParseInt(num, 16, 64)
		}
		if err != nil {
			return match
		}

		// Second conversion
		if secondOp == "bin" {
			return fmt.Sprintf("%b", val)
		} else if secondOp == "hex" {
			if firstOp == "bin" {
				return fmt.Sprintf("%X", val)
			}
			return fmt.Sprintf("%d", val)
		}

		return match
	})
}

func (tp *TextProcessor) processConversioncount(text string) string {
	// Handle binary conversion
	reBin := regexp.MustCompile(`\(bin, (-?\d+)\)`)
	text = reBin.ReplaceAllStringFunc(text, func(match string) string {
		parts := reBin.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		number, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return match
		}
		return fmt.Sprintf("%b", number)
	})

	// Handle hexadecimal conversion
	reHex := regexp.MustCompile(`\(hex, (-?\d+)\)`)
	text = reHex.ReplaceAllStringFunc(text, func(match string) string {
		parts := reHex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		number, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return match
		}
		return fmt.Sprintf("%X", number)
	})

	return text
}

// // binary
func (tp *TextProcessor) processBinaryNumbers(text string) string {
	prevText := ""
	for prevText != text {
		prevText = text
		reBinNum := regexp.MustCompile(`(\d+)\s*\(\s*bin\s*\)`)
		text = reBinNum.ReplaceAllStringFunc(text, func(match string) string {
			parts := reBinNum.FindStringSubmatch(match)
			bin := parts[1]
			decimal, err := strconv.ParseInt(bin, 2, 64)
			if err != nil {
				return match
			}
			return fmt.Sprintf("%d", decimal)
		})
	}
	reStandalone := regexp.MustCompile(`\(\s*bin\s*\)`)
	text = reStandalone.ReplaceAllString(text, "")

	reWord := regexp.MustCompile(`(\w+)\s*\(\s*bin\s*\)`)
	return reWord.ReplaceAllString(text, "$1")
}

// // HEX
func (tp *TextProcessor) processHexNumbers(text string) string {
	prevText := ""
	for prevText != text {
		prevText = text
		reHexNum := regexp.MustCompile(`([0-9A-Fa-f]+)\s*\(\s*hex\s*\)`)
		text = reHexNum.ReplaceAllStringFunc(text, func(match string) string {
			parts := reHexNum.FindStringSubmatch(match)
			hex := parts[1]
			decimal, err := strconv.ParseInt(hex, 16, 64)
			if err != nil {
				return hex
			}
			return fmt.Sprintf("%d", decimal)
		})
	}

	reStandalone := regexp.MustCompile(`\(\s*hex\s*\)`)
	text = reStandalone.ReplaceAllString(text, "")

	reWord := regexp.MustCompile(`(\w+)\s*\(\s*hex\s*\)`)
	return reWord.ReplaceAllString(text, "$1")
}

// alone delete up low cap
func (tp *TextProcessor) processStandaloneModifiers(text string) string {
	// Remove standalone modifiers with newlines
	text = regexp.MustCompile(`(?m)^\s*\(\s*(up|low|cap)\s*\)\s*$`).ReplaceAllString(text, "")

	// Remove standalone modifiers within text
	text = regexp.MustCompile(`\s*\(\s*(up|low|cap)\s*\)\s*`).ReplaceAllString(text, "")

	return text
}

func (tp *TextProcessor) processNestedModifiers(input string) string {
	// Regex to find innermost parentheses with modifiers and word count
	re := regexp.MustCompile(`\((\s*(cap|low|up)\s*,\s*(-?\d+)\s*([^\(\)]*)\s*)\)`)

	// Loop until there are no more parentheses
	for re.MatchString(input) {
		// Find the innermost match
		input = re.ReplaceAllStringFunc(input, func(match string) string {
			// Extract parts of the match
			parts := re.FindStringSubmatch(match)
			modifier := parts[2]                // Modifier: cap, low, or up
			count, _ := strconv.Atoi(parts[3])  // Word count
			text := strings.TrimSpace(parts[4]) // Text inside parentheses

			// Apply the modifier
			processedText := tp.applyModifier(text, modifier, count)
			return processedText // Replace the match with the processed text
		})
	}

	return input
}

// Helper function to apply a single modifier to a text
func (tp *TextProcessor) applyModifier(text, modifier string, count int) string {
	words := strings.Fields(text)

	// Determine the range of words to modify
	start, end := 0, len(words)
	if count > 0 {
		start = len(words) - count
		if start < 0 {
			start = 0
		}
	} else if count < 0 {
		end = -count
		if end > len(words) {
			end = len(words)
		}
	}

	// Apply the modifier to the selected range
	for i := start; i < end; i++ {
		switch modifier {
		case "cap":
			words[i] = strings.Title(strings.ToLower(words[i]))
		case "low":
			words[i] = strings.ToLower(words[i])
		case "up":
			words[i] = strings.ToUpper(words[i])
		}
	}

	return strings.Join(words, " ")
}

// cap up low and cap low up with count
func (tp *TextProcessor) processCaseModifiers(text string) string {
	if _, err := strconv.Atoi(text); err == nil {
		return text
	}
	re := regexp.MustCompile(`((?:\w+\s+){0,10}?\w+)\s*\(\s*(up|low|cap),\s*(\d+)\s*\)`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		words := strings.Fields(parts[1])
		count, _ := strconv.Atoi(parts[3])
		if count > len(words) {
			count = len(words)
		}
		start := len(words) - count
		for i := start; i < len(words); i++ {
			switch parts[2] {
			case "up":
				words[i] = strings.ToUpper(words[i])
			case "low":
				words[i] = strings.ToLower(words[i])
			case "cap":
				words[i] = strings.Title(strings.ToLower(words[i]))
			default:
				return match
			}
		}
		return strings.Join(words, " ")
	})
	re = regexp.MustCompile(`(\w+)\s*\(\s*(up|low|cap)\s*\)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		word := parts[1]
		switch parts[2] {
		case "up":
			return strings.ToUpper(word)
		case "low":
			return strings.ToLower(word)
		case "cap":
			return strings.Title(strings.ToLower(word))
		default:
			return word
		}
	})
}
func (tp *TextProcessor) processMultiModifiers(text string) string {
	reMulti := regexp.MustCompile(`(\w+)(?:\((up|low|cap)\))+(\s*)`)
	return reMulti.ReplaceAllStringFunc(text, func(match string) string {
		word := regexp.MustCompile(`(\w+)`).FindString(match)
		modifiers := regexp.MustCompile(`\((up|low|cap)\)`).FindAllStringSubmatch(match, -1)
		space := regexp.MustCompile(`(\s*)$`).FindString(match)
		result := word
		for _, mod := range modifiers {
			switch mod[1] {
			case "up":
				result = strings.ToUpper(result)
			case "low":
				result = strings.ToLower(result)
			case "cap":
				if len(result) > 0 {
					result = string(unicode.ToUpper(rune(result[0]))) + strings.ToLower(result[1:])
				}
			}
		}
		return result + space
	})
}

func (tp *TextProcessor) processMultipleTypes(text string) string {
	reMulti := regexp.MustCompile(`([0-9A-Fa-f]+)(?:\((up|low|cap|hex|bin)\))+`)
	return reMulti.ReplaceAllStringFunc(text, func(match string) string {
		value := regexp.MustCompile(`([0-9A-Fa-f]+)`).FindString(match)
		operations := regexp.MustCompile(`\((\w+)\)`).FindAllStringSubmatch(match, -1)

		// Track if conversion happened
		converted := false

		// First apply case modifiers
		for _, op := range operations {
			if op[1] == "cap" || op[1] == "low" || op[1] == "up" {
				switch op[1] {
				case "cap":
					value = strings.Title(strings.ToLower(value))
				case "low":
					value = strings.ToLower(value)
				case "up":
					value = strings.ToUpper(value)
				}
			}
		}

		// Then apply conversion
		for _, op := range operations {
			if !converted && (op[1] == "hex" || op[1] == "bin") {
				switch op[1] {
				case "hex":
					if val, err := strconv.ParseInt(value, 16, 64); err == nil {
						value = fmt.Sprintf("%d", val)
						converted = true
					}
				case "bin":
					if val, err := strconv.ParseInt(value, 2, 64); err == nil {
						value = fmt.Sprintf("%d", val)
						converted = true
					}
				}
			}
		}

		return value
	})
}

func (tp *TextProcessor) processPunctuation(text string) string {
	// Dots
	re := regexp.MustCompile(`\s*(\.*\.)\s*`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		dots := strings.ReplaceAll(match, " ", "")
		return dots + " "
	})
	// commas and semicolons and exclamation mark and question mark
	text = regexp.MustCompile(`(\S+)\s*(!+)`).ReplaceAllString(text, "$1$2")
	text = regexp.MustCompile(`(!+)\s+(\S+)`).ReplaceAllString(text, "$1 $2")

	text = regexp.MustCompile(`(\S+)\s*(\?+)`).ReplaceAllString(text, "$1$2")
	text = regexp.MustCompile(`(\?+)\s+(\S+)`).ReplaceAllString(text, "$1 $2")
	// Handle question marks and exclamation marks

	text = regexp.MustCompile(`\s*\!\?\s*`).ReplaceAllString(text, "!? ")
	text = regexp.MustCompile(`\s*\?\!\s*`).ReplaceAllString(text, "?! ")

	text = regexp.MustCompile(`(![\?\!]|[\?\!]!)\s*\.\s*`).ReplaceAllString(text, "$1.")

	re = regexp.MustCompile(`\s*([,;:])\s*`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		punct := strings.TrimSpace(match)
		return punct + " "
	})

	text = regexp.MustCompile(`([!?])\s*([!?])`).ReplaceAllString(text, "$1$2")
	text = regexp.MustCompile(`([!?])\s*$`).ReplaceAllString(text, "$1 ")

	text = regexp.MustCompile(`\s{2,}`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func (tp *TextProcessor) processQuotes(text string) string {
	//  single quotes
	singleQuoteRe := regexp.MustCompile(`'\s*([^"]+?)\s*'`)
	text = singleQuoteRe.ReplaceAllStringFunc(text, func(match string) string {
		parts := singleQuoteRe.FindStringSubmatch(match)
		innerText := strings.TrimSpace(parts[1])
		return "'" + innerText + "'"
	})

	//  double quotes
	doubleQuoteRe := regexp.MustCompile(`"\s*([^"]+?)\s*"`)
	text = doubleQuoteRe.ReplaceAllStringFunc(text, func(match string) string {
		parts := doubleQuoteRe.FindStringSubmatch(match)
		innerText := strings.TrimSpace(parts[1])
		return "\"" + innerText + "\""
	})
	return text //strings.TrimSpace(text)
}

// a and an
func (tp *TextProcessor) processArticles(text string) string {
	re := regexp.MustCompile(`\b([Aa][Nn]?)\s+(\w+)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := strings.Fields(match)
		article := parts[0]
		word := parts[1]

		// Check if word starts with vowel
		startsWithVowel := regexp.MustCompile(`^[aeiouAEIOU]`).MatchString(word)

		// Determine correct article based on case and vowel
		var newArticle string
		isUpper := article == strings.ToUpper(article)

		if startsWithVowel {
			if isUpper {
				newArticle = "AN"
			} else {
				newArticle = "an"
			}
		} else {
			if isUpper {
				newArticle = "A"
			} else {
				newArticle = "a"
			}
		}

		return newArticle + " " + word
	})
}

// Ensure file exists
func ensureFileExists(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func main() {
	// Get file names from command line args or use defaults
	inputFile := "sample.txt"
	outputFile := "result.txt"
	if len(os.Args) == 3 {
		inputFile = os.Args[1]
		outputFile = os.Args[2]
	}

	// Ensure files exist
	if err := ensureFileExists(inputFile); err != nil {
		fmt.Printf("Error creating %s: %v\n", inputFile, err)
		return
	}
	if err := ensureFileExists(outputFile); err != nil {
		fmt.Printf("Error creating %s: %v\n", outputFile, err)
		return
	}

	// Open input file
	input, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", inputFile, err)
		return
	}
	defer input.Close()

	// Open output file
	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", outputFile, err)
		return
	}
	defer output.Close()

	// Read and process input line-by-line
	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush() // Ensure all data is written at the end

	tp := &TextProcessor{}

	for scanner.Scan() {
		line := scanner.Text()         // Read one line
		line = strings.TrimSpace(line) // Remove leading/trailing spaces
		if line == "" {                // Skip empty lines
			continue
		}
		processedLine := tp.Process(line)                  // Process the line
		_, err := writer.WriteString(processedLine + "\n") // Write processed line with newline
		if err != nil {
			fmt.Printf("Error writing to file %s: %v\n", outputFile, err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", inputFile, err)
		return
	}

	fmt.Printf("Successfully processed %s and saved to %s\n", inputFile, outputFile)
}
