package cli

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/commerce-cloud-integrations/emporix-cli/internal/emporix"
)

// What consumes the AI credits of the AI Service: text generation,
// completions and chats with agents, and the attachments of a chat.

// openAICredits says what text generation and completions consume.
const openAICredits = "It consumes AI credits: the LLM tokens of the tenant's own OpenAI account where the tenant\n" +
	"configuration openAiApiToken holds its token, or else Emporix's, of which a tenant may use\n" +
	"3 million input and 1 million output tokens."

// agentCredits says what a chat with an agent consumes.
const agentCredits = "It consumes AI credits: the LLM tokens of the agent's model, and of the agents it\n" +
	"collaborates with."

// replyTimeout says that a reply, which Emporix sends once it is generated,
// may take longer than the timeout of a request.
func replyTimeout() string {
	return "A reply may take longer than the --timeout of a request (" + defaultTimeout.String() +
		" by default): then give a longer\none, e.g. --timeout 5m"
}

// sessionFlag is the flag that names the session of a chat or attachment.
const sessionFlag = "session"

// sessionHeader is the header of a chat or attachment that names its
// session, which --session sets; a value that cannot be sent as a header is
// refused while flags are parsed.
func sessionHeader() paramFlag {
	return paramFlag{param: "session-id", flag: sessionFlag, usage: "the `ID` of the session to go on with, " +
		"the sessionId of an earlier chat or attachment; without it, a new one", check: emporix.CheckHeaderValue}
}

// messageRoles are the roles of the messages of a completion.
func messageRoles() []string { return []string{"SYSTEM", "USER", "ASSISTANT"} }

// completionMessage is a message of the chat history of a completion.
type completionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// parseMessage parses a --message of a completion, ROLE:TEXT, e.g.
// "USER:Describe product 100200", whose role is one of messageRoles in any
// case, sent upper-cased.
func parseMessage(value string) (completionMessage, error) {
	role, text, found := strings.Cut(value, ":")
	role, text = strings.ToUpper(strings.TrimSpace(role)), strings.TrimSpace(text)
	if !found || text == "" || !slices.Contains(messageRoles(), role) {
		return completionMessage{}, errors.New("use ROLE:TEXT, with ROLE " + alternatives(messageRoles()))
	}
	return completionMessage{Role: role, Content: text}, nil
}

// chatAttachment is a file that a chat sends to its agent, attached first.
type chatAttachment struct {
	ID string `json:"attachmentId"`
}

// parseAttachment parses an --attachment of a chat, the ID of an attachment.
func parseAttachment(value string) (chatAttachment, error) {
	if value == "" {
		return chatAttachment{}, errors.New("empty")
	}
	return chatAttachment{ID: value}, nil
}

// chatBody is the body of a chat with an agent that its flags make: the
// agent, the message and the attachments that it sends, which are those of a
// session, so --attachment applies only with --session.
func chatBody() bodyBinder {
	return objectBody{
		stringField("agentId", "agent", "the `ID` of the agent, as `epx ai agents list` shows it").require(),
		stringField("message", "message", "the `TEXT` to send to the agent").require(),
		bodyField{member: "attachments", value: objectsValue("attachment", "the `ID` of an attachment of the "+
			"session to send with the message, as `epx ai agents attach` prints it (repeatable)",
			parseAttachment).optional()}.onlyWith(sessionFlag),
	}
}

// chatHelp says how a chat goes on in a session, and with attachments.
const chatHelp = "With --session, the chat goes on in that session, which the agent remembers where its\n" +
	"enabledMemory is true; an attachment must be one of the session, attached to the same agent\n" +
	"(`epx ai agents attach`). --file gives the body instead, with the caption and purpose of\n" +
	"each attachment."

// agentChat sends a message to an agent at agentic/<use>, which Emporix
// answers as answer says, following "Sends the message ... to the agent
// --agent, ": chat, with the agent's reply, and chat-async, with a job.
func agentChat(use, short, answer string) *write {
	return &write{
		use:   use,
		short: short + " (consumes AI credits)",
		long: "Sends the message that --message gives to the agent --agent, " + answer + "\n" + agentCredits + "\n" +
			chatHelp,
		example: "  epx ai " + use + " --agent complaint-agent --message \"Find order EON1243\" --yes",
		method:  http.MethodPost,
		req:     request{path: agenticPath + use},
		scopes:  []string{aiChatScope},
		body:    &payload{typed: chatBody()},
		headers: []paramFlag{sessionHeader()},
	}
}

// textGeneration generates a text from a prompt, which consumes AI credits.
func textGeneration() *write {
	return &write{
		use:   "generate-text",
		short: "Generate a text from a prompt (consumes AI credits)",
		long: "Generates a text from the prompt that --prompt gives, such as a product description, and\n" +
			"prints it as result.\n" + openAICredits + "\n" + replyTimeout() + ".",
		example: "  epx ai generate-text --prompt \"Describe product 100200 in German\" --max-tokens 500",
		method:  http.MethodPost,
		req:     request{path: aiPath + "texts"},
		scopes:  []string{aiTextScope},
		body: &payload{typed: objectBody{
			stringField("prompt", "prompt", "the `TEXT` that asks for the text").require(),
			countField("maxTokens", "max-tokens", "the most LLM `TOKENS` to use (Emporix: 4097 without it)"),
		}},
	}
}

// completionGeneration generates the next message of a chat history, which
// consumes AI credits.
func completionGeneration() *write {
	return &write{
		use:   "generate-completion",
		short: "Generate the next message of a chat history (consumes AI credits)",
		long: "Generates the next message of the chat history that --message gives, one message per\n" +
			"--message in their order, and prints it as result.\n" + openAICredits + "\n" + replyTimeout() + ".",
		example: "  epx ai generate-completion --message \"SYSTEM:You write product texts\" \\\n" +
			"    --message \"USER:Describe product 100200\"",
		method: http.MethodPost,
		req:    request{path: aiPath + "completions"},
		scopes: []string{aiCompletionScope},
		body: &payload{typed: objectBody{bodyField{member: "messages", value: objectsValue("message",
			"a message of the chat history as `ROLE:TEXT`, ROLE "+alternatives(messageRoles())+" (repeatable)",
			parseMessage).require()}}},
	}
}

// mediaFlag is the flag of an attachment that names a media asset to assign.
const mediaFlag = "media"

// agentAttach attaches a file to an agent in a session: a file it uploads,
// or a media asset that Emporix stores already.
func agentAttach() *write {
	return &write{
		use:   "attach AGENT_ID",
		short: "Attach a file to a chat with an agent",
		long: "Uploads the file that --" + uploadFlag + " names (jpg, jpeg, png, pdf, csv, text or txt), or assigns the\n" +
			"media asset that --" + mediaFlag + " names (`epx media list`), to the agent AGENT_ID, the ID that\n" +
			"`epx ai agents list` shows, in the session that --session names, or a new one. Emporix\n" +
			"answers with the attachment's id and the sessionId, which `epx ai chat` with that agent\n" +
			"takes as --attachment and --session.",
		example: "  epx ai agents attach complaint-agent --upload order.pdf --yes",
		method:  http.MethodPost,
		req:     request{path: agenticPath + placeholderID + "/attachments"},
		scopes:  []string{aiChatScope},
		body: &formBody{form: form{
			uploadPart("attachment", "the file in `PATH` to upload ("+uploadLimitHelp+"), sent with its name"),
			fieldPart("attachmentId", mediaFlag, "the `ID` of a media asset to attach"),
		}, oneOf: []string{uploadFlag, mediaFlag}},
		headers: []paramFlag{sessionHeader()},
		ids:     argumentIDs{"AGENT_ID": aiAgents().ids()},
	}
}
