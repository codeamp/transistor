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
	Name     EventName `json:"name"`

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

func NewEvent(eventName EventName, action Action, payload interface{}) Event {
	event := Event{
		ID:           uuid.NewV4(),
		Name:         eventName,
		Payload:      payload,
		CreatedAt:    time.Now(),
		Action:       action,
		State:        State("waiting"),
		StateMessage: "Waiting for event to run",
	}

	if payload != nil {
		event.PayloadModel = reflect.TypeOf(payload).String()
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

func (e *Event) NewEvent(action Action, payload interface{}) Event {
	event := NewEvent(e.Name, action, payload)
	event.ParentID = e.ID
	event.State = State("running")
	event.StateMessage = fmt.Sprintf("Running event '%s'", e.Name)
	return event
}

func (e *Event) SetState(state State, stateMessage string) {
	e.State = state
	e.StateMessage = stateMessage
}

func (e *Event) Dump() {
	event, _ := json.MarshalRole("dummy", e)
	log.Info(string(event))
}

func (e *Event) Event() string {
	return fmt.Sprintf("%s:%s", e.Name, e.Action)
}

func (e *Event) Matches(name string) bool {
	matched, err := regexp.MatchString(name, e.Event())
	if err != nil {
		log.ErrorWithFields("Event regex match encountered an error", log.Fields{
			"regex":  name,
			"string": e.Event(),
			"error":  err,
		})
	}

	if matched {
		return true
	}

	// Not that important because there will be events that will
	// fail without being an error condition because there will obviously
	// be some events that do not match. Leaving here for future debugging, but disabling for sake of DEBUG channel
	// ADB
	// log.DebugWithFields("Event regex not matched", log.Fields{
	// 	"regex":  name,
	// 	"string": e.Event(),
	// })

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
