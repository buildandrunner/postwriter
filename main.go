package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ollama/ollama/api"
)

// PostGenerator es una interfaz para los bots de generación de posts.
type PostGenerator interface {
	ExtractBusinessNameFromAbout(ctx context.Context, aboutContent string) (string, error)
	RefinePrompt(ctx context.Context, userPremise string) (string, error)
	GenTitle(ctx context.Context, businessInfo, premise string) (string, error)
	GenContent(ctx context.Context, businessInfo, title string) (string, error)
	GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error)
}

// ollamaPostGenerator genera contenido usando un modelo local vía Ollama.
type ollamaPostGenerator struct {
	client *api.Client
	model  string
}

// NewOllamaPostGenerator crea una nueva instancia que usa Ollama.
func NewOllamaPostGenerator(client *api.Client, model string) PostGenerator {
	if model == "" {
		model = "qwen3:8b"
	}
	return &ollamaPostGenerator{client: client, model: model}
}

// generate es un helper reutilizable para evitar duplicación en llamadas a Ollama.
func (o *ollamaPostGenerator) generate(ctx context.Context, system, prompt string) (string, error) {
	if prompt == "" {
		return "", fmt.Errorf("prompt cannot be empty")
	}

	req := &api.GenerateRequest{
		Model:  o.model,
		System: system,
		Prompt: prompt,
		Think:  &api.ThinkValue{Value: false},
	}

	var sb strings.Builder

	errFn := func(res api.GenerateResponse) error {
		fmt.Print(res.Response)
		_, err := sb.WriteString(res.Response)
		return err
	}

	if err := o.client.Generate(ctx, req, errFn); err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return strings.TrimSpace(sb.String()), nil
}

func (o *ollamaPostGenerator) ExtractBusinessNameFromAbout(ctx context.Context, aboutContent string) (string, error) {
	system := "You are in charge of determining what is the business name from this about us text. Only return the business name."
	return o.generate(ctx, system, aboutContent)
}

func (o *ollamaPostGenerator) RefinePrompt(ctx context.Context, userPremise string) (string, error) {
	system := "Your job is to refine the user premise idea for a facebook post. Refine the idea into something useful for an llm. Only return the refined prompt string sentence."
	return o.generate(ctx, system, userPremise)
}

func (o *ollamaPostGenerator) GenTitle(ctx context.Context, businessInfo, premise string) (string, error) {
	system := fmt.Sprintf(
		"You are a social media marketing expert. Create a short, catchy Facebook post title (less than 60 characters) based on the business info and the user's idea. Only return the title, nothing else.\n\nBusiness Info: %s",
		businessInfo,
	)
	return o.generate(ctx, system, premise)
}

func (o *ollamaPostGenerator) GenContent(ctx context.Context, businessInfo, title string) (string, error) {
	system := fmt.Sprintf(
		"You are a social media content writer. Write an engaging, friendly Facebook post based on the business info and the given title. Make it attention-grabbing, include emojis if appropriate, and end with a clear call to action (e.g., 'Visit us today!', 'DM for inquiries').\n\nBusiness Info: %s",
		businessInfo,
	)
	prompt := fmt.Sprintf("Write a Facebook post with this title: %s", title)
	return o.generate(ctx, system, prompt)
}

func (o *ollamaPostGenerator) GenImagePrompt(ctx context.Context, businessInfo, title, content string) (string, error) {
	system := fmt.Sprintf(
		"You are a visual artist and social media expert. Create a detailed, vivid image description that captures the essence of a Facebook post. Include style (e.g., realistic, cartoon, minimalist), lighting, mood, key elements, and composition. This will be used by an AI image generator.\n\nBusiness Info: %s",
		businessInfo,
	)
	prompt := fmt.Sprintf("Title: %s\n\nContent: %s", title, content)
	return o.generate(ctx, system, prompt)
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

	// Inicializar cliente Ollama
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatalln("Error al crear cliente de Ollama:", err)
	}

	// Usar Ollama como generador
	pg := NewOllamaPostGenerator(client, "qwen3:8b")

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
