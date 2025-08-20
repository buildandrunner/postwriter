# 📝 Generador de Contenido para Redes Sociales con Ollama

## 🌟 Descripción

Este proyecto es una herramienta en Go que utiliza modelos de inteligencia artificial locales (mediante [Ollama](https://ollama.com)) para generar contenido de redes sociales de forma automatizada. A partir de una breve idea del usuario y la información de un negocio (en `about.md`), genera:

- Un **título atractivo** para una publicación de Facebook.
- Un **contenido completo y llamativo** con emojis y llamado a la acción.
- Una **descripción detallada para generar una imagen** con IA (útil para DALL·E, Midjourney, etc.).
- Todo se guarda de forma organizada en carpetas numeradas.

Ideal para emprendedores, community managers o desarrolladores que deseen automatizar la creación de contenido usando modelos locales sin depender de APIs externas.


## 🛠️ Tecnologías utilizadas

- [Go](https://go.dev) – Lenguaje principal.
- [Ollama](https://ollama.com) – Para ejecutar modelos de IA localmente.
- Modelo recomendado: `qwen3:0.6b` (por ligero y rápido), pero puedes usar cualquier modelo compatible.


## 📦 Requisitos previos

1. **Instalar Ollama**:  
   Visita [https://ollama.com](https://ollama.com) y sigue las instrucciones para tu sistema operativo.

2. **Descargar un modelo de IA**, por ejemplo:  
   ```bash
   ollama pull qwen3:0.6b
   ```
   Puedes usar otros modelos como `llama3`, `phi3`, etc., si lo prefieres.

3. **Tener Go instalado** (versión 1.20 o superior recomendada).

4. **Crear un archivo `about.md`** en el directorio raíz con información sobre tu negocio, por ejemplo:
   ```markdown
   Somos Cafetería Buenos Días, un pequeño negocio familiar ubicado en el centro de la ciudad. Ofrecemos café de especialidad, pasteles caseros y un ambiente acogedor para trabajar o reunirse con amigos.
   ```


## ▶️ Cómo usar

1. **Clona o crea el proyecto** y coloca tu descripción en `about.md`.

2. **Ejecuta el programa** pasando como argumento la idea para tu publicación:
   ```bash
   go run . "Hoy queremos promocionar nuestro café con leche especial"
   ```

3. El programa hará lo siguiente:
   - Extrae el nombre del negocio usando IA.
   - Refina tu idea.
   - Genera título, contenido y prompt de imagen.
   - Guarda todo en `posts/<nombre_negocio>/XXXX/`.

4. **Resultado guardado en**:
   ```
   posts/
   └── Cafetería Buenos Días/
       └── 0001/
           ├── post.md
           └── image_prompt.txt
   ```


## 📁 Estructura de salida

- `post.md`: Contiene el título y el contenido del post.
- `image_prompt.txt`: Prompt detallado para generar una imagen con IA.


## ⚙️ Personalización

- Cambia el modelo predeterminado pasando uno al `NewOllamaPostGenerator`, por ejemplo:
  ```go
  pg := NewOllamaPostGenerator(client, "llama3")
  ```
- Modifica los mensajes del sistema (`system` prompts) para adaptar el estilo del contenido.


## ⏳ Retraso integrado

El programa incluye un `delay` opcional para evitar saturar el modelo local si se hacen múltiples ejecuciones. Puedes ajustarlo o comentarlo si no es necesario.


## ✅ Ejemplo de uso

```bash
go run . "Ofrecemos descuento del 20% los miércoles"
```

Salida posible:
- **Título**: ¡Miércoles de Café a Mitad de Precio!
- **Contenido**:  
  ¿Sabías que cada miércoles tenemos 20% de descuento en todos los cafés? 🎉  
  Tráete a tus amigos y disfruta de nuestro ambiente acogedor. ☕  
  ¡Ven hoy y haz del miércoles tu día favorito! 🧡  
  → Visítanos desde las 8:00 AM.

- **Prompt de imagen**:  
  Una mesa de madera con dos tazas de café humeantes, un cartel que dice "20% OFF", ambiente cálido y luminoso, estilo fotorealista, luz natural, detalles en vapor...


## 📂 Próximos pasos (ideas)

- Agregar soporte para múltiples redes sociales (Instagram, LinkedIn).
- Generar imágenes automáticamente usando Stable Diffusion u otra API.
- Soporte para imágenes en base64 o exportación directa.
- Interfaz web sencilla.


## 📄 Licencia

Este proyecto está bajo la Licencia MIT.  

🚀 **¡Automatiza tu contenido y enfócate en lo que más importa!**
