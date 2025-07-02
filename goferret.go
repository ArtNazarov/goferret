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
)

// Model представляет страницу с её атрибутами и шаблоном
type Model struct {
	ID       string
	Data     map[string]string
	Template string
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
func processPage(pagePath string) (*Model, error) {
	pageID := filepath.Base(pagePath)
	model := &Model{
		ID:   pageID,
		Data: make(map[string]string),
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
			model, err := processPage(pagePath)
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

	fmt.Println(msgSiteGenerationDone)
}
