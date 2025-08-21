package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// PostGenerator es una interfaz para los bots de generación de posts.
type PostGenerator interface {
	ExtractBusinessNameFromAbout(ctx context.Context, aboutContent string) (string, error)
	RefinePrompt(ctx context.Context, userPremise string) (string, error)
	GenTitle(ctx context.Context, businessInfo, premise string) (string, error)
	GenContent(ctx context.Context, businessInfo, title string) (string, error)
	GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error)
}

type openaiPostGenerator struct {
	client *openai.Client
}

func (o *openaiPostGenerator) ExtractBusinessNameFromAbout(ctx context.Context, aboutContent string) (string, error) {
	systemPrompt := "You are in charge of determining what is the business name from this about us text. Only return the business name."

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: aboutContent,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error extracting business name: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (o *openaiPostGenerator) RefinePrompt(ctx context.Context, userPremise string) (string, error) {
	systemPrompt := "Your job is to refine the user premise idea for a facebook post. Refine the idea into something useful for an llm. Only return the refined prompt string sentence."

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPremise,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error refining prompt: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (o *openaiPostGenerator) GenTitle(ctx context.Context, businessInfo, premise string) (string, error) {
	systemPrompt := fmt.Sprintf(
		"You are a social media marketing expert. Create a short, catchy Facebook post title (less than 60 characters) based on the business info and the user's idea. Only return the title, nothing else.\n\nBusiness Info: %s",
		businessInfo,
	)

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: premise,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error generating title: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (o *openaiPostGenerator) GenContent(ctx context.Context, businessInfo, title string) (string, error) {
	systemPrompt := fmt.Sprintf(
		"You are a social media content writer. Write an engaging, friendly Facebook post based on the business info and the given title. Make it attention-grabbing, include emojis if appropriate, and end with a clear call to action (e.g., 'Visit us today!', 'DM for inquiries').\n\nBusiness Info: %s",
		businessInfo,
	)
	userPrompt := fmt.Sprintf("Write a Facebook post with this title: %s", title)

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (o *openaiPostGenerator) GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error) {
	systemPrompt := fmt.Sprintf(
		"You are a visual artist and social media expert. Create a detailed, vivid image description that captures the essence of a Facebook post. Include style (e.g., realistic, cartoon, minimalist), lighting, mood, key elements, and composition. This will be used by an AI image generator.\n\nBusiness Info: %s",
		businessInfo,
	)
	userPrompt := fmt.Sprintf("Title: %s\n\nContent: %s", title, content)

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error generating image prompt: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func NewOpenaiPostGenerator(client *openai.Client) PostGenerator {
	return &openaiPostGenerator{client: client}
}

// loadAbout lee el archivo "about.md" del directorio actual.
func loadAbout() (string, error) {
	contentBytes, err := os.ReadFile("about.md")
	if err != nil {
		return "", fmt.Errorf("error al leer about.md: %w", err)
	}
	return string(contentBytes), nil
}

// savePost guarda el título, el contenido y el prompt de la imagen en archivos separados.
func savePost(businessName, title, content, imagePrompt string) error {
	baseDir := filepath.Join("posts", businessName)
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear el directorio base '%s': %w", baseDir, err)
	}

	// Leer directorios existentes para encontrar el siguiente número
	dirEntries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("error al leer el directorio '%s': %w", baseDir, err)
	}

	nextNumber := 1
	for _, entry := range dirEntries {
		if entry.IsDir() {
			if num, err := strconv.Atoi(entry.Name()); err == nil && num >= nextNumber {
				nextNumber = num + 1
			}
		}
	}

	newDirName := fmt.Sprintf("%04d", nextNumber)
	newDirPath := filepath.Join(baseDir, newDirName)

	if err := os.MkdirAll(newDirPath, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear el directorio de post: %w", err)
	}

	// Guardar post.md
	postContent := fmt.Sprintf("# %s\n\n%s", title, content)
	postPath := filepath.Join(newDirPath, "post.md")
	if err := os.WriteFile(postPath, []byte(postContent), 0644); err != nil {
		return fmt.Errorf("error al guardar post.md: %w", err)
	}

	// Guardar image_prompt.txt
	imagePromptPath := filepath.Join(newDirPath, "image_prompt.txt")
	if err := os.WriteFile(imagePromptPath, []byte(imagePrompt), 0644); err != nil {
		return fmt.Errorf("error al guardar image_prompt.txt: %w", err)
	}

	log.Printf("Post guardado para el negocio '%s' en: %s", businessName, newDirPath)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Uso: go run . \"Tu idea para el post aquí\"")
	}

	userPremise := os.Args[1]

	// Inicializar cliente OpenAI
	openaiClient := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// Usar OpenAI como generador (cambiar a NewOllamaPostGenerator para usar Ollama)
	pg := NewOpenaiPostGenerator(openaiClient)

	// Cargar información del negocio
	businessInfo, err := loadAbout()
	if err != nil {
		log.Fatalln("Error al cargar la información del negocio:", err)
	}

	// Extraer nombre del negocio
	log.Println("Extrayendo el nombre del negocio con IA...")
	businessName, err := pg.ExtractBusinessNameFromAbout(context.Background(), businessInfo)
	if err != nil {
		log.Fatalln("Error al extraer el nombre del negocio:", err)
	}
	businessName = strings.TrimSpace(businessName)
	if businessName == "" {
		businessName = "negocio_desconocido"
	}
	log.Println("Nombre del negocio extraído:", businessName)

	// Refinar idea del usuario
	log.Println("Refinando la idea del usuario...")
	refinedPrompt, err := pg.RefinePrompt(context.Background(), userPremise)
	if err != nil {
		log.Fatalln("Error al refinar el prompt:", err)
	}
	log.Println("Prompt refinado:", refinedPrompt)

	// Generar título
	log.Println("Generando título...")
	title, err := pg.GenTitle(context.Background(), businessInfo, refinedPrompt)
	if err != nil {
		log.Fatalln("Error al generar el título:", err)
	}
	log.Println("Título:", title)

	// Generar contenido
	log.Println("Generando contenido...")
	content, err := pg.GenContent(context.Background(), businessInfo, title)
	if err != nil {
		log.Fatalln("Error al generar el contenido:", err)
	}
	log.Println("Contenido:", content)

	// Generar prompt de imagen
	log.Println("Generando descripción de imagen...")
	imagePrompt, err := pg.GenImagePrompt(context.Background(), businessInfo, title, content)
	if err != nil {
		log.Fatalln("Error al generar el prompt de imagen:", err)
	}
	log.Println("Prompt de imagen:", imagePrompt)

	// Guardar todo
	if err := savePost(businessName, title, content, imagePrompt); err != nil {
		log.Fatalln("Error al guardar el post:", err)
	}

	log.Println("✅ Post generado y guardado con éxito.")
}
