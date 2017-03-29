package main

import "time"
import (
	"github.com/eapache/queue"
)

type StableWeightListener struct {
	in             chan float64
	out            chan float64
	stableDuration time.Duration
	lastValue      float64
	eventCache     *queue.Queue // note: not threadsafe, okay here though
	isChanging     bool
}

type WeightEvent struct {
	timestamp time.Time
	weight    float64
}

func StartStableWeightListener(in chan float64, stableDuration time.Duration) chan float64 {
	out := make(chan float64)
	go StableWeightListener{
		in,
		out,
		stableDuration,
		-999999, // so first stable value is an event
		queue.New(),
		true,
	}.run()
	return out
}

func (s StableWeightListener) run() {
	for event := range s.in {
		e := WeightEvent{time.Now(), event}
		s.eventCache.Add(e)
		s.pruneOld(time.Now())
		if s.isChanging && s.isStable() {
			s.isChanging = false
			// became stable, send event if it is different than last value
			if s.lastValue != event {
				s.lastValue = event
				s.out <- event
			}
		} else if !s.isChanging && !s.isStable() {
			// started changing
			s.isChanging = true
		}
	}
}

func (s StableWeightListener) pruneOld(t time.Time) {
	for i := 0; i < s.eventCache.Length(); i++ {
		event := s.eventCache.Get(i).(WeightEvent)
		if event.timestamp.Before(t.Add(-s.stableDuration)) {
			s.eventCache.Remove()
		} else {
			// queue is sorted, no need to go ahead
			return
		}
	}
}

func (s StableWeightListener) isStable() bool {
	// we'll call Get(0) twice, oh well :) gotta init weight
	weight := s.eventCache.Get(0).(WeightEvent).weight
	for i := 0; i < s.eventCache.Length(); i++ {
		event := s.eventCache.Get(i).(WeightEvent)
		if weight != event.weight {
			return false
		}
	}
	return true
}
