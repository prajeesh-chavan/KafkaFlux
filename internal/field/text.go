package field

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var loremWords = []string{
	"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing",
	"elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore",
	"et", "dolore", "magna", "aliqua", "enim", "ad", "minim", "veniam",
	"quis", "nostrud", "exercitation", "ullamco", "laboris", "nisi", "ut",
	"aliquip", "ex", "ea", "commodo", "consequat", "duis", "aute", "irure",
	"reprehenderit", "voluptate", "velit", "esse", "cillum", "eu", "fugiat",
	"nulla", "pariatur", "excepteur", "sint", "occaecat", "cupidatat", "non",
	"proident", "sunt", "culpa", "qui", "officia", "deserunt", "mollit",
}

var productAdjectives = []string{
	"Premium", "Eco", "Smart", "Ultra", "Pro", "Lite", "Advanced", "Classic",
}

var productNouns = []string{
	"Widget", "Gadget", "Device", "Tool", "Kit", "Pack", "Bundle", "Sensor",
}

func genWord() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return loremWords[r.Intn(len(loremWords))]
	}
}

func genSentence() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		n := r.Intn(10) + 5
		words := make([]string, n)
		for i := range words {
			words[i] = loremWords[r.Intn(len(loremWords))]
		}
		sentence := strings.Join(words, " ")
		sentence = strings.ToUpper(sentence[:1]) + sentence[1:] + "."
		return sentence
	}
}

func genParagraph() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		n := r.Intn(5) + 3
		sentences := make([]string, n)
		for i := range sentences {
			s := r.Intn(10) + 5
			words := make([]string, s)
			for j := range words {
				words[j] = loremWords[r.Intn(len(loremWords))]
			}
			sentences[i] = strings.ToUpper(words[0][:1]) + words[0][1:] + " " + strings.Join(words[1:], " ") + "."
		}
		return strings.Join(sentences, " ")
	}
}

func genProductName() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return productAdjectives[r.Intn(len(productAdjectives))] + " " +
			productNouns[r.Intn(len(productNouns))] + " v" +
			fmt.Sprintf("%d.%d", r.Intn(5)+1, r.Intn(10))
	}
}

func genSKU() FieldGen {
	categories := []string{"ELEC", "HOME", "CLTH", "FOOD", "TOYS", "MED", "OFFC"}
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		return fmt.Sprintf("%s-%04d-%s",
			categories[r.Intn(len(categories))],
			r.Intn(10000),
			strings.ToUpper(string(loremWords[r.Intn(len(loremWords))][:3])),
		)
	}
}

func genPastTimestamp() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		offset := time.Duration(r.Intn(365*24*3600)) * time.Second
		return time.Now().Add(-offset).Format(time.RFC3339)
	}
}

func genFutureTimestamp() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		offset := time.Duration(r.Intn(365*24*3600)) * time.Second
		return time.Now().Add(offset).Format(time.RFC3339)
	}
}

func genDate() FieldGen {
	return func(r *rand.Rand, _ map[string]interface{}) interface{} {
		offset := time.Duration(r.Intn(365*24*3600)) * time.Second
		return time.Now().Add(-offset).Format("2006-01-02")
	}
}
