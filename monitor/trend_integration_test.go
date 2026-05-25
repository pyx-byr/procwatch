package monitor

import (
	"sync"
	"testing"
)

func TestTrendAnalyzer_ConcurrentAnalyze(t *testing.T) {
	ta := NewTrendAnalyzer(0.5)
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8}

	var wg sync.WaitGroup
	results := make([]TrendResult, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = ta.Analyze("proc", "cpu", values)
		}(i)
	}
	wg.Wait()

	for _, r := range results {
		if r.Direction != TrendRising {
			t.Fatalf("expected rising in concurrent run, got %s", r.Direction)
		}
	}
}

func TestTrendAnalyzer_MultipleProcesses(t *testing.T) {
	ta := NewTrendAnalyzer(0.5)

	procs := map[string][]float64{
		"rising":  {1, 2, 3, 4, 5},
		"falling": {5, 4, 3, 2, 1},
		"stable":  {3, 3, 3, 3, 3},
	}
	expected := map[string]TrendDirection{
		"rising":  TrendRising,
		"falling": TrendFalling,
		"stable":  TrendStable,
	}

	var wg sync.WaitGroup
	for name, vals := range procs {
		wg.Add(1)
		go func(n string, v []float64) {
			defer wg.Done()
			res := ta.Analyze(n, "cpu", v)
			if res.Direction != expected[n] {
				t.Errorf("process %s: expected %s, got %s", n, expected[n], res.Direction)
			}
		}(name, vals)
	}
	wg.Wait()
}
