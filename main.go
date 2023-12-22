package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/jmorganca/ollama/api"
	"github.com/olahol/melody"
	"go.uber.org/zap"
)

type WebsocketMessage struct {
	UserMessage string `json:"user-message"`
}

type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message,omitempty"`

	Done bool `json:"done"`
}

type Message struct {
	ID      string `json:"id"`
	Role    string `json:"role"` // one of ["system", "user", "assistant"]
	Content string `json:"content"`
	Append  bool
}

var History []api.Message

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	ollamaClient, err := api.ClientFromEnvironment()
	if err != nil {
		sugar.Fatal(err)
	}

	ws := melody.New()

	// Render initial chat page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sugar.Info("Serving index.html")

		indexTemplate := template.Must(template.ParseFiles("index.html"))
		indexTemplate.Execute(w, nil)
	})

	// Not used yet
	http.HandleFunc("/models", func(w http.ResponseWriter, r *http.Request) {
		res, err := ollamaClient.List(context.TODO())
		if err != nil {
			panic(err)
		}

		respBytes, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBytes)
	})

	// Upgrades http requests to websocket connections and dispatches them to be handled by the melody instance
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		sugar.Info("Upgrading to websocket connection")
		ws.HandleRequest(w, r)
	})

	// Handles messages received from the websocket connection
	ws.HandleMessage(func(s *melody.Session, msg []byte) {
		sugar.Infow("Received message from websocket connection", "message", string(msg))

		var wsMessage WebsocketMessage
		if err := json.Unmarshal(msg, &wsMessage); err != nil {
			sugar.Errorw("Failed to unmarshal message", "error", err)
		}

		// Send back the user message to the websocket connection so that it can be displayed in the chat
		userMessage := api.Message{
			Role:    "user",
			Content: wsMessage.UserMessage,
		}

		var userMsgWSResp bytes.Buffer
		tmpl := template.Must(template.ParseFiles("index.html"))

		// Render chat div with single message, `hx-swap-oob="beforeend"`
		// will append the message to the chat
		tmpl.ExecuteTemplate(&userMsgWSResp, "chat-list", ChatResponse{
			CreatedAt: time.Now().Format(time.TimeOnly),
			Message: Message{
				ID:      "toto",
				Role:    "user",
				Content: wsMessage.UserMessage,
			},
		})

		sugar.Infow("Broadcasting message to websocket connections", "message", userMsgWSResp.String())
		ws.Broadcast(userMsgWSResp.Bytes())

		// Add the user message to the history for context
		History = append(History, userMessage)

		// Chat request to Ollama API
		stream := false
		ollamaChatRequest := &api.ChatRequest{
			Model:    "mistral",
			Messages: History,
			Stream:   &stream,
		}

		sugar.Infow("Requesting Ollama chat API", "request", ollamaChatRequest)

		err = ollamaClient.Chat(ctx, ollamaChatRequest, func(ollamaChatResponse api.ChatResponse) error {
			if ollamaChatResponse.Done {
				sugar.Infow("Chat request done", "response", ollamaChatResponse)

				// Keep the assistant message in the history for context
				History = append(History, *ollamaChatResponse.Message)

				var assistantMsgWSResp bytes.Buffer
				tmpl := template.Must(template.ParseFiles("index.html"))

				// Render chat div with single message, `hx-swap-oob="beforeend"`
				// will append the message to the chat
				tmpl.ExecuteTemplate(&assistantMsgWSResp, "chat-list", ChatResponse{
					CreatedAt: time.Now().Format(time.TimeOnly),
					Message: Message{
						ID:      "toto",
						Role:    "assistant",
						Content: ollamaChatResponse.Message.Content,
					},
				})

				sugar.Infow("Broadcasting message to websocket connections", "message", assistantMsgWSResp.String())
				ws.Broadcast(assistantMsgWSResp.Bytes())
			}

			return nil
		})
		if err != nil {
			sugar.Errorw("Failed to request Ollama chat API", "error", err)
		}
	})

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
