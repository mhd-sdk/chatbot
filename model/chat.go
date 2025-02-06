package model

import (
	"github.com/ollama/ollama/api"
	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	Messages []Message `json:"messages"`
}

type Message struct {
	gorm.Model `json:"-"`
	Role       string `json:"role"`
	Content    string `json:"content"`
	ChatID     uint   `json:"-"`
}

func (m *Chat) AddMessage(msg Message) {
	m.Messages = append(m.Messages, msg)
}

func (m *Chat) OllamaMessages() (messages []api.Message) {
	for _, msg := range m.Messages {
		messages = append(messages, api.Message{
			Role:      msg.Role,
			Content:   msg.Content,
			Images:    nil,
			ToolCalls: nil,
		})
	}
	return messages
}
