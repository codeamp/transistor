package transistor

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	json "github.com/bww/go-json"
	log "github.com/codeamp/logger"
	uuid "github.com/satori/go.uuid"
)

type Action string
type State string
type EventName string

type Event struct {
	ID       uuid.UUID `json:"id"`
	ParentID uuid.UUID `json:"parentId"`
	Name     string    `json:"name"`

	Action       Action `json:"action"`
	State        State  `json:"state"`
	StateMessage string `json:"stateMessage"`

	Payload      interface{} `json:"payload"`
	PayloadModel string      `json:"payloadModel"`
	CreatedAt    time.Time   `json:"createdAt"`
	Caller       Caller      `json:"caller"`
	Artifacts    []Artifact  `json:"artifacts"`
}

type Caller struct {
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
}

type Artifact struct {
	Source string      `json:"source,omitempty"`
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Secret bool        `json:"secret"`
}

func (a *Artifact) String() string {
	return a.Value.(string)
}

func (a *Artifact) Int() int {
	i, err := strconv.Atoi(a.Value.(string))
	if err != nil {
		log.Error(err)
	}

	return i
}

func (a *Artifact) StringMap() map[string]interface{} {
	return a.Value.(map[string]interface{})
}

func (a *Artifact) StringSlice() []interface{} {
	return a.Value.([]interface{})
}

func NewEvent(event EventName, action Action, payload interface{}) Event {
	event := Event{
		ID:           uuid.NewV4(),
		Payload:      payload,
		CreatedAt:    time.Now(),
		Action:       action,
		State:        state,
		StateMessage: stateMessage,
	}

	if payload != nil {
		event.PayloadModel = reflect.TypeOf(payload).String()

		slug := reflect.ValueOf(payload)
		event.Name = fmt.Sprintf("%s:%s:%s", event.PayloadModel, action, slug.FieldByName("Slug"))
	} else {
		event.Name = string(action)
	}

	// for debugging purposes
	_, file, no, ok := runtime.Caller(1)
	if ok {
		event.Caller = Caller{
			File:       file,
			LineNumber: no,
		}
	}

	return event
}

func (e *Event) NewEvent(action Action, state State, stateMessage string, payload interface{}) Event {
	event := NewEvent(action, state, stateMessage, payload)
	event.ParentID = e.ID
	return evt
}

func (e *Event) Dump() {
	event, _ := json.MarshalRole("dummy", e)
	log.Info(string(event))
}

func (e *Event) Matches(name string) bool {
	matched, err := regexp.MatchString(name, e.Name)
	if err != nil {
		log.ErrorWithFields("Event regex match encountered an error", log.Fields{
			"regex":  name,
			"string": e.Name,
			"error":  err,
		})
	}

	if matched {
		return true
	}

	log.DebugWithFields("Event regex not matched", log.Fields{
		"regex":  name,
		"string": e.Name,
	})

	return false
}

func (e *Event) AddArtifact(key string, value interface{}, secret bool) {
	artifact := Artifact{
		Key:    key,
		Value:  value,
		Secret: secret,
	}

	exists := false
	for i, _artifact := range e.Artifacts {
		if strings.ToLower(_artifact.Key) == strings.ToLower(key) {
			exists = true
			e.Artifacts[i] = artifact
		}
	}

	if !exists {
		e.Artifacts = append(e.Artifacts, artifact)
	}
}

func (e *Event) GetArtifact(key string) (Artifact, error) {
	for _, artifact := range e.Artifacts {
		if artifact.Source == "" && strings.ToLower(artifact.Key) == strings.ToLower(key) {
			return artifact, nil
		}
	}

	return Artifact{}, errors.New(fmt.Sprintf("Artifact %s not found", key))
}

func (e *Event) GetArtifactFromSource(key string, source string) (Artifact, error) {
	for _, artifact := range e.Artifacts {
		if strings.ToLower(artifact.Source) == strings.ToLower(source) && strings.ToLower(artifact.Key) == strings.ToLower(key) {
			return artifact, nil
		}
	}

	return Artifact{}, errors.New(fmt.Sprintf("Artifact %s not found", key))
}
