package main

import "log"

type LitterBoxListener struct {
	in         chan float64
	out        chan LitterBoxEvent
	lastWeight float64
	state      LitterBoxState
}

type WeightRange struct {
	Min float64
	Max float64
}

func (w WeightRange) inRange(value float64) bool {
	return value >= w.Min && value <= w.Max
}

func (w WeightRange) aboveRange(value float64) bool {
	return value > w.Max
}

type LitterBoxEvent struct {
	Type      LitterBoxEventType
	State     LitterBoxState
	Weight    float64
	LowLitter bool
}

// state
type LitterBoxState string

const (
	InitState    = "InitState"
	NormalState  = "NormalState"
	WithCatState = "WithCatState"
	EmptyState   = "EmptyState"
)

type LitterBoxEventType string

const (
	InitType        = "InitType"
	CatOnType       = "CatOnType"
	CatOffType      = "CatOffType"
	ScoopedType     = "ScoopedType"
	LitterAddedType = "LitterAddedType"
	EmptiedType     = "EmptiedType"
	RefilledType    = "RefilledType"
)

var (
	catWeight    = WeightRange{9, 12}
	scoopWeight  = WeightRange{-2, 0}
	normalWeight = WeightRange{2, 20}
	lowLitter    = WeightRange{2, 7}
	litterAdded  = WeightRange{0, 20}
	emptyWeight  = WeightRange{0, 2}
)

func StartLitterBoxListener(in chan float64) chan LitterBoxEvent {
	out := make(chan LitterBoxEvent)
	go LitterBoxListener{
		in,
		out,
		0,
		InitState,
	}.run()
	return out
}

func (l LitterBoxListener) run() {
	for weight := range l.in {
		log.Println(weight)
		weightChange := weight - l.lastWeight
		var lType LitterBoxEventType
		var lState LitterBoxState
		var litterWeight float64 = weight
		switch l.state {
		case InitState:
			lType = InitType
			if normalWeight.inRange(weight) {
				lState = NormalState
			} else if normalWeight.aboveRange(weight) {
				lState = WithCatState
			} else if emptyWeight.inRange(weight) {
				lState = EmptyState
			}
		case NormalState:
			// could be cat on, scooped, litter added, refilled
			if catWeight.inRange(weightChange) {
				lType = CatOnType
				lState = WithCatState
				litterWeight = l.lastWeight
			} else if scoopWeight.inRange(weightChange) {
				lType = ScoopedType
				lState = NormalState
				scoops.Inc()
			} else if litterAdded.inRange(weightChange) {
				lType = LitterAddedType
				lState = NormalState
				litterAdd.Inc()
			} else if emptyWeight.inRange(weight) {
				lType = EmptiedType
				lState = EmptyState
				emptied.Inc()
			} else {
				log.Println("Unrecognized event in normal state")
			}
		case WithCatState:
			if catWeight.inRange(-weightChange) {
				lType = CatOffType
				lState = NormalState
				catWent.Inc()
			} else {
				log.Println("Unrecognized event while cat on litter box")
			}
		case EmptyState:
			if normalWeight.inRange(weight) {
				lType = RefilledType
				lState = NormalState
				refilled.Inc()
			} else {
				log.Println("Unrecognized event while empty")
			}
		}
		low := lowLitter.inRange(litterWeight)
		l.state = lState
		l.lastWeight = weight
		l.out <- LitterBoxEvent{lType, lState, weight, low}
	}
}
