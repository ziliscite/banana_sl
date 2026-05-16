package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

// ─────────────────────────────────────────────
// Nano Banana model aliases
// ─────────────────────────────────────────────
//
// Nano Banana 2   → gemini-3.1-flash-image-preview  (default)
// Nano Banana Pro → gemini-3-pro-image-preview
// Nano Banana     → gemini-2.5-flash-image
//
// Imagen models are also supported via -mode imagen.

// ─────────────────────────────────────────────
// CLI flags
// ─────────────────────────────────────────────

type config struct {
	// Input
	prompt    string
	imagePath string

	// Output
	outputDir  string
	outputName string

	// Mode & model
	// "gemini" mode  → client.Models.GenerateContent (Nano Banana family)
	// "imagen" mode  → client.Models.GenerateImages  (Imagen 3/4)
	mode  string
	model string

	// ── imagen-specific ──────────────────────────────────
	aspectRatio        string
	numberOfImages     int
	negativePrompt     string
	guidanceScale      float64
	seed               int64
	outputMIMEType     string
	outputQuality      int
	addWatermark       bool
	enhancePrompt      bool
	personGeneration   string
	safetyFilterLevel  string
	includeRAIReason   bool
	includeSafetyAttrs bool

	// ── gemini GenerateContent generation config ─────────
	temperature     float64
	topP            float64
	topK            float64
	maxOutputTokens int
	gcSeed          int64

	// Safety settings for GenerateContent (per-category)
	safetyHarassment       string
	safetyHateSpeech       string
	safetySexuallyExplicit string
	safetyDangerous        string

	verbose bool
}

func parseFlags() config {
	var cfg config

	// ── Required ──────────────────────────────────────────
	flag.StringVar(&cfg.prompt, "prompt", "", "[required] Text prompt for image generation")

	// ── Input image ───────────────────────────────────────
	flag.StringVar(&cfg.imagePath, "image", "", "Path to an input image (enables image-to-image / editing mode)")

	// ── Output ────────────────────────────────────────────
	flag.StringVar(&cfg.outputDir, "out-dir", ".", "Directory to save generated images")
	flag.StringVar(&cfg.outputName, "out-name", "", "Base filename (no extension). Default: generated_<timestamp>")

	// ── Mode & model ──────────────────────────────────────
	flag.StringVar(&cfg.mode, "mode", "gemini",
		"API mode:\n"+
			"  \"gemini\" (default) — GenerateContent; supports Nano Banana family + image-to-image\n"+
			"  \"imagen\"            — GenerateImages; supports aspect-ratio, seed, watermark")
	flag.StringVar(&cfg.model, "model", "gemini-3.1-flash-image-preview",
		"Model ID.\n"+
			"  Nano Banana family (gemini mode):\n"+
			"    gemini-3.1-flash-image-preview  — Nano Banana 2 (default): fast, high-volume\n"+
			"    gemini-3-pro-image-preview      — Nano Banana Pro: best quality, text rendering\n"+
			"    gemini-2.5-flash-image          — Nano Banana: speed/efficiency optimised\n"+
			"  Imagen (imagen mode):\n"+
			"    imagen-3.0-generate-002         — Imagen 3\n"+
			"    imagen-3.0-fast-generate-001    — Imagen 3 Fast\n"+
			"    imagen-4.0-generate-001         — Imagen 4\n"+
			"    imagen-4.0-fast-generate-001    — Imagen 4 Fast\n"+
			"    imagen-4.0-ultra-generate-001   — Imagen 4 Ultra")

	// ── Imagen-specific ───────────────────────────────────
	flag.StringVar(&cfg.aspectRatio, "aspect-ratio", "4:5",
		"(imagen) Aspect ratio: 1:1 | 3:4 | 4:3 | 9:16 | 16:9")
	flag.IntVar(&cfg.numberOfImages, "num-images", 1,
		"(imagen) Number of images to generate (1-4)")
	flag.StringVar(&cfg.negativePrompt, "negative-prompt", "",
		"(imagen) Description of what to discourage")
	flag.Float64Var(&cfg.guidanceScale, "guidance-scale", -1,
		"(imagen) Prompt adherence strength; higher = more literal. -1 = model default")
	flag.Int64Var(&cfg.seed, "seed", -1,
		"(imagen) Fixed seed for reproducibility. -1 = random. Note: incompatible with -add-watermark")
	flag.StringVar(&cfg.outputMIMEType, "output-mime", "image/png",
		"(imagen) Output MIME type: image/png | image/jpeg | image/webp")
	flag.IntVar(&cfg.outputQuality, "output-quality", -1,
		"(imagen) JPEG compression quality 0-100. -1 = default (only for image/jpeg)")
	flag.BoolVar(&cfg.addWatermark, "add-watermark", false,
		"(imagen) Add SynthID watermark to generated images")
	flag.BoolVar(&cfg.enhancePrompt, "enhance-prompt", false,
		"(imagen) Let model rewrite and improve your prompt (not supported on Gemini API backend)")
	flag.StringVar(&cfg.personGeneration, "person-generation", "ALLOW_ALL",
		"(imagen) Person generation policy: DONT_ALLOW | ALLOW_ADULT | ALLOW_ALL")
	flag.StringVar(&cfg.safetyFilterLevel, "safety-filter-level", "BLOCK_NONE",
		"(imagen) Global safety filter level: BLOCK_LOW_AND_ABOVE | BLOCK_MEDIUM_AND_ABOVE | BLOCK_ONLY_HIGH | BLOCK_NONE")
	flag.BoolVar(&cfg.includeRAIReason, "include-rai-reason", true,
		"(imagen) Include RAI filter reason when an image is filtered")
	flag.BoolVar(&cfg.includeSafetyAttrs, "include-safety-attrs", true,
		"(imagen) Include safety attribute scores in response")

	// ── Gemini GenerateContent tuning ─────────────────────
	flag.Float64Var(&cfg.temperature, "temperature", 1,
		"(gemini) Sampling temperature 0.0-2.0. -1 = model default")
	flag.Float64Var(&cfg.topP, "top-p", 0.95,
		"(gemini) Top-P nucleus sampling 0.0-1.0. -1 = model default")
	flag.Float64Var(&cfg.topK, "top-k", -1,
		"(gemini) Top-K sampling. -1 = model default")
	flag.IntVar(&cfg.maxOutputTokens, "max-output-tokens", -1,
		"(gemini) Max output tokens. -1 = model default")
	flag.Int64Var(&cfg.gcSeed, "gc-seed", -1,
		"(gemini) Random seed for GenerateContent. -1 = random")

	// ── Gemini safety settings (per-category) ─────────────
	flag.StringVar(&cfg.safetyHarassment, "safety-harassment", "BLOCK_NONE",
		"(gemini) Harassment threshold: OFF | BLOCK_NONE | BLOCK_ONLY_HIGH | BLOCK_MEDIUM_AND_ABOVE | BLOCK_LOW_AND_ABOVE")
	flag.StringVar(&cfg.safetyHateSpeech, "safety-hate-speech", "BLOCK_NONE",
		"(gemini) Hate-speech threshold (same values)")
	flag.StringVar(&cfg.safetySexuallyExplicit, "safety-sexually-explicit", "BLOCK_NONE",
		"(gemini) Sexually-explicit threshold (same values)")
	flag.StringVar(&cfg.safetyDangerous, "safety-dangerous", "BLOCK_NONE",
		"(gemini) Dangerous-content threshold (same values)")

	flag.BoolVar(&cfg.verbose, "verbose", false, "Print extra model text responses to stdout")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Nano Banana Image Generator — powered by Google Gemini / Imagen")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  Nano Banana 2   (default) → gemini-3.1-flash-image-preview")
		fmt.Fprintln(os.Stderr, "  Nano Banana Pro           → gemini-3-pro-image-preview")
		fmt.Fprintln(os.Stderr, "  Nano Banana               → gemini-2.5-flash-image")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  imggen -prompt \"<text>\" [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Set your API key:  $env:GEMINI_API_KEY = \"your_key\"  (PowerShell)")
		fmt.Fprintln(os.Stderr, "                   export GEMINI_API_KEY=your_key      (bash)")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.prompt == "" {
		fmt.Fprintln(os.Stderr, "error: -prompt is required")
		flag.Usage()
		os.Exit(1)
	}

	if cfg.outputName == "" {
		cfg.outputName = "generated_" + time.Now().Format("20060102_150405")
	}

	return cfg
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func detectMIMEType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	f, err := os.Open(path)
	if err != nil {
		return "image/png"
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	return http.DetectContentType(buf[:n])
}

func mimeToExt(mimeType string) string {
	switch strings.Split(mimeType, ";")[0] {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

// rawImageBytes handles the fact that some SDK versions return raw bytes and
// others return base64-encoded bytes for InlineData.Data.
func rawImageBytes(data []byte) []byte {
	if len(data) >= 4 {
		switch {
		case data[0] == 0x89 && data[1] == 'P': // PNG
			return data
		case data[0] == 0xFF && data[1] == 0xD8: // JPEG
			return data
		case data[0] == 'G' && data[1] == 'I' && data[2] == 'F': // GIF
			return data
		case data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F': // WEBP
			return data
		}
	}
	if dec, err := base64.StdEncoding.DecodeString(string(data)); err == nil {
		return dec
	}
	if dec, err := base64.RawStdEncoding.DecodeString(string(data)); err == nil {
		return dec
	}
	return data
}

func saveImage(dir, baseName string, idx int, ext string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := baseName
	if idx > 0 {
		name = fmt.Sprintf("%s_%d", baseName, idx+1)
	}
	path := filepath.Join(dir, name+ext)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

// ─────────────────────────────────────────────
// Imagen mode — GenerateImages
// ─────────────────────────────────────────────

func runImagen(ctx context.Context, client *genai.Client, cfg config) error {
	icfg := &genai.GenerateImagesConfig{
		AspectRatio:             cfg.aspectRatio,
		NumberOfImages:          int32(cfg.numberOfImages),
		NegativePrompt:          cfg.negativePrompt,
		OutputMIMEType:          cfg.outputMIMEType,
		AddWatermark:            cfg.addWatermark,
		PersonGeneration:        genai.PersonGeneration(strings.ToUpper(cfg.personGeneration)),
		SafetyFilterLevel:       genai.SafetyFilterLevel(strings.ToUpper(cfg.safetyFilterLevel)),
		IncludeRAIReason:        cfg.includeRAIReason,
		IncludeSafetyAttributes: cfg.includeSafetyAttrs,
	}
	if cfg.seed >= 0 {
		v := int32(cfg.seed)
		icfg.Seed = &v
	}
	if cfg.guidanceScale >= 0 {
		v := float32(cfg.guidanceScale)
		icfg.GuidanceScale = &v
	}
	if cfg.outputQuality >= 0 {
		v := int32(cfg.outputQuality)
		icfg.OutputCompressionQuality = &v
	}

	fmt.Printf("→ Mode        : imagen (GenerateImages)\n")
	fmt.Printf("→ Model       : %s\n", cfg.model)
	fmt.Printf("→ Aspect ratio: %s\n", cfg.aspectRatio)
	fmt.Printf("→ Num images  : %d\n", cfg.numberOfImages)
	fmt.Printf("→ Output MIME : %s\n", cfg.outputMIMEType)
	fmt.Printf("→ Safety level: %s\n", cfg.safetyFilterLevel)
	fmt.Printf("→ Person gen  : %s\n", cfg.personGeneration)
	if cfg.seed >= 0 {
		fmt.Printf("→ Seed        : %d\n", cfg.seed)
	}
	fmt.Println("→ Generating…")

	result, err := client.Models.GenerateImages(ctx, cfg.model, cfg.prompt, icfg)
	if err != nil {
		return fmt.Errorf("GenerateImages: %w", err)
	}

	if len(result.GeneratedImages) == 0 {
		return fmt.Errorf("no images returned — the request may have been filtered")
	}

	ext := mimeToExt(cfg.outputMIMEType)
	for i, img := range result.GeneratedImages {
		if img.RAIFilteredReason != "" {
			fmt.Printf("⚠  Image %d filtered — reason: %s\n", i+1, img.RAIFilteredReason)
			continue
		}
		imgBytes := rawImageBytes(img.Image.ImageBytes)
		path, err := saveImage(cfg.outputDir, cfg.outputName, i, ext, imgBytes)
		if err != nil {
			return fmt.Errorf("saving image %d: %w", i+1, err)
		}
		fmt.Printf("✓ Saved → %s (%d bytes)\n", path, len(imgBytes))

		if cfg.includeSafetyAttrs && img.SafetyAttributes != nil {
			fmt.Printf("  Safety scores: %v\n", img.SafetyAttributes)
		}
	}
	return nil
}

// ─────────────────────────────────────────────
// Gemini mode — GenerateContent with image modality
// ─────────────────────────────────────────────

func buildSafetySettings(cfg config) []*genai.SafetySetting {
	type pair struct{ cat, val string }
	pairs := []pair{
		{"HARM_CATEGORY_HARASSMENT", cfg.safetyHarassment},
		{"HARM_CATEGORY_HATE_SPEECH", cfg.safetyHateSpeech},
		{"HARM_CATEGORY_SEXUALLY_EXPLICIT", cfg.safetySexuallyExplicit},
		{"HARM_CATEGORY_DANGEROUS_CONTENT", cfg.safetyDangerous},
	}
	var ss []*genai.SafetySetting
	for _, p := range pairs {
		if p.val == "" {
			continue
		}
		ss = append(ss, &genai.SafetySetting{
			Category:  genai.HarmCategory(p.cat),
			Threshold: genai.HarmBlockThreshold(strings.ToUpper(p.val)),
		})
	}
	return ss
}

func runGemini(ctx context.Context, client *genai.Client, cfg config) error {
	gcfg := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
		SafetySettings:     buildSafetySettings(cfg),
	}

	if cfg.temperature >= 0 {
		v := float32(cfg.temperature)
		gcfg.Temperature = &v
	}
	if cfg.topP >= 0 {
		v := float32(cfg.topP)
		gcfg.TopP = &v
	}
	if cfg.topK >= 0 {
		v := float32(cfg.topK)
		gcfg.TopK = &v
	}
	if cfg.maxOutputTokens >= 0 {
		gcfg.MaxOutputTokens = int32(cfg.maxOutputTokens)
	}
	if cfg.gcSeed >= 0 {
		v := int32(cfg.gcSeed)
		gcfg.Seed = &v
	}

	// Build content parts
	parts := []*genai.Part{genai.NewPartFromText(cfg.prompt)}
	if cfg.imagePath != "" {
		imgBytes, err := os.ReadFile(cfg.imagePath)
		if err != nil {
			return fmt.Errorf("reading input image %q: %w", cfg.imagePath, err)
		}
		mimeType := detectMIMEType(cfg.imagePath)
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{MIMEType: mimeType, Data: imgBytes},
		})
		fmt.Printf("→ Input image : %s (%s, %d bytes)\n", cfg.imagePath, mimeType, len(imgBytes))
	}
	contents := []*genai.Content{genai.NewContentFromParts(parts, genai.RoleUser)}

	fmt.Printf("→ Mode        : gemini (GenerateContent)\n")
	fmt.Printf("→ Model       : %s\n", cfg.model)
	if len(gcfg.SafetySettings) > 0 {
		fmt.Printf("→ Safety settings (%d configured):\n", len(gcfg.SafetySettings))
		for _, s := range gcfg.SafetySettings {
			fmt.Printf("     %-40s → %s\n", s.Category, s.Threshold)
		}
	}
	fmt.Println("→ Generating…")

	result, err := client.Models.GenerateContent(ctx, cfg.model, contents, gcfg)
	if err != nil {
		return fmt.Errorf("GenerateContent: %w", err)
	}

	// ── Always dump usage ──────────────────────────────────
	if result.UsageMetadata != nil {
		fmt.Printf("→ Tokens used : prompt=%d output=%d total=%d\n",
			result.UsageMetadata.PromptTokenCount,
			result.UsageMetadata.CandidatesTokenCount,
			result.UsageMetadata.TotalTokenCount)
	}

	// ── Prompt-level block ────────────────────────────────
	if result.PromptFeedback != nil && result.PromptFeedback.BlockReason != "" {
		return fmt.Errorf("prompt blocked — reason: %s", result.PromptFeedback.BlockReason)
	}

	fmt.Printf("→ Candidates  : %d\n", len(result.Candidates))

	imgIdx := 0
	for ci, candidate := range result.Candidates {
		fmt.Printf("→ Candidate %d : finish_reason=%s parts=%d\n",
			ci+1, candidate.FinishReason, len(candidate.Content.Parts))

		// Always print safety ratings
		for _, r := range candidate.SafetyRatings {
			fmt.Printf("     safety: %-45s prob=%-10s blocked=%v\n",
				r.Category, r.Probability, r.Blocked)
		}

		if candidate.FinishReason == "SAFETY" {
			fmt.Printf("⚠  Candidate %d blocked (SAFETY)\n", ci+1)
			continue
		}

		for pi, part := range candidate.Content.Parts {
			if part.Text != "" {
				// Always print model text (not just in verbose mode) so we can see refusals
				fmt.Printf("   part[%d] text (%d chars): %s\n", pi, len(part.Text), part.Text)
			}
			if part.InlineData != nil {
				mtype := part.InlineData.MIMEType
				if mtype == "" {
					mtype = "image/png"
				}
				fmt.Printf("   part[%d] image: mime=%s raw_bytes=%d\n", pi, mtype, len(part.InlineData.Data))
				ext := mimeToExt(mtype)
				imgBytes := rawImageBytes(part.InlineData.Data)
				path, err := saveImage(cfg.outputDir, cfg.outputName, imgIdx, ext, imgBytes)
				if err != nil {
					return fmt.Errorf("saving image %d: %w", imgIdx+1, err)
				}
				fmt.Printf("✓ Saved → %s (%d bytes, %s)\n", path, len(imgBytes), mtype)
				imgIdx++
			}
			if part.Text == "" && part.InlineData == nil {
				// Dump the raw part so we can see what the model sent
				raw, _ := json.Marshal(part)
				fmt.Printf("   part[%d] unknown: %s\n", pi, string(raw))
			}
		}
	}

	if imgIdx == 0 {
		// Print the full raw response for diagnosis
		raw, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("\n── Full response dump ──\n%s\n── End dump ──\n", string(raw))
		return fmt.Errorf("no images returned — see dump above")
	}
	return nil
}

// ─────────────────────────────────────────────
// Entry point
// ─────────────────────────────────────────────

func main() {
	cfg := parseFlags()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is not set.\nGet your key at https://aistudio.google.com/")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("creating Gemini client: %v", err)
	}

	switch strings.ToLower(cfg.mode) {
	case "imagen":
		if err := runImagen(ctx, client, cfg); err != nil {
			log.Fatal(err)
		}
	case "gemini":
		if err := runGemini(ctx, client, cfg); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown mode %q — use \"gemini\" or \"imagen\"", cfg.mode)
	}
}
