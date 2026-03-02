package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ariolwork/collections/slices"
)

type answerRecord struct {
	Name string `json:name`
	Type int    `json:type`
	TTL  int    `json:TTL`
	Data string `json:data`
}

type answer struct {
	Answer []answerRecord `json:Answer`
}

const ClowdDns = "https://cloudflare-dns.com"

type Resolver interface {
	Resolve(ctx context.Context, name string) string
}

type client struct {
	dnsServerUrl string
	logger       *log.Logger
}

func New(log *log.Logger) Resolver {
	return &client{
		dnsServerUrl: ClowdDns,
		logger:       log,
	}
}

func (c *client) Resolve(ctx context.Context, name string) string {
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/dns-query?name=%s", c.dnsServerUrl, name), nil)
	request.Header.Add("accept", "application/dns-json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return name
	}

	body, err := io.ReadAll(response.Body)

	ips := &answer{}

	if err := json.Unmarshal(body, ips); err != nil {
		return name
	}

	namesTable := slices.ToMapBuckets(ips.Answer, func(r answerRecord) string {
		return r.Name
	})

	for {
		if data, ok := namesTable[name]; ok {
			name = data[0].Data
		} else {
			return name
		}
	}
}

func ToMap[S ~[]T, T any, K comparable](seq S, keySelector func(T) K) map[K]T {
	result := make(map[K]T, len(seq))

	for _, val := range seq {
		result[keySelector(val)] = val
	}

	return result
}
