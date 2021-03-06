package main

type WeightListener struct {
	scale         Scale
	events        chan float64
}

func StartWeightListener() chan float64 {
	events := make(chan float64)
	go WeightListener{
		scale:         InitScale(),
		events:        events,
	}.run()
	return events
}

func (w WeightListener) run() {
	defer w.scale.Close()
	// note: read as fast as there is data available
	// (approximately 6 times per second in practice)
	for {
		weight := w.scale.Read()
		currentWeight.Set(weight)
		w.events <- weight
	}
}
