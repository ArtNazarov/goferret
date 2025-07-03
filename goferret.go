/*
Author: Артем Назаров
Email: programmist.nazarov@gmail.com
Created: 2025-07-02
Description: Генератор статических сайтов на основе шаблонов и атрибутов.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"encoding/json"
	"sync"
	"time"
)

// Сообщения для пользователя
const (
	msgTemplatesDirNotFound = "Ошибка: директория 'templates' не найдена"
	msgContentDirNotFound   = "Ошибка: директория 'content' не найдена"
	msgErrorReadingContent  = "Ошибка при чтении директории content: %v\n"
	msgErrorProcessingPage  = "Ошибка при обработке страницы %s: %v\n"
	msgWarningNoTemplate    = "Предупреждение: для страницы %s не указан шаблон\n"
	msgErrorLoadingTemplate = "Ошибка при загрузке шаблона для страницы %s: %v\n"
	msgErrorRendering       = "Ошибка при рендеринге шаблона для страницы %s: %v\n"
	msgErrorWritingOutput   = "Ошибка при записи вывода для страницы %s: %v\n"
	msgGenerated            = "Сгенерировано: %s\n"
	msgSiteGenerationDone   = "Генерация сайта завершена!"
	msgErrorReadingTemplate = "Ошибка при чтении шаблона %s: %v"
	msgErrorReadingSetting  = "Ошибка при чтении файла template.setting для %s: %v"
	msgErrorReadingPageDir  = "Ошибка при чтении директории страницы %s: %v"
	msgErrorReadingAttr     = "Ошибка при чтении атрибута %s для страницы %s: %v"
	msgErrorReadingCategory = "Ошибка при чтении категории для страницы %s: %v"
)

// Model представляет страницу с её атрибутами и шаблоном
type Model struct {
	ID       string
	Data     map[string]string
	Template string
	Category string
}

// parseTemplateVars извлекает все переменные шаблона из файла шаблона
func parseTemplateVars(templateContent string) map[string]string {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(templateContent, -1)

	vars := make(map[string]string)
	for _, match := range matches {
		if len(match) > 1 {
			vars[match[1]] = ""
		}
	}
	return vars
}

// loadTemplate загружает файл шаблона и возвращает его содержимое и переменные
func loadTemplate(templateName string) (string, map[string]string, error) {
	path := filepath.Join("templates", templateName+".tpl")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf(msgErrorReadingTemplate, templateName, err)
	}

	templateStr := string(content)
	vars := parseTemplateVars(templateStr)
	return templateStr, vars, nil
}

// processPage обрабатывает директорию одной страницы и возвращает Model
func processPage(pagePath string, blocks map[string]string) (*Model, error) {
	// Print blocks hashmap to the terminal
	/*
	fmt.Println("Blocks hashmap in processPage:")
	for k, v := range blocks {
		fmt.Printf("  %s: %s\n", k, v)
	}
	*/
	pageID := filepath.Base(pagePath)
	model := &Model{
		ID:   pageID,
		Data: make(map[string]string),
	}

	// Read category.val if exists
	categoryPath := filepath.Join(pagePath, "category.val")
	if _, err := os.Stat(categoryPath); err == nil {
		category, err := ioutil.ReadFile(categoryPath)
		if err != nil {
			return nil, fmt.Errorf(msgErrorReadingCategory, pageID, err)
		}
		model.Category = strings.TrimSpace(string(category))
	}

	// Чтение template.setting
	templateSettingPath := filepath.Join(pagePath, "template.setting")
	if _, err := os.Stat(templateSettingPath); err == nil {
		templateName, err := ioutil.ReadFile(templateSettingPath)
		if err != nil {
			return nil, fmt.Errorf(msgErrorReadingSetting, pageID, err)
		}
		model.Template = strings.TrimSpace(string(templateName))
	}

	// Чтение всех файлов attribute.val
	files, err := ioutil.ReadDir(pagePath)
	if err != nil {
		return nil, fmt.Errorf(msgErrorReadingPageDir, pageID, err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".val") {
			attrName := strings.TrimSuffix(file.Name(), ".val")
			content, err := ioutil.ReadFile(filepath.Join(pagePath, file.Name()))
			if err != nil {
				return nil, fmt.Errorf(msgErrorReadingAttr, attrName, pageID, err)
			}
			model.Data[attrName] = strings.TrimSpace(string(content))
		}
	}

	// Инициализируем атрибуты модели значениями из blocks
	for k, v := range blocks {
		model.Data[k] = v
	}

	// Print model key and values to the terminal
	/*
	fmt.Printf("Model: %s\n", model.ID)
	for k, v := range model.Data {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println("---")
	*/
	return model, nil
}

// renderTemplate применяет данные модели к шаблону
func renderTemplate(templateStr string, model *Model) (string, error) {
	re := regexp.MustCompile(`\{([^}]+)\}`)
	result := re.ReplaceAllStringFunc(templateStr, func(match string) string {
		key := match[1 : len(match)-1] // Удаляем фигурные скобки
		if value, exists := model.Data[key]; exists {
			return value
		}
		return match // Возвращаем оригинал, если не найдено
	})
	return result, nil
}

// Add new function to generate category files
type CategoryTask struct {
	Category string
	Items    []map[string]string
}

// processCategory handles the generation of JSON and HTML files for a single category
func processCategory(task CategoryTask, blocks map[string]string) error {
	// Generate JSON file
	jsonData, err := json.MarshalIndent(task.Items, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка при маршалинге JSON для категории %s: %v", task.Category, err)
	}
	jsonPath := filepath.Join("build", fmt.Sprintf("%s.json", task.Category))
	if err := ioutil.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("ошибка при записи JSON файла для категории %s: %v", task.Category, err)
	}

	// Generate HTML file
	htmlTplPath := filepath.Join("collections", "category.tpl")
	htmlBytes, err := ioutil.ReadFile(htmlTplPath)
	if err != nil {
		return fmt.Errorf("ошибка при чтении шаблона категории: %v", err)
	}
	htmlContent := string(htmlBytes)

	htmlContent = strings.ReplaceAll(htmlContent, "{{CATEGORY}}", task.Category)

	// Replace {key} with blocks[key] for all keys in blocks
	for k, v := range blocks {
		htmlContent = strings.ReplaceAll(htmlContent, "{"+k+"}", v)
	}

	htmlPath := filepath.Join("build", task.Category+".html")
	if err := ioutil.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("ошибка при записи HTML файла: %v", err)
	}

	return nil
}

// generateCategoryFiles now processes categories in parallel using a worker pool
func generateCategoryFiles(models []*Model, blocks map[string]string) error {
	// Group models by category
	categories := make(map[string][]map[string]string)
	for _, model := range models {
		if model.Category == "" {
			continue
		}

		if _, exists := categories[model.Category]; !exists {
			categories[model.Category] = make([]map[string]string, 0)
		}

		item := map[string]string{
			"title": model.Data["title"],
			"url":   fmt.Sprintf("/%s.html", model.ID),
		}
		categories[model.Category] = append(categories[model.Category], item)
	}

	// Create build directory if not exists
	if _, err := os.Stat("build"); os.IsNotExist(err) {
		os.Mkdir("build", 0755)
	}

	chTasks := make(chan CategoryTask, len(categories))
	chErrors := make(chan error, len(categories))
	var wg sync.WaitGroup

	numWorkers := 4 // Adjust as needed
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range chTasks {
				if err := processCategory(task, blocks); err != nil {
					chErrors <- err
				}
			}
		}()
	}

	for category, items := range categories {
		chTasks <- CategoryTask{Category: category, Items: items}
	}
	close(chTasks)

	wg.Wait()
	close(chErrors)

	for err := range chErrors {
		if err != nil {
			return err
		}
	}

	return nil
}

// getBlocksSubModel reads all .tpl files from ./blocks, stores their contents in a Blocks map,
// and substitutes {blockname} for other blocks using strings.Replace (no recursion, no self-reference).
func getBlocksSubModel() (map[string]string, error) {
	blocksDir := "blocks"
	blocks := make(map[string]string)
	blockOrder := make([]string, 0)

	// Read all .tpl files in blocksDir
	dirEntries, err := os.ReadDir(blocksDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".tpl") {
			/*
			fmt.Printf("Processing block: %s\n", name);
			*/
			blockName := strings.TrimSuffix(name, ".tpl")
			content, err := os.ReadFile(filepath.Join(blocksDir, name))
			if err != nil {
				return nil, err
			}
			blocks[blockName] = string(content)
			blockOrder = append(blockOrder, blockName)
		}
	}

	// Check for self-references
	for blockName, content := range blocks {
		if strings.Contains(content, "{"+blockName+"}") {
			return nil, fmt.Errorf("block '%s' contains a forbidden self-reference", blockName)
		}
	}

	/*
	// Substitute {blockname} for other blocks using strings.Replace
	for _, blockName := range blockOrder {
		for _, otherName := range blockOrder {
			if blockName == otherName {
				continue
			}
			blocks[blockName] = strings.ReplaceAll(blocks[blockName], "{"+otherName+"}", blocks[otherName])
		}
	}
	*/
	// Print all block names and their values to the terminal
	/*
	for k, v := range blocks {
		fmt.Printf("Block: %s\nValue:\n%s\n---\n", k, v)
	}
	*/
	return blocks, nil
}

// WriteTask is used to send output path and data to writing goroutines
type WriteTask struct {
	Path string
	Data []byte
}

func main() {
	start := time.Now()
	// Проверяем наличие необходимых директорий
	if _, err := os.Stat("templates"); os.IsNotExist(err) {
		fmt.Println(msgTemplatesDirNotFound)
		return
	}

	if _, err := os.Stat("content"); os.IsNotExist(err) {
		fmt.Println(msgContentDirNotFound)
		return
	}

	// Создаём директорию build, если она не существует
	if _, err := os.Stat("build"); os.IsNotExist(err) {
		os.Mkdir("build", 0755)
	}

	// Получаем блоки
	blocks, err := getBlocksSubModel()
	if err != nil {
		fmt.Printf("Ошибка при обработке блоков: %v\n", err)
		return
	}

	// Обрабатываем все страницы параллельно
	pageDirs, err := ioutil.ReadDir("content")
	if err != nil {
		fmt.Printf(msgErrorReadingContent, err)
		return
	}

	// Prepare list of page paths
	pages := make([]string, 0)
	for _, pageDir := range pageDirs {
		if pageDir.IsDir() {
			pages = append(pages, filepath.Join("content", pageDir.Name()))
		}
	}

	chModels := make(chan string, 10)
	chWriting := make(chan WriteTask, 10)
	chCollectedModels := make(chan *Model, len(pages))

	var wgReaders sync.WaitGroup
	var wgProcessors sync.WaitGroup
	var wgWriters sync.WaitGroup

	numReaders := 2
	numProcessors := 4
	numWriters := 2

	// Feed page paths to chModels in parallel
	pagesCh := make(chan string, len(pages))
	for _, p := range pages {
		pagesCh <- p
	}
	close(pagesCh)

	for i := 0; i < numReaders; i++ {
		wgReaders.Add(1)
		go func() {
			defer wgReaders.Done()
			for pagePath := range pagesCh {
				chModels <- pagePath
			}
		}()
	}

	// Processors: read from chModels, process, send to chWriting, collect models
	for i := 0; i < numProcessors; i++ {
		wgProcessors.Add(1)
		go func() {
			defer wgProcessors.Done()
			for pagePath := range chModels {
				model, err := processPage(pagePath, blocks)
				if err != nil {
					fmt.Printf(msgErrorProcessingPage, filepath.Base(pagePath), err)
					continue
				}
				if model.Template == "" {
					fmt.Printf(msgWarningNoTemplate, model.ID)
					continue
				}
				templateContent, templateVars, err := loadTemplate(model.Template)
				if err != nil {
					fmt.Printf(msgErrorLoadingTemplate, model.ID, err)
					continue
				}
				for k, v := range templateVars {
					if _, exists := model.Data[k]; !exists {
						model.Data[k] = v
					}
				}
				output, err := renderTemplate(templateContent, model)
				if err != nil {
					fmt.Printf(msgErrorRendering, model.ID, err)
					continue
				}
				outputPath := filepath.Join("build", model.ID+".html")
				chWriting <- WriteTask{Path: outputPath, Data: []byte(output)}
				chCollectedModels <- model
			}
		}()
	}

	// Writers: write files from chWriting
	for i := 0; i < numWriters; i++ {
		wgWriters.Add(1)
		go func() {
			defer wgWriters.Done()
			for task := range chWriting {
				err := ioutil.WriteFile(task.Path, task.Data, 0644)
				if err != nil {
					fmt.Printf(msgErrorWritingOutput, task.Path, err)
					continue
				}
				/*
				fmt.Printf(msgGenerated, task.Path)
				*/
			}
		}()
	}

	// Wait for all readers to finish, then close chModels
	wgReaders.Wait()
	close(chModels)
	// Wait for all processors to finish, then close chWriting
	wgProcessors.Wait()
	close(chWriting)

	// Collect all models from chCollectedModels
	var models []*Model
	for i := 0; i < len(pages); i++ {
		model := <-chCollectedModels
		if model != nil {
			models = append(models, model)
		}
	}
	close(chCollectedModels)

	// Wait for all writers to finish
	wgWriters.Wait()

	// Generate category files (single-threaded, as before)
	if err := generateCategoryFiles(models, blocks); err != nil {
		fmt.Printf("Ошибка при генерации файлов категорий: %v\n", err)
	}

	fmt.Println(msgSiteGenerationDone)
	elapsed := time.Since(start)
	fmt.Printf("Время выполнения: %d мс\n", elapsed.Milliseconds())
}
