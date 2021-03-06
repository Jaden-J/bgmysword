package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	//"github.com/mattn/go-sqlite3"
	//"github.com/fatih/color"
)

type TextContent []string

type VerseData struct {
	Number         int
	HtmlClass      string
	ContainsPoetry bool
	Content        TextContent
}

type ChapterData struct {
	Doc        *goquery.Document
	HtmlClass  string
	VerseIndex int
	Verses     []string
	Footnotes  map[string]string
}

const (
	PARAGRAPH            = "paragraph"
	TITLE                = "title"
	POETRY_LINE_NORMAL   = "line"
	POETRY_LINE_INDENTED = "indent-1"
	INDENT_1             = "<PI1>"
	INDENT_2             = "<PI2>"
	INDENT_3             = "<PI3>"
	INDENT_CLOSE         = "<CI>"
	PARAGRAPH_BREAK      = "<CM>"
	TITLE_OPEN           = "<TS>"
	TITLE_CLOSE          = "<Ts>"
	WOJ_OPEN             = "<FR>"
	WOJ_CLOSE            = "<Fr>"
	FOOTNOTE_OPEN        = "<RF>"
	FOOTNOTE_CLOSE       = "<Rf>"
)

var (
	Chapter ChapterData
	Verse   VerseData
	Text    TextContent
)

func GenDoc(url string) (doc *goquery.Document) {
	var err error
	var tried bool
	doc, err = goquery.NewDocument(url)
	if err != nil {
		if tried {
			log.Fatal(err)
		} else {
			tried = true
			doc = GenDoc(url)
			return
		}
	}
	return
}

func CopyrightFetch() {
	var copyrightInfo, publisherInfo string
	url := GenFullUrl("Genesis", "1")
	copyrightDoc := GenDoc(url)

	copyrightInfo = copyrightDoc.Find(".publisher-info-bottom p").Text()
	publisherInfo = copyrightDoc.Find(".publisher-info-bottom p a").Text()
	transName = copyrightDoc.Find(".publisher-info-bottom strong").Text()

	copyrightAccept(copyrightInfo, publisherInfo)
}

func (Chapter *ChapterData) Parse(url string) (chapterText []string) {
	// After formatting the current chapter,
	//    and returning its formatted text to function 'ChapterLoop',
	//        remove the chapter's data from memory
	defer Chapter.Clear()

	// Fetch the webpage for the current chapter.
	Chapter.GenDoc(url)

	// Filter out unneeded content.
	Chapter.CleanDoc()

	// Find the number of verses and the html class for the current chapter.
	Chapter.GenHtmlClass()

	// Find the chapter's footnotes,
	//    and match each footnote's letter with its expanded text.
	Chapter.GenFootnotes()

	// Mark titles and paragraphs.
	Chapter.TagDoc()

	// Parse the webpage.
	Chapter.AddVerses()

	chapterText = Chapter.Verses
	return
}

func (Chapter *ChapterData) Clear() {
	*Chapter = ChapterData{}
}

func (Chapter *ChapterData) GenDoc(url string) {
	Chapter.Doc = GenDoc(url)
}

func (Chapter *ChapterData) CleanDoc() {
	// Remove chapter numbers, verse numbers, and cross references
	//    (because MySword formats them differently/automatically).
	rejects := []string{".chapternum", ".versenum", ".crossreference"}

	for _, reject := range rejects {
		Chapter.Doc.Find(reject).Remove()
	}
}

func (Chapter *ChapterData) GenHtmlClass() {
	// Note the number of verses and the html class which denotes passage text.
	lastClass, _ := Chapter.Doc.Find("p .text").Last().Attr("class")

	doubleClass := prefixHyphenSecond(lastClass)
	chapterClass := suffixSpace(doubleClass)

	lastVerse := suffixHyphenSecond(lastClass)
	totalVerses, _ := strconv.Atoi(lastVerse)

	Chapter.HtmlClass = chapterClass
	Chapter.VerseIndex = totalVerses
}

func (Chapter *ChapterData) GenFootnotes() {
	// Make a map of the chapter's footnotes
	//     with their letters as keys and their expanded text as values.
	Chapter.Footnotes = make(map[string]string)

	footnoteTags := Chapter.Doc.Find("ol")
	containsFootnotes := IsNotEmpty(footnoteTags.Text())

	if containsFootnotes {
		footnoteTags.Contents().Each(func(i int, s *goquery.Selection) {
			id, _ := s.Attr("id")
			isFootnote := strings.Contains(id, "fen")

			if isFootnote {
				footnoteLetter := lastLetter(id)
				footnoteText := s.Find(".footnote-text").Text()

				Chapter.Footnotes[footnoteLetter] = footnoteText
			}
		})
	}
}

func (Chapter *ChapterData) TagDoc() {
	// Mark titles.
	Chapter.Doc.Find("h3 .text").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("id", "title")
	})
	Chapter.Doc.Find("h4 .text").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("id", "title")
	})

	// Mark the opening line of each paragraph.
	Chapter.Doc.Find("p").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Contents().Attr("class")
		isPassageContent := strings.Contains(class, Chapter.HtmlClass)

		if isPassageContent {
			s.Contents().First().SetAttr("id", "paragraph")
		}

	})
}

func (Chapter *ChapterData) AddVerses() {
	lastVerse := Chapter.VerseIndex

	for verse := 1; verse <= lastVerse; verse++ {
		Chapter.AddVerse(verse)
	}
}

func (Chapter *ChapterData) AddVerse(verseNumber int) {
	// After formatting the current verse,
	//    and adding its text to the chapter's data,
	//        remove the verse's data from memory.
	defer Verse.Clear()

	// Generate the html class for the current verse.
	Verse.GenHtmlClass(verseNumber)

	// Parse and format the text of the current verse.
	Verse.GenContent()

	// Clean the indentation for poetry lines.
	Verse.CleanIndent()

	// Add the newly formatted text for the current verse
	//    to the chapter's list of formatted verses.
	Chapter.Append(Verse.String())
}

func (Chapter *ChapterData) Append(verseText string) {
	Chapter.Verses = append(Chapter.Verses, verseText)
}

func (Verse *VerseData) Clear() {
	*Verse = VerseData{}
}

func (Verse *VerseData) String() string {
	return str(Verse.Content)
}

func (Verse *VerseData) Append(args ...string) {
	Verse.Content = append(Verse.Content, args...)
}

func (Verse *VerseData) GenHtmlClass(verseNumber int) {
	verseNumAsString := strconv.Itoa(verseNumber)

	Verse.HtmlClass = concat("text ", Chapter.HtmlClass, verseNumAsString)

}

func (Verse *VerseData) GenContent() {
	Chapter.Doc.Find(".text").Each(func(i int, s *goquery.Selection) {
		currClass, _ := s.Attr("class")

		if currClass == Verse.HtmlClass {
			// If text belongs in the current verse:
			Verse.AnalyseGenre(s)
		}

	})
}

func (Verse *VerseData) AnalyseGenre(s *goquery.Selection) {
	var classParents string
	parents := s.Parents()
	classParents, _ = parents.Attr("class")

	// Workaround for the first verse in the chapter.
	if strings.Contains(classParents, "chapter") {
		parents := s.Parents().Parents()
		classParents, _ = parents.Attr("class")
	}

	switch classParents {
	case POETRY_LINE_NORMAL:
		// If first-line poetry:
		Verse.ContainsPoetry = true
		Verse.ParsePoetryLine(s)
		return
	case POETRY_LINE_INDENTED:
		// If second-line/indented poetry:
		Verse.ParsePoetryIndent(s)
		return
	default:
		// If normal text:
		Verse.ParseProse(s)
		return
	}
}

func (Verse *VerseData) ParseProse(s *goquery.Selection) {
	id, _ := s.Attr("id")

	switch id {
	case "title":
		// If this html id tag was labeled "title" earlier:
		Verse.Title(s)
		return
	case "paragraph":
		// If this html id tag was labeled "paragraph" earlier:
		Verse.Paragraph()
		fallthrough
	default:
		// If normal content:
		Verse.ParseContents(s.Contents())
		return
	}
}

func (Verse *VerseData) ParsePoetryLine(s *goquery.Selection) {
	id, _ := s.Attr("id")

	switch id {
	// If this html id tag was labeled "title" earlier:
	case "title":
		Verse.Title(s)
		return
	// If this html id tag was labeled "paragraph" earlier:
	case "paragraph":
		Verse.Paragraph()
		fallthrough
	// If normal content:
	default:
		Verse.Append(INDENT_1)            // Tag to single-indent
		Verse.ParseContents(s.Contents()) //     Indented text
		Verse.Append(INDENT_CLOSE)        // Tag to close indent
		return
	}
}

func (Verse *VerseData) ParsePoetryIndent(s *goquery.Selection) {
	id, _ := s.Contents().Attr("id")

	switch id {
	// If this html id tag was labeled "title" earlier:
	case "title":
		Verse.Title(s)
		return
	// If this html id tag was labeled "paragraph" earlier:
	case "paragraph":
		Verse.Paragraph()
		fallthrough
	// If normal content:
	default:
		Verse.Append(INDENT_3)            // Tag to triple-indent
		Verse.ParseContents(s.Contents()) //     Indented text
		Verse.Append(INDENT_CLOSE)        // Tag to close indent
		return

	}
}

func (Verse *VerseData) Title(s *goquery.Selection) {
	Verse.Append(TITLE_OPEN)  // Opening title tag
	Verse.ParseContents(s)    //    Title text
	Verse.Append(TITLE_CLOSE) // Closing title tag
}

func (Verse *VerseData) Paragraph() {
	contentLength := len(Verse.Content)

	switch contentLength {
	case 0:
		// If the new paragraph starts at the begininning of the verse:
		Chapter.AddParagraph()
		return
	default:
		// If the new paragraph starts in the middle of the verse:
		Verse.AddParagraph()
		return
	}
}

func (Chapter *ChapterData) AddParagraph() {
	var versesAdded int
	versesAdded = len(Chapter.Verses)

	switch versesAdded {
	case 0:
		// If this is the first verse in the chapter,
		//    do nothing.
		return
	default:
		// Otherwise,
		//    move the paragraph tag to the very end of the previous verse.
		lastIndex := versesAdded - 1
		lastVerse := Chapter.Verses[lastIndex]
		prevVerses := Chapter.Verses[:lastIndex]
		lastVerseNew := concat(lastVerse, PARAGRAPH_BREAK)
		temp := append(prevVerses, lastVerseNew)

		Chapter.Verses = nil
		Chapter.Verses = temp
		return
	}
}

func (Verse *VerseData) AddParagraph() {
	contentLength := len(Verse.Content)
	lastIndex := contentLength - 1
	prevContent := Verse.Content[lastIndex]

	switch prevContent {
	case TITLE_CLOSE:
		// If the new paragraph DOES begin after a title,
		//    the paragraph tag is NOT unneeded.
		return
	default:
		// If the new paragraph does NOT begin after a title,
		//    a paragraph tag IS needed.
		Verse.Append(PARAGRAPH_BREAK)
		return
	}
}

func (Verse *VerseData) ParseContents(sel *goquery.Selection) {
	defer Text.Clear()
	for i := range sel.Nodes {
		// For each part of the current verse's text in the html Document:
		s := sel.Eq(i)
		Text.Analyse(s)
	}

	Verse.Append(Text.String())
}

func (Text *TextContent) Clear() {
	*Text = TextContent{}
}

func (Text *TextContent) Analyse(s *goquery.Selection) {
	class, _ := s.Attr("class")
	woj := strings.Contains(class, "woj")              // Words of Jesus
	footnote := strings.Contains(class, "footnote")    // Footnote
	smallCaps := strings.Contains(class, "small-caps") // Small-caps (e.g., "LORD")

	switch {
	case woj:
		Text.WoJ(s.Contents())
	case footnote:
		Text.Footnote(s)
	case smallCaps:
		Text.SmallCaps(s)
	default:
		Text.Append(s.Text())
	}
}

func (Text *TextContent) String() string {
	return str(*Text)
}

func (Text *TextContent) Append(args ...string) {
	*Text = append(*Text, args...)
}

func (Text *TextContent) WoJ(sel *goquery.Selection) {
	// Opening tag for words of Jesus
	Text.Append(WOJ_OPEN)

	//    Passage text for words of Jesus
	for i := range sel.Nodes {
		// For each html segment of Jesus's words:
		s := sel.Eq(i)
		Text.Analyse(s)
	}

	// Closing tag for words of Jesus
	Text.Append(WOJ_CLOSE)
}

func (Text *TextContent) Footnote(s *goquery.Selection) {
	fnLetter := s.Find("a").Text()
	fnText := Chapter.Footnotes[fnLetter] // Footnote text that was marked earlier.

	Text.Append(FOOTNOTE_OPEN, fnText, FOOTNOTE_CLOSE)
}

func (Text *TextContent) SmallCaps(s *goquery.Selection) {
	Text.Append(strings.ToUpper(s.Text()))
}

func (Verse *VerseData) CleanIndent() {
	if Verse.ContainsPoetry {
		Verse.CleanIndentTags()
	}
}

func (Verse *VerseData) CleanIndentTags() {
	var temp TextContent
	var isFirstInVerse bool = true
	for _, s := range Verse.Content {
		isPoetryLine := strings.Contains(s, INDENT_1)

		switch isPoetryLine {
		case true:

			switch isFirstInVerse {
			case true:
				//If line IS the first one in the verse,
				//    keep it single-indented.
				isFirstInVerse = false
				temp.Append(s)
			case false:
				//If line is NOT the first one in the verse,
				//    double its indentation.
				newLine := strings.Replace(s, INDENT_1, INDENT_2, -1)
				temp.Append(newLine)
			}

		default:
			temp.Append(s)
		}

	}

	Verse.Content = nil
	Verse.Content = temp
}
