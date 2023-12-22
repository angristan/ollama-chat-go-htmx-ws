# Ollama Chat

Just a little PoC of a chat app running an LLM locally on Ollama. Just an excuse to have fun with websockets, htmx and Go.

![](https://i.imgur.com/KUoaHPU.png)

Using:

- The [Ollama Chat API](https://github.com/jmorganca/ollama/blob/b80081022f9460df682116af08fe4687172b23ae/docs/api.md#generate-a-chat-completion)
  - With the [official Go client](https://github.com/jmorganca/ollama/blob/b80081022f9460df682116af08fe4687172b23ae/api/client.go)
  - Running the [Mistral model](https://ollama.ai/library/mistral) locally
  - The chat history is stored in-memory and sent to Ollama to have a working context
- [htmx](https://htmx.org/) for the frontend interactions
- Go's [text/template](https://pkg.go.dev/text/template) for the HTML templating
- Websocket for the realtime communication
  - The htmx [WebSockets](https://htmx.org/extensions/web-sockets/) extension is used to handle the connection, send and receive messages
  - [Melody](https://github.com/olahol/melody) for websocket handling on the backend
- TailwindCSS for the styling
  - Chat template found on [Codepen](https://codepen.io/robstinson/pen/oNLaLMN)
- [DiceBear](https://www.dicebear.com/) for the funny avatars
