package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// PostGenerator es una interfaz para los bots de generación de posts.
type PostGenerator interface {
	GenTitle(ctx context.Context, businessInfo, premise string) (string, error)
	GenContent(ctx context.Context, businessInfo, title string) (string, error)
	GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error)
	RefinePrompt(ctx context.Context, userPremise string) (string, error)
	ExtractBusinessNameFromAI(ctx context.Context, aboutContent string) (string, error)
}

// openaiPostGenerator es la implementación de PostGenerator que usa el cliente de OpenAI.
type openaiPostGenerator struct {
	client *openai.Client
}

// NewOpenAIPostGenerator crea una nueva instancia de openaiPostGenerator.
func NewOpenAIPostGenerator(apiKey string) PostGenerator {
	client := openai.NewClient(apiKey)
	return &openaiPostGenerator{
		client: client,
	}
}

// ExtractBusinessNameFromAI utiliza IA para extraer el nombre del negocio del contenido de about.md.
func (o *openaiPostGenerator) ExtractBusinessNameFromAI(ctx context.Context, aboutContent string) (string, error) {
	systemMsg := "Eres un asistente de IA experto en análisis de texto. El usuario te proporcionará la información de un negocio. Tu única tarea es extraer el nombre del negocio de este texto. Devuelve solo el nombre del negocio, sin comillas ni texto adicional. Si no estás seguro, devuelve 'negocio desconocido'."

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: aboutContent,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error al extraer el nombre del negocio: %w", err)
	}
	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no se recibió una respuesta del modelo")
}

// RefinePrompt toma una idea simple del usuario y la convierte en un prompt detallado para la IA.
func (o *openaiPostGenerator) RefinePrompt(ctx context.Context, userPremise string) (string, error) {
	systemMsg := "Eres un asistente de IA experto en redacción de prompts. El usuario te dará una idea simple para un post. Tu tarea es expandir esa idea en un prompt detallado y claro que sirva como base para generar un título y un contenido de post de alta calidad."
	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userPremise,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error al refinar el prompt: %w", err)
	}
	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no se recibió una respuesta del modelo")
}

// GenTitle utiliza un modelo de IA para generar un título.
func (o *openaiPostGenerator) GenTitle(ctx context.Context, businessInfo string, premise string) (string, error) {
	systemMsg := "Eres un experto en marketing social. Crea un título para un post de Facebook basado en la información del negocio y el tema proporcionado. Devuelve solo un título corto de menos de 60 caracteres."
	systemMsg += " Información del negocio: " + businessInfo

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: premise,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error al generar el título: %w", err)
	}
	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no se recibió una respuesta del modelo")
}

// GenContent utiliza un modelo de IA para generar el contenido del post.
func (o *openaiPostGenerator) GenContent(ctx context.Context, businessInfo string, title string) (string, error) {
	systemMsg := "Eres un experto en marketing social. Basándote en la información del negocio y el título proporcionado, crea el contenido completo de un post de Facebook que sea atractivo y capte la atención del usuario. Incluye un llamado a la acción al final."
	systemMsg += " Información del negocio: " + businessInfo

	prompt := "Crea el contenido del post. El título es: " + title

	resp, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error al generar el contenido: %w", err)
	}
	if len(resp.Choices) > 0 {
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("no se recibió una respuesta del modelo")
}

// GenImagePrompt utiliza un modelo de IA para generar una descripción de la imagen para el post.
func (o *openaiPostGenerator) GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error) {
	systemMsg := "Eres un experto en marketing social y un artista visual. Basándote en el título y el contenido del post, crea una descripción detallada y vívida de la imagen que mejor represente el post. La descripción debe ser ideal para un generador de imágenes de IA. Incluye detalles sobre el estilo, la iluminación y la composición."
	systemMsg += " Información del negocio: " + businessInfo

	prompt := "Crea una descripción de la imagen. Título: " + title + "\nContenido: " + content

	respText, err := o.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMsg,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error al generar la descripción de la imagen: %w", err)
	}

	if len(respText.Choices) == 0 {
		return "", fmt.Errorf("no se recibió una respuesta del modelo para el prompt de imagen")
	}

	return strings.TrimSpace(respText.Choices[0].Message.Content), nil
}

// loadAbout lee el archivo "about.md" del directorio actual.
func loadAbout() (string, error) {
	contentBytes, err := os.ReadFile("about.md")
	if err != nil {
		return "", fmt.Errorf("error al leer about.md: %w", err)
	}
	return string(contentBytes), nil
}

// savePost guarda el título, el contenido y el prompt de la imagen en archivos separados en el directorio del negocio.
func savePost(businessName, title, content, imagePrompt string) error {
	baseDir := filepath.Join("posts", businessName)
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return fmt.Errorf("error al crear el directorio base '%s': %w", baseDir, err)
	}

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

	postContent := fmt.Sprintf("# %s\n\n%s", title, content)
	postPath := filepath.Join(newDirPath, "post.md")
	if err := os.WriteFile(postPath, []byte(postContent), 0644); err != nil {
		return fmt.Errorf("error al guardar post.md: %w", err)
	}

	imagePromptPath := filepath.Join(newDirPath, "image_prompt.txt")
	if err := os.WriteFile(imagePromptPath, []byte(imagePrompt), 0644); err != nil {
		return fmt.Errorf("error al guardar image_prompt.txt: %w", err)
	}

	log.Printf("Post guardado para el negocio '%s' en: %s", businessName, newDirPath)
	return nil
}

// delay detiene la ejecución del programa durante un número de segundos especificado.
func delay(seconds int) {
	log.Printf("Esperando %d segundos para evitar la limitación de tasa...", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Uso: go run . \"Tu idea para el post aquí\"")
	}

	userPremise := os.Args[1]

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatalln("La variable de entorno OPENAI_API_KEY no está configurada")
	}
	pg := NewOpenAIPostGenerator(apiKey)

	// Carga la información del negocio
	businessInfo, err := loadAbout()
	if err != nil {
		log.Fatalln("Error al cargar la información del negocio:", err)
	}

	// Extrae el nombre del negocio usando la IA
	log.Println("Extrayendo el nombre del negocio con IA...")
	businessName, err := pg.ExtractBusinessNameFromAI(context.Background(), businessInfo)
	if err != nil {
		log.Fatalln("Error al extraer el nombre del negocio:", err)
	}
	log.Println("Nombre del negocio extraído:", businessName)
	log.Println("---")
	delay(5)

	// Refina la idea inicial del usuario.
	log.Println("Refinando la idea del post...")
	refinedPrompt, err := pg.RefinePrompt(context.Background(), userPremise)
	if err != nil {
		log.Fatalln("Error al refinar el prompt:", err)
	}
	log.Println("Prompt refinado:", refinedPrompt)
	log.Println("---")
	delay(5)

	// Genera el título.
	log.Printf("Generando post para el negocio: %s...", businessName)
	title, err := pg.GenTitle(context.Background(), businessInfo, refinedPrompt)
	if err != nil {
		log.Fatalln("Error al generar el título:", err)
	}
	delay(5)

	// Genera el contenido del post.
	content, err := pg.GenContent(context.Background(), businessInfo, title)
	if err != nil {
		log.Fatalln("Error al generar el contenido:", err)
	}
	delay(5)

	// Genera el prompt de la imagen.
	imagePrompt, err := pg.GenImagePrompt(context.Background(), businessInfo, title, content)
	if err != nil {
		log.Fatalln("Error al generar el prompt de la imagen:", err)
	}
	delay(5)

	// Guarda el post completo
	if err := savePost(businessName, title, content, imagePrompt); err != nil {
		log.Fatalln("Error al guardar el post:", err)
	}
	log.Println("---")
}
