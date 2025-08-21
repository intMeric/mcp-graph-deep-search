package object

import (
	"strings"
	"time"
	
	"mgds/pkg/node"
)

const URLNodeType = "url"

// URLNode represents a web page node in the graph
type URLNode struct {
	mgdsId      string
	displayName string
	description string
	url         string
	title       string
	content     string
	isScrapped  bool
	createdAt   time.Time
	updatedAt   time.Time
	properties  map[string]any
}

// NewURLNode creates a new URL node
func NewURLNode(mgdsId, displayName, description, url string) node.Interface {
	now := time.Now()
	return &URLNode{
		mgdsId:      mgdsId,
		displayName: displayName,
		description: description,
		url:         url,
		createdAt:   now,
		updatedAt:   now,
		properties:  make(map[string]any),
	}
}

// NewURLNodeWithDetails creates a new URL node with additional details
func NewURLNodeWithDetails(mgdsId, displayName, description, url, title, content string) node.Interface {
	now := time.Now()
	return &URLNode{
		mgdsId:      mgdsId,
		displayName: displayName,
		description: description,
		url:         url,
		title:       title,
		content:     content,
		createdAt:   now,
		updatedAt:   now,
		properties:  make(map[string]any),
	}
}

// Core interface methods
func (u *URLNode) GetMgdsId() string {
	return u.mgdsId
}

func (u *URLNode) GetDisplayName() string {
	return u.displayName
}

func (u *URLNode) GetDescription() string {
	return u.description
}

func (u *URLNode) GetType() string {
	return URLNodeType
}

// Generic properties
func (u *URLNode) GetProperty(key string) (any, bool) {
	if u.properties == nil {
		return nil, false
	}
	value, exists := u.properties[key]
	return value, exists
}

func (u *URLNode) SetProperty(key string, value any) {
	if u.properties == nil {
		u.properties = make(map[string]any)
	}
	u.properties[key] = value
	u.updatedAt = time.Now()
}

func (u *URLNode) IsValid() bool {
	return strings.TrimSpace(u.mgdsId) != "" &&
		strings.TrimSpace(u.displayName) != "" &&
		strings.TrimSpace(u.description) != "" &&
		strings.TrimSpace(u.url) != ""
}

// URLNode specific methods
func (u *URLNode) GetURL() string {
	return u.url
}

func (u *URLNode) SetURL(url string) {
	u.url = url
	u.updatedAt = time.Now()
}

func (u *URLNode) GetTitle() string {
	return u.title
}

func (u *URLNode) SetTitle(title string) {
	u.title = title
	u.updatedAt = time.Now()
}

func (u *URLNode) GetContent() string {
	return u.content
}

func (u *URLNode) SetContent(content string) {
	u.content = content
	u.updatedAt = time.Now()
}

func (u *URLNode) GetCreatedAt() time.Time {
	return u.createdAt
}

func (u *URLNode) GetUpdatedAt() time.Time {
	return u.updatedAt
}

func (u *URLNode) IsScrapped() bool {
	return u.isScrapped
}

func (u *URLNode) SetScrapped(scrapped bool) {
	u.isScrapped = scrapped
	u.updatedAt = time.Now()
}

// Serialize converts URLNode to map for storage
func (u *URLNode) Serialize() (map[string]any, error) {
	data := map[string]any{
		"mgdsId":      u.mgdsId,
		"displayName": u.displayName,
		"description": u.description,
		"type":        URLNodeType,
		"url":         u.url,
		"title":       u.title,
		"content":     u.content,
		"isScrapped":  u.isScrapped,
		"createdAt":   u.createdAt,
		"updatedAt":   u.updatedAt,
		"properties":  u.properties,
	}
	return data, nil
}

// Deserialize restores URLNode from map
func (u *URLNode) Deserialize(data map[string]any) error {
	if mgdsId, ok := data["mgdsId"].(string); ok {
		u.mgdsId = mgdsId
	}
	if displayName, ok := data["displayName"].(string); ok {
		u.displayName = displayName
	}
	if description, ok := data["description"].(string); ok {
		u.description = description
	}
	if url, ok := data["url"].(string); ok {
		u.url = url
	}
	if title, ok := data["title"].(string); ok {
		u.title = title
	}
	if content, ok := data["content"].(string); ok {
		u.content = content
	}
	if isScrapped, ok := data["isScrapped"].(bool); ok {
		u.isScrapped = isScrapped
	}
	if createdAt, ok := data["createdAt"].(time.Time); ok {
		u.createdAt = createdAt
	}
	if updatedAt, ok := data["updatedAt"].(time.Time); ok {
		u.updatedAt = updatedAt
	}
	if properties, ok := data["properties"].(map[string]any); ok {
		u.properties = properties
	}
	return nil
}