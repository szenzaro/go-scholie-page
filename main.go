package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"
)

// HighlightedText contains informations about the scholie to be highlighted
type HighlightedText struct {
	Highlighted bool
	Text        string
}

type entry struct {
	homer   string
	scholie []HighlightedText
}

type exportScholie struct {
	Chant   int
	Verse   int
	Homer   string
	Scholie []interface{}
}

func main() {
	scholiePath := "data/Scholies_Paraphrase_1_3_corr.xlsx"
	coloredScholiePath := "data/Scholies_Paraphrase_1_3.xlsx"

	scholie, err := xlsx.OpenFile(scholiePath)
	if err != nil {
		log.Fatalln(err)
	}

	scholieData := map[string]map[string]map[string][]string{} // chant -> verse -> []entry

	for _, sheet := range scholie.Sheets {
		for i, row := range sheet.Rows {
			if i == 0 {
				continue
			}

			scholie := []string{}
			for _, v := range row.Cells[4:] {
				s := strings.TrimSpace(v.Value)
				if s != "" {
					scholie = append(scholie, strings.TrimSpace(v.Value))
				}
			}

			chant := strings.TrimSpace(row.Cells[0].Value)
			verse := strings.TrimSpace(row.Cells[1].Value)
			homer := strings.TrimSpace(row.Cells[2].Value)

			if _, ok := scholieData[chant]; !ok {
				scholieData[chant] = map[string]map[string][]string{}
			}
			if _, ok := scholieData[chant][verse]; !ok {
				scholieData[chant][verse] = map[string][]string{}
			}
			scholieData[chant][verse][homer] = scholie
		}
	}

	colored, err := xlsx.OpenFile(coloredScholiePath)
	if err != nil {
		log.Fatalln(err)
	}

	coloredData := []exportScholie{} // chant -> verse -> homer -> []highilightdata

	for _, sheet := range colored.Sheets {
		for i, row := range sheet.Rows {
			if i == 0 {
				continue
			}
			chant := strings.TrimSpace(row.Cells[0].Value)
			verse := strings.TrimSpace(row.Cells[1].Value)
			homer := strings.TrimSpace(row.Cells[2].Value)
			ss := []HighlightedText{{false, strings.TrimSpace(row.Cells[3].Value)}}

			for _, k := range scholieData[chant][verse][homer] {
				ss, _ = splitHighlight(ss, k)
			}
			f := []interface{}{}
			for _, hd := range ss {
				f = append(f, []interface{}{hd.Highlighted, hd.Text})
			}
			cn, err := strconv.Atoi(chant)
			if err != nil {
				log.Fatalln(err)
			}

			vn, err := strconv.Atoi(verse)
			if err != nil {
				log.Fatalln(err)
			}
			coloredData = append(coloredData, exportScholie{
				Chant:   cn,
				Verse:   vn,
				Homer:   homer,
				Scholie: f,
			})
		}
	}
	writeToJSON("out", "out/scholie-page.json", coloredDataToJSON(coloredData))
}

func coloredDataToJSON(d []exportScholie) []interface{} {
	data := []interface{}{}

	for _, e := range d {
		data = append(data, []interface{}{e.Chant, e.Verse, e.Homer, e.Scholie})
	}
	// sort.Slice(data, func(i, j int) bool { return data[i].Chant < data[j].Chant || data[i].Verse < data[j].Verse })
	return data
}

func writeToJSON(folder, path string, data interface{}) {
	fmt.Println("Savind JSON to ", path)

	if err := os.MkdirAll(folder, 0777); err != nil {
		log.Fatalln(err)
	}

	d, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}
	err = ioutil.WriteFile(path, d, 0664)
	if err != nil {
		log.Fatalln(err)
	}
}

func split(s, scholia string) []HighlightedText {
	data := []HighlightedText{}
	sch := scholia // normalizeText(scholia)
	// text := strings.ToLower(normalizeText(s))
	if strings.Contains(s, "´") && (strings.HasSuffix(sch, "α") ||
		strings.HasSuffix(sch, "ε") ||
		strings.HasSuffix(sch, "ι") ||
		strings.HasSuffix(sch, "ο")) { // There is an elision§
		sch = scholia[:len(scholia)-2]
	}

	idx := strings.Index(strings.ToLower(s), strings.ToLower(sch))
	if idx != -1 {
		leftText := s[:idx]
		if len(leftText) > 0 {
			data = append(data, HighlightedText{Text: leftText, Highlighted: false})
		}
		middleText := s[idx : idx+len(sch)]
		data = append(data, HighlightedText{Text: middleText, Highlighted: true})
		rightText := s[idx+len(sch):]
		if len(rightText) > 0 {
			data = append(data, HighlightedText{Text: rightText, Highlighted: false})
		}
	}

	return data
}

func splitHighlight(s []HighlightedText, scholia string) ([]HighlightedText, bool) {
	data := append(s[:0:0], s...)
	foundSplit := false
	for i := 0; i < len(data); i++ {
		if !data[i].Highlighted {
			// check if can be splitted
			splitted := split(data[i].Text, scholia)
			if len(splitted) > 0 {
				before := append(data[:0:0], data[:i]...)
				after := append(data[:0:0], data[i+1:]...)
				data = before
				data = append(data, splitted...)
				data = append(data, after...)
				i = 0
				foundSplit = true
			}
		}
	}
	return data, foundSplit
}
