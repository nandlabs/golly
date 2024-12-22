package genai

import (
	"fmt"

	"oss.nandlabs.io/golly/ioutils"
)

const (
	// userInfoAttrName
	userInfoAttrName    = "user_info"
	userInfoFmtTemplate = `
	User Info:
	Name: %s
	Locale: %s
	Timezone: %s
	`
	previousQuestionsVar  = "PreviousQuestions"
	currentQuestionVar    = "CurrentQuestion"
	contextualiseTemplate = `
		You are an advanced assistant designed to analyze user queries in a chat session. Your task is to determine the intent of the user's current question based on the context of their previous questions. Use the provided history of user questions to identify the intent clearly. If the intent is unclear, suggest what additional context might be needed.

		Previous Questions:
		{{- range .PreviousQuestions }}
		- {{ . }}
		{{- end }}

		Current Question:
		{{ .CurrentQuestion }}

		Determine the user's intent for the current question. Provide a clear interpretation or suggest additional context if needed.
		`
)

// Session is the interface that represents a session
type Session interface {
	// Id returns the id of the session. This is expected to be unique.
	Id() string
	//Model returns the model of the session
	Model() Model
	// Attributes returns the attributes of the session
	Attributes() map[string]any

	// Last returns the current exchange of the session
	CurrentExchange() (Exchange, error)
	// Exchanges returns the exchanges of the session
	Exchanges() ([]Exchange, error)
	// Memory returns the memory of the session
	SaveExchange(exchange Exchange) error
	// Contextualise returns the contextualised message based on the last n exchanges
	Contextualise(text string, n int) (string, error)
}

// LocalSession is a local session
// It is a session that is stored in local physical memory

type LocalSession struct {
	id                    string
	model                 Model
	attributes            map[string]any
	memory                Memory
	contextualiseTemplate PromptTemplate
}

// Id returns the id of the session. This is expected to be unique.
func (s *LocalSession) Id() string {
	return s.id
}

// Model returns the model of the session
func (s *LocalSession) Model() Model {
	return s.model
}

// Attributes returns the attributes of the session
func (s *LocalSession) Attributes() map[string]any {
	return s.attributes
}

// CurrentExchange returns the current exchange of the session
func (s *LocalSession) CurrentExchange() (exchange Exchange, err error) {
	var exchanges []Exchange
	exchanges, err = s.memory.Last(s.id, 1)
	if err == nil && len(exchanges) > 0 {
		exchange = exchanges[0]
	}
	return

}

// Exchanges returns the exchanges of the session
func (s *LocalSession) Exchanges() (exchanges []Exchange, err error) {
	return s.memory.Last(s.id, -1)
}

// SaveExchange saves the exchange of the session
func (s *LocalSession) SaveExchange(exchange Exchange) (err error) {
	return s.memory.Add(s.id, exchange)
}

// Contextualise rewrites the query based on the last n exchanges
// The LocalSession uses only the user's previous queries to contextualise the current query
func (s *LocalSession) Contextualise(text string, n int) (newQuestion string, err error) {
	var exchanges []Exchange
	var previous []string
	newQuestion = text
	exchanges, err = s.memory.Last(s.id, n)
	if err != nil {
		return
	}
	if len(exchanges) > 0 {
		// current = exchanges[len(exchanges)-1].CurrentMessage().Text()
		for _, exg := range exchanges {
			if !exg.HasMsgsFrmActors(UserActor) {
				continue
			}

			msgs := exg.MsgsByActors(UserActor)
			for _, msg := range msgs {
				if msg.mimeType != ioutils.MimeTextPlain {
					continue
				}
				previous = append(previous, msg.String())
			}
			if msgs != nil {
				templateAttrs := make(map[string]any)
				templateAttrs[previousQuestionsVar] = previous
				templateAttrs[currentQuestionVar] = text
				exg := NewExchange("message-reformatter")
				textMsg, _ := exg.AddTxtMsg(text, UserActor)
				err = s.contextualiseTemplate.WriteTo(textMsg, templateAttrs)
				if err != nil {
					return
				}
				err = s.model.Generate(exg)
				if err != nil {
					return
				}
				reWrittenMsg := exg.MsgsByActors(AIActor)
				if reWrittenMsg != nil {
					newQuestion = reWrittenMsg[0].String()
				}
			}
		}

	}
	return
}

// UserInfo is a type that represents the user info
// Avoid any sensitive information in the user info
type UserInfo struct {
	Name     string
	Locale   string
	Timezone string
}

// String returns the string representation of the user info
func (u *UserInfo) String() string {
	// Retrun all user attributes with names if they are not empty in new lines
	return fmt.Sprintf(userInfoFmtTemplate, u.Name, u.Locale, u.Timezone)

}
