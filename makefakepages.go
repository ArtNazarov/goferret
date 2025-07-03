package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	nPages      = 100000
	dirContent  = "content"

	templateVal = "blog"
	categoryVal = "main"
)

var (
	titleWords   = []string{"Amazing", "Quick", "Lazy", "Bright", "Silent", "Loud", "Happy", "Sad", "Clever", "Brave", "Wild", "Calm", "Funky", "Cosmic", "Magic", "Lucky", "Epic", "Tiny", "Giant", "Fresh"}
	contentWords = []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit", "sed", "do", "eiusmod", "tempor", "incididunt", "ut", "labore", "et", "dolore", "magna", "aliqua", "ut", "enim", "ad", "minim", "veniam"}
)

func randomID(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomTitle() string {
	w1 := titleWords[rand.Intn(len(titleWords))]
	w2 := titleWords[rand.Intn(len(titleWords))]
	return fmt.Sprintf("%s %s Page", w1, w2)
}

func randomContent() string {
	words := make([]string, rand.Intn(20)+10)
	for i := range words {
		words[i] = contentWords[rand.Intn(len(contentWords))]
	}
	return strings.Title(strings.Join(words, " ")) + "."
}

func main() {
	rand.Seed(time.Now().UnixNano())
	if _, err := os.Stat(dirContent); os.IsNotExist(err) {
		if err := os.Mkdir(dirContent, 0755); err != nil {
			panic(err)
		}
	}
	for i := 0; i < nPages; i++ {
		pageID := randomID(8)
		pageDir := filepath.Join(dirContent, pageID)
		if err := os.Mkdir(pageDir, 0755); err != nil {
			fmt.Printf("Ошибка при создании директории %s: %v\n", pageDir, err)
			continue
		}
		title := randomTitle()
		content := randomContent()
		if err := os.WriteFile(filepath.Join(pageDir, "title.val"), []byte(title), 0644); err != nil {
			fmt.Printf("Ошибка при записи title.val: %v\n", err)
		}
		if err := os.WriteFile(filepath.Join(pageDir, "content.val"), []byte(content), 0644); err != nil {
			fmt.Printf("Ошибка при записи content.val: %v\n", err)
		}
		if err := os.WriteFile(filepath.Join(pageDir, "template.setting"), []byte(templateVal), 0644); err != nil {
			fmt.Printf("Ошибка при записи template.setting: %v\n", err)
		}
		if err := os.WriteFile(filepath.Join(pageDir, "category.val"), []byte(categoryVal), 0644); err != nil {
			fmt.Printf("Ошибка при записи category.val: %v\n", err)
		}
	}
	fmt.Printf("Сгенерировано %d страниц в директории %s\n", nPages, dirContent)
}
