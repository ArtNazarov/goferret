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
	fmt.Println("Blocks hashmap in processPage:")
	for k, v := range blocks {
		fmt.Printf("  %s: %s\n", k, v)
	}

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
	fmt.Printf("Model: %s\n", model.ID)
	for k, v := range model.Data {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Println("---")

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
func generateCategoryFiles(models []*Model) error {
	// Group models by category
	categories := make(map[string][]map[string]string)
	for _, model := range models {
		if model.Category == "" {
			continue
		}

		if _, exists := categories[model.Category]; !exists {
			categories[model.Category] = make([]map[string]string, 0)
		}

		// Create simplified model for JSON
		item := map[string]string{
			"title": model.Data["title"], // Assuming each model has a title
			"url":   fmt.Sprintf("/%s.html", model.ID),
		}
		categories[model.Category] = append(categories[model.Category], item)
	}

	// Create build directory if not exists
	if _, err := os.Stat("build"); os.IsNotExist(err) {
		os.Mkdir("build", 0755)
	}

	// Generate JSON files per category
	for category, items := range categories {
		jsonData, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return fmt.Errorf("ошибка при маршалинге JSON для категории %s: %v", category, err)
		}
		jsonPath := filepath.Join("build", fmt.Sprintf("%s.json", category))
		if err := ioutil.WriteFile(jsonPath, jsonData, 0644); err != nil {
			return fmt.Errorf("ошибка при записи JSON файла для категории %s: %v", category, err)
		}
		fmt.Printf("Сгенерировано: %s\n", jsonPath)

		// Generate HTML file
	htmlContent := `<!DOCTYPE html>
	<html lang="ru">
	<head>
		<meta charset="UTF-8">
		<title>Категории</title>
		<style>
			table { border-collapse: collapse; width: 100%; }
			th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
			.category { margin-bottom: 20px; }
			h2 { margin-top: 30px; }
			.pagination { margin: 20px 0; text-align: center; }
			.pagination button { margin: 0 2px; padding: 5px 10px; }
		</style>
		<script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
	</head>
	<body>
		<h1>Категории</h1>
		<div id="categoriesContainer"></div>
		<div class="pagination" id="pagination"></div>

		<script>
			const category = "{{CATEGORY}}";
			const url = window.location.protocol + '//' + window.location.host + '/' + category + '.json';
			const ROWS_PER_PAGE = 10;
			let currentPage = 1;
			let data = [];

			function renderTable(page) {
				const container = $('#categoriesContainer');
				container.empty();
				const start = (page - 1) * ROWS_PER_PAGE;
				const end = start + ROWS_PER_PAGE;
				const pageData = data.slice(start, end);

				const categoryDiv = $('<div>').addClass('category');
				const header = $('<h2>').text(category);
				categoryDiv.append(header);

				const table = $('<table>');
				const thead = $('<thead>').html('<tr><th>Заголовок</th><th>Ссылка</th></tr>');
				table.append(thead);
				const tbody = $('<tbody>');
				pageData.forEach(function(item) {
					const row = $('<tr>');
					row.html('<td>' + item.title + '</td><td><a href="' + item.url + '">' + item.url + '</a></td>');
					tbody.append(row);
				});
				table.append(tbody);
				categoryDiv.append(table);
				container.append(categoryDiv);
			}

			function renderPagination() {
				const totalPages = Math.ceil(data.length / ROWS_PER_PAGE);
				const pagination = $('#pagination');
				pagination.empty();
				if (totalPages <= 1) return;
				for (let i = 1; i <= totalPages; i++) {
					const btn = $('<button>').text(i);
					if (i === currentPage) btn.attr('disabled', true);
					btn.on('click', function() {
						currentPage = i;
						renderTable(currentPage);
						renderPagination();
					});
					pagination.append(btn);
				}
			}

			$(document).ready(function() {
				$.getJSON(url, function(json) {
					data = json;
					currentPage = 1;
					renderTable(currentPage);
					renderPagination();
				}).fail(function() {
					$('#categoriesContainer').html('<p>Ошибка загрузки данных категорий</p>');
				});
			});
		</script>
	</body>
	</html>`
	
		htmlContent = strings.ReplaceAll(htmlContent, "{{CATEGORY}}", category)

		htmlPath := filepath.Join("build", category+".html")
		if err := ioutil.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
			return fmt.Errorf("ошибка при записи HTML файла: %v", err)
		}
		fmt.Printf("Сгенерировано: %s\n", htmlPath)
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
			fmt.Printf("Processing block: %s\n", name);
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
	for k, v := range blocks {
		fmt.Printf("Block: %s\nValue:\n%s\n---\n", k, v)
	}
	return blocks, nil
}

func main() {
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

	// Обрабатываем все страницы
	var models []*Model

	pageDirs, err := ioutil.ReadDir("content")
	if err != nil {
		fmt.Printf(msgErrorReadingContent, err)
		return
	}

	for _, pageDir := range pageDirs {
		if pageDir.IsDir() {
			pagePath := filepath.Join("content", pageDir.Name())
			model, err := processPage(pagePath, blocks)
			if err != nil {
				fmt.Printf(msgErrorProcessingPage, pageDir.Name(), err)
				continue
			}
			models = append(models, model)
		}
	}

	// Обрабатываем каждую модель и генерируем вывод
	for _, model := range models {
		if model.Template == "" {
			fmt.Printf(msgWarningNoTemplate, model.ID)
			continue
		}

		templateContent, templateVars, err := loadTemplate(model.Template)
		if err != nil {
			fmt.Printf(msgErrorLoadingTemplate, model.ID, err)
			continue
		}

		// Объединяем переменные шаблона с данными модели (переменные шаблона имеют пустые значения по умолчанию)
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
		err = ioutil.WriteFile(outputPath, []byte(output), 0644)
		if err != nil {
			fmt.Printf(msgErrorWritingOutput, model.ID, err)
			continue
		}

		fmt.Printf(msgGenerated, outputPath)
	}

	// Generate category files
	if err := generateCategoryFiles(models); err != nil {
		fmt.Printf("Ошибка при генерации файлов категорий: %v\n", err)
	}

	fmt.Println(msgSiteGenerationDone)
}
