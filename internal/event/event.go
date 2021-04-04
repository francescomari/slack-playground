package event

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Callback interface {
	OnEvent(ctx context.Context, envelope *Envelope)
}

type Event struct {
	Type    string
	User    string
	Text    string
	Channel string
}

type Envelope struct {
	Token     string
	Challenge string
	Type      string
	Event     Event
}

type Handler struct {
	SigningSecret string
	Callback      Callback
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Events are always sent as POST requests.

	if r.Method != http.MethodPost {
		log.Printf("invalid method: %s", r.Method)
		return
	}

	// Check the timestamp to avoid replay attacks.

	timestamp := r.Header.Get("X-Slack-Request-Timestamp")
	if timestamp == "" {
		log.Printf("request timestamp not found")
		return
	}

	unix, err := strconv.Atoi(timestamp)
	if err != nil {
		log.Printf("invalid request timestamp: %v", err)
		return
	}

	if time.Unix(int64(unix), 0).Before(time.Now().Add(-5 * time.Minute)) {
		log.Printf("request timestamp too far in the past")
		return
	}

	// Compute the signature against the body and the timestamp and compare it
	// to the signature passed by Slack.

	signature := r.Header.Get("X-Slack-Signature")
	if signature == "" {
		log.Printf("request signature not found")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body: %v", err)
		return
	}

	if signature != h.computeSignature(timestamp, body) {
		log.Printf("invalid request signature")
		return
	}

	// The request is valid, dispatch the event. Only 'url_verification' events
	// are handled here, since they need to write the challenge back to the
	// caller.

	var e Envelope

	if err := json.Unmarshal(body, &e); err != nil {
		log.Printf("decode request: %v", err)
		return
	}

	switch e.Type {
	case "url_verification":
		fmt.Fprint(w, e.Challenge)
	default:

		// We shouldn't handle the event as part of the request handling code.
		// If the response to the event takes longer than 3s, Slack will
		// eventually rate limit the app. In a production-grade app, we should
		// decouple the request handling code and the event handling code. That
		// said, since this is not a production-grade app, we handle the event
		// here.

		h.Callback.OnEvent(r.Context(), &e)
	}
}

func (h *Handler) computeSignature(timestamp string, body []byte) string {
	hash := hmac.New(sha256.New, []byte(h.SigningSecret))
	hash.Write([]byte(fmt.Sprintf("v0:%s:%s", timestamp, string(body))))
	return "v0=" + hex.EncodeToString(hash.Sum(nil))
}
