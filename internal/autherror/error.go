package autherror

import (
	"encoding/json"
	"strings"
)

const envelopeType = "auth_error"

type Envelope struct {
	Type    string   `json:"type" yaml:"type"`
	Code    string   `json:"code" yaml:"code"`
	Message string   `json:"message" yaml:"message"`
	Hint    string   `json:"hint,omitempty" yaml:"hint,omitempty"`
	Hints   []string `json:"hints,omitempty" yaml:"hints,omitempty"`
}

type Error struct {
	envelope Envelope
}

func New(code, message, hint string) *Error {
	if strings.TrimSpace(hint) == "" {
		return NewWithHints(code, message)
	}
	return NewWithHints(code, message, hint)
}

func NewWithHints(code, message string, hints ...string) *Error {
	normalized := normalizeHints(hints)
	envelope := Envelope{
		Type:    envelopeType,
		Code:    code,
		Message: message,
		Hints:   normalized,
	}
	if len(normalized) > 0 {
		envelope.Hint = normalized[0]
	}
	return &Error{
		envelope: envelope,
	}
}

func (e *Error) Envelope() Envelope {
	if e == nil {
		return Envelope{}
	}
	return e.envelope
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.envelope.Hint == "" {
		return e.envelope.Message
	}
	return e.envelope.Message + ": " + e.envelope.Hint
}

func (e *Error) RawJSON() string {
	if e == nil {
		return "{}"
	}
	body, err := json.Marshal(e.envelope)
	if err != nil {
		return `{"type":"auth_error","code":"serialization_failed","message":"failed to serialize auth error"}`
	}
	return string(body)
}

func normalizeHints(hints []string) []string {
	if len(hints) == 0 {
		return nil
	}

	result := make([]string, 0, len(hints))
	seen := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		hint = strings.TrimSpace(hint)
		if hint == "" {
			continue
		}
		if _, ok := seen[hint]; ok {
			continue
		}
		seen[hint] = struct{}{}
		result = append(result, hint)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
