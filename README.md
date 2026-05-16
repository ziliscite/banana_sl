# banana_sl

A CLI tool to generate images using the **Nano Banana** family of Gemini models (or Imagen), from a text prompt or an input image + prompt (image-to-image).

All generated images include a [SynthID watermark](https://ai.responsible.google/docs/safeguards/synthid).

---

## Nano Banana Model Family

| Alias | Model ID | Description |
|-------|----------|-------------|
| **Nano Banana 2** *(default)* | `gemini-3.1-flash-image-preview` | High-efficiency, optimised for speed and high-volume developer use cases |
| **Nano Banana Pro** | `gemini-3-pro-image-preview` | Professional asset production; advanced reasoning ("Thinking") for complex instructions and high-fidelity text rendering |
| **Nano Banana** | `gemini-2.5-flash-image` | Speed and efficiency, optimised for high-volume, low-latency tasks |

Imagen 3 / 4 models are also supported via `-mode imagen`.

---

## Requirements

- Go 1.23+
- A [Google AI Studio API key](https://aistudio.google.com/)

## Setup

```bash
git clone https://github.com/ziliscite/banana_sl
cd banana_sl
go mod tidy        # generates go.sum
go build -o imggen .
export GEMINI_API_KEY=your_api_key_here
```

---

## Usage

```
imggen -prompt "<text>" [options]
```

---

## Modes

| Mode | Flag | API call | Best for |
|------|------|----------|---------|
| `gemini` **(default)** | `-mode gemini` | `GenerateContent` | Nano Banana family; image-to-image editing; per-category safety tuning |
| `imagen` | `-mode imagen` | `GenerateImages` | Photorealistic images; full aspect-ratio, seed, watermark control |

---

## All Flags

### Input / Output

| Flag | Default | Description |
|------|---------|-------------|
| `-prompt` | **required** | Text prompt |
| `-image` | — | Path to input image (image-to-image / editing mode) |
| `-out-dir` | `.` | Output directory |
| `-out-name` | `generated_<timestamp>` | Base filename (no extension) |

### Model Selection

| Flag | Default | Description |
|------|---------|-------------|
| `-mode` | `gemini` | `gemini` or `imagen` |
| `-model` | `gemini-3.1-flash-image-preview` | Model ID (see table above) |

---

### Gemini / Nano Banana Flags

#### Generation Config

| Flag | Default | Description |
|------|---------|-------------|
| `-temperature` | model default | Sampling temperature 0.0–2.0 |
| `-top-p` | model default | Top-P nucleus sampling 0.0–1.0 |
| `-top-k` | model default | Top-K sampling |
| `-max-output-tokens` | model default | Max tokens in response |
| `-gc-seed` | random | Fixed seed for reproducibility |
| `-verbose` | `false` | Print model text responses alongside images |

#### Per-Category Safety Settings

Each flag accepts: `OFF` \| `BLOCK_NONE` \| `BLOCK_ONLY_HIGH` \| `BLOCK_MEDIUM_AND_ABOVE` \| `BLOCK_LOW_AND_ABOVE`

| Flag | Category |
|------|----------|
| `-safety-harassment` | `HARM_CATEGORY_HARASSMENT` |
| `-safety-hate-speech` | `HARM_CATEGORY_HATE_SPEECH` |
| `-safety-sexually-explicit` | `HARM_CATEGORY_SEXUALLY_EXPLICIT` |
| `-safety-dangerous` | `HARM_CATEGORY_DANGEROUS_CONTENT` |

---

### Imagen-specific Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-aspect-ratio` | `1:1` | `1:1` \| `3:4` \| `4:3` \| `9:16` \| `16:9` |
| `-num-images` | `1` | Number of images (1–4) |
| `-negative-prompt` | — | What to discourage in the image |
| `-guidance-scale` | model default | Prompt adherence (higher = more literal) |
| `-seed` | random | Fixed seed (incompatible with `-add-watermark`) |
| `-output-mime` | `image/png` | `image/png` \| `image/jpeg` \| `image/webp` |
| `-output-quality` | model default | JPEG compression quality 0–100 |
| `-add-watermark` | `false` | Add SynthID invisible watermark |
| `-enhance-prompt` | `true` | Let model rewrite/improve your prompt |
| `-person-generation` | `ALLOW_ADULT` | `DONT_ALLOW` \| `ALLOW_ADULT` \| `ALLOW_ALL` |
| `-safety-filter-level` | `BLOCK_MEDIUM_AND_ABOVE` | `BLOCK_LOW_AND_ABOVE` \| `BLOCK_MEDIUM_AND_ABOVE` \| `BLOCK_ONLY_HIGH` \| `BLOCK_NONE` |
| `-include-rai-reason` | `false` | Include RAI reason when image is filtered |
| `-include-safety-attrs` | `false` | Include safety attribute scores in output |

---

## Examples

```bash
# Nano Banana 2 (default) — text to image
imggen -prompt "A cyberpunk city at night"

# Nano Banana Pro — complex instruction with text rendering
imggen \
  -model gemini-3-pro-image-preview \
  -prompt "A vintage travel poster for Tokyo with bold text VISIT TOKYO at the top"

# Nano Banana — high-volume, low-latency
imggen \
  -model gemini-2.5-flash-image \
  -prompt "A simple icon of a rocket ship"

# Nano Banana 2 — image-to-image editing
imggen \
  -prompt "Transform this into a watercolor painting" \
  -image ./my_photo.jpg \
  -out-name watercolor_edit

# Nano Banana 2 — custom safety settings
imggen \
  -prompt "A dramatic medieval battle scene" \
  -safety-harassment BLOCK_NONE \
  -safety-dangerous BLOCK_ONLY_HIGH \
  -verbose

# Nano Banana 2 — seed + temperature
imggen \
  -prompt "An astronaut floating above a neon planet" \
  -gc-seed 42 \
  -temperature 1.2 \
  -out-dir ./output \
  -out-name astronaut

# Imagen 3 — photorealistic, landscape
imggen \
  -mode imagen \
  -model imagen-3.0-generate-002 \
  -prompt "Aerial view of a volcanic island at sunset" \
  -aspect-ratio 16:9 \
  -num-images 2 \
  -out-dir ./output

# Imagen 4 preview — JPEG output, seed
imggen \
  -mode imagen \
  -model imagen-4.0-generate-preview-05-20 \
  -prompt "Hyperdetailed macro photo of a dewdrop on a spider web" \
  -seed 1337 \
  -output-mime image/jpeg \
  -output-quality 95
```

---

## Project Structure

```
banana_sl/
├── main.go     # Single-file CLI application
├── go.mod
├── go.sum      # generated by: go mod tidy
└── README.md
```

## References

- [Gemini Image Generation docs](https://ai.google.dev/gemini-api/docs/image-generation)
- [Gemini 3.1 Flash Image Preview](https://ai.google.dev/gemini-api/docs/models/gemini-3.1-flash-image-preview)
- [Gemini 3 Pro Image Preview](https://ai.google.dev/gemini-api/docs/models/gemini-3-pro-image-preview)
- [Gemini 2.5 Flash Image](https://ai.google.dev/gemini-api/docs/models/gemini-2.5-flash-image)
- [Safety Settings docs](https://ai.google.dev/gemini-api/docs/safety-settings)
- [SynthID Watermark](https://ai.responsible.google/docs/safeguards/synthid)
- [Google AI Go SDK](https://github.com/googleapis/go-genai)
