package graph

import (
	gocontext "context"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const localEmbeddingDims = 384
const openAIEmbeddingDims = 1536

const (
	localEmbeddingProvider  = "local-hash-v1"
	openAIEmbeddingProvider = "openai:text-embedding-3-small"
)

var tokenPattern = regexp.MustCompile(`[a-z0-9]+(?:-[a-z0-9]+)*`)

type embeddingSpec struct {
	Provider string
	Dims     int
	Explicit bool
}

// getEmbedding calls the OpenAI text-embedding-3-small API and returns a float32 vector.
func getEmbedding(apiKey, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}
	client := openai.NewClient(apiKey)
	resp, err := client.CreateEmbeddings(gocontext.Background(), openai.EmbeddingRequestStrings{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding API: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return resp.Data[0].Embedding, nil
}

// embedText returns an embedding for text. OpenAI is used when configured; a
// deterministic local lexical embedding keeps graph/tripwire behavior available
// in offline demos and tests.
func embedText(text string) ([]float32, error) {
	spec := desiredEmbeddingSpec()
	return embedTextWithSpec(text, spec)
}

func desiredEmbeddingSpec() embeddingSpec {
	provider, explicit := os.LookupEnv("GITWHY_EMBEDDING_PROVIDER")
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "openai":
		return embeddingSpec{Provider: openAIEmbeddingProvider, Dims: openAIEmbeddingDims, Explicit: explicit}
	case "local":
		return embeddingSpec{Provider: localEmbeddingProvider, Dims: localEmbeddingDims, Explicit: explicit}
	case "":
		if os.Getenv("OPENAI_API_KEY") != "" {
			return embeddingSpec{Provider: openAIEmbeddingProvider, Dims: openAIEmbeddingDims, Explicit: false}
		}
		return embeddingSpec{Provider: localEmbeddingProvider, Dims: localEmbeddingDims, Explicit: false}
	default:
		return embeddingSpec{Provider: provider, Dims: 0, Explicit: explicit}
	}
}

func embedTextWithSpec(text string, spec embeddingSpec) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	switch spec.Provider {
	case localEmbeddingProvider:
		return localEmbedding(text), nil
	case openAIEmbeddingProvider:
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		embedding, err := getEmbedding(apiKey, text)
		if err != nil {
			return nil, err
		}
		if len(embedding) != spec.Dims {
			return nil, fmt.Errorf("embedding provider %s returned %d dims, expected %d", spec.Provider, len(embedding), spec.Dims)
		}
		return embedding, nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider %q", spec.Provider)
	}
}

func localEmbedding(text string) []float32 {
	tokens := tokenizeText(text)
	vec := make([]float32, localEmbeddingDims)
	for i, token := range tokens {
		addHashedWeight(vec, token, 1.0)
		if i > 0 {
			addHashedWeight(vec, tokens[i-1]+" "+token, 0.55)
		}
	}
	normalize(vec)
	return vec
}

func tokenizeText(text string) []string {
	text = strings.ToLower(text)
	return tokenPattern.FindAllString(text, -1)
}

func addHashedWeight(vec []float32, token string, weight float32) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(token))
	idx := int(h.Sum32() % uint32(len(vec)))
	vec[idx] += weight
}

func normalize(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	if sum == 0 {
		return
	}
	scale := float32(1 / math.Sqrt(sum))
	for i := range vec {
		vec[i] *= scale
	}
}

