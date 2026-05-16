# gemini-imggen

A CLI tool to generate images using **Google Gemini / Imagen** (Nano Banana 2) from a text prompt,
or optionally from an input image + prompt (image-to-image).

---

## Requirements

- Go 1.23+
- A [Google AI Studio API key](https://aistudio.google.com/)

## Setup

```bash
export GEMINI_API_KEY=your_api_key_here
go build -o gemini-imggen .
```

---

## Usage

```
gemini-imggen -prompt "<text>" [options]
```

---

## Modes

| Mode | Flag | API call | Best for |
|------|------|----------|---------|
| `imagen` (default) | `-mode imagen` | `GenerateImages` | High-quality photorealistic images; full aspect-ratio, seed, watermark control |
| `gemini` | `-mode gemini` | `GenerateContent` | Conversational image generation; image-to-image editing |

---

## All Flags

### Input / Output

| Flag | Default | Description |
|------|---------|-------------|
| `-prompt` | **required** | Text prompt |
| `-image` | — | Path to input image (gemini mode: image-to-image; imagen mode: not used) |
| `-out-dir` | `.` | Output directory |
| `-out-name` | `generated_<timestamp>` | Base filename (no extension) |

### Model Selection

| Flag | Default | Description |
|------|---------|-------------|
| `-mode` | `imagen` | `imagen` or `gemini` |
| `-model` | `imagen-3.0-generate-002` | Model ID |

**Imagen models:**
- `imagen-3.0-generate-002` — Imagen 3 (Nano Banana 2), high quality
- `imagen-3.0-fast-generate-001` — Imagen 3 Fast, lower latency
- `imagen-4.0-generate-preview-05-20` — Imagen 4 preview

**Gemini models:**
- `gemini-2.0-flash-preview-image-generation`
- `gemini-2.5-flash-preview-05-20`

---

### Imagen-specific Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-aspect-ratio` | `1:1` | `1:1` \| `3:4` \| `4:3` \| `9:16` \| `16:9` |
| `-num-images` | `1` | Number of images to generate (1–4) |
| `-negative-prompt` | — | What to discourage in the image |
| `-guidance-scale` | model default | Prompt adherence (higher = more literal) |
| `-seed` | random | Fixed seed for reproducibility (incompatible with `-add-watermark`) |
| `-output-mime` | `image/png` | `image/png` \| `image/jpeg` \| `image/webp` |
| `-output-quality` | model default | JPEG compression quality 0–100 |
| `-add-watermark` | `false` | Add SynthID invisible watermark |
| `-enhance-prompt` | `true` | Let model rewrite/improve your prompt |
| `-person-generation` | `ALLOW_ADULT` | `DONT_ALLOW` \| `ALLOW_ADULT` \| `ALLOW_ALL` |
| `-safety-filter-level` | `BLOCK_MEDIUM_AND_ABOVE` | `BLOCK_LOW_AND_ABOVE` \| `BLOCK_MEDIUM_AND_ABOVE` \| `BLOCK_ONLY_HIGH` \| `BLOCK_NONE` |
| `-include-rai-reason` | `false` | Include RAI reason when image is filtered |
| `-include-safety-attrs` | `false` | Include safety attribute scores in output |

---

### Gemini GenerateContent Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-temperature` | model default | Sampling temperature 0.0–2.0 |
| `-top-p` | model default | Top-P nucleus sampling 0.0–1.0 |
| `-top-k` | model default | Top-K sampling |
| `-max-output-tokens` | model default | Max tokens in response |
| `-gc-seed` | random | Fixed seed for GenerateContent |
| `-verbose` | `false` | Print model text responses alongside images |

### Gemini Per-Category Safety Settings

Each flag accepts: `OFF` \| `BLOCK_NONE` \| `BLOCK_ONLY_HIGH` \| `BLOCK_MEDIUM_AND_ABOVE` \| `BLOCK_LOW_AND_ABOVE`

| Flag | Category |
|------|----------|
| `-safety-harassment` | `HARM_CATEGORY_HARASSMENT` |
| `-safety-hate-speech` | `HARM_CATEGORY_HATE_SPEECH` |
| `-safety-sexually-explicit` | `HARM_CATEGORY_SEXUALLY_EXPLICIT` |
| `-safety-dangerous` | `HARM_CATEGORY_DANGEROUS_CONTENT` |

---

## Examples

```bash
# Imagen 3 — basic text-to-image
gemini-imggen -prompt "A photorealistic wolf standing in a misty forest at dawn"

# Imagen 3 — landscape with seed, multiple images, no watermark
gemini-imggen \
  -prompt "Aerial view of a volcanic island at sunset" \
  -aspect-ratio 16:9 \
  -num-images 4 \
  -seed 1337 \
  -add-watermark=false \
  -out-dir ./output \
  -out-name volcano_island

# Imagen 3 — JPEG output, tight safety filter
gemini-imggen \
  -prompt "Studio portrait of a professional" \
  -output-mime image/jpeg \
  -output-quality 95 \
  -person-generation ALLOW_ALL \
  -safety-filter-level BLOCK_ONLY_HIGH

# Imagen 4 preview
gemini-imggen \
  -model imagen-4.0-generate-preview-05-20 \
  -prompt "Hyperdetailed macro photo of a dewdrop on a spider web" \
  -aspect-ratio 3:4

# Gemini image generation mode
gemini-imggen \
  -mode gemini \
  -model gemini-2.0-flash-preview-image-generation \
  -prompt "Draw a cartoon cat riding a skateboard" \
  -temperature 1.0

# Gemini image-to-image
gemini-imggen \
  -mode gemini \
  -model gemini-2.0-flash-preview-image-generation \
  -prompt "Transform this into a cyberpunk illustration with neon colors" \
  -image ./my_photo.jpg \
  -out-name cyberpunk_edit

# Gemini with all safety settings loosened
gemini-imggen \
  -mode gemini \
  -prompt "A dramatic battle scene" \
  -safety-harassment BLOCK_NONE \
  -safety-hate-speech BLOCK_NONE \
  -safety-sexually-explicit BLOCK_ONLY_HIGH \
  -safety-dangerous BLOCK_ONLY_HIGH
```

---

## Project Structure

```
gemini-imggen/
├── main.go       # Single-file CLI application
├── go.mod
├── go.sum
└── README.md
```

## References

- [Gemini Image Generation docs](https://ai.google.dev/gemini-api/docs/image-generation)
- [Safety Settings docs](https://ai.google.dev/gemini-api/docs/safety-settings)
- [Google AI Go SDK](https://github.com/googleapis/go-genai)
