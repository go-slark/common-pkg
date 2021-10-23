package xevent

type Event interface {
	Topic() string
	EventId() string
	SetEventId(eventId string)
	Content() []byte
}

//type Entity interface {
//	//The unique ID of the entity.
//	Identity() string
//	//Returns the requested Worer
//	//Worker() Worker
//	Marshal() ([]byte, error)
//	//Add a publishing event.
//	AddPubEvent(Event)
//	//Get all publishing events..
//	GetPubEvents() []Event
//	//Delete all publishing events.
//	RemoveAllPubEvent()
//	//Add a subscription event.
//	AddSubEvent(Event)
//	//Get all subscription events..
//	GetSubEvents() []Event
//	//Delete all subscription events.
//	RemoveAllSubEvent()
//}
//
//type entity struct {
//	identity     string
//	entityObject interface{}
//	pubEvents    []Event
//	subEvents    []Event
//}
//
//func (e *entity) Identity() string {
//	if e.identity == "" {
//		u, _ := uuid.NewUUID()
//		e.identity = u.String()
//	}
//	return e.identity
//}
//
//func (e *entity) AddPubEvent(event Event) {
//	if reflect.ValueOf(event.EventId()).IsZero() {
//		event.SetEventId(uuid.New().String())
//	}
//
//	e.pubEvents = append(e.pubEvents, event)
//}
//
//// GetPubEvent .
//func (e *entity) GetPubEvents() (result []Event) {
//	return e.pubEvents
//}
//
//// RemoveAllPubEvent .
//func (e *entity) RemoveAllPubEvent() {
//	e.pubEvents = []Event{}
//}
//
//// AddSubEvent.
//func (e *entity) AddSubEvent(event Event) {
//	e.subEvents = append(e.subEvents, event)
//}
//
//// GetSubEvent .
//func (e *entity) GetSubEvents() (result []Event) {
//	return e.subEvents
//}
//
//// RemoveAllSubEvent .
//func (e *entity) RemoveAllSubEvent() {
//	e.subEvents = []Event{}
//}

