//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package circuit

import (
	"fmt"
	"os"
	"time"

	"github.com/markkurossi/tabulate"
)

type Timing struct {
	Start   time.Time
	Samples []*Sample
}

func NewTiming() *Timing {
	return &Timing{
		Start: time.Now(),
	}
}

func (t *Timing) Sample(label string, cols []string) *Sample {
	start := t.Start
	if len(t.Samples) > 0 {
		start = t.Samples[len(t.Samples)-1].End
	}
	sample := &Sample{
		Label: label,
		Start: start,
		End:   time.Now(),
		Cols:  cols,
	}
	t.Samples = append(t.Samples, sample)
	return sample
}

func (t *Timing) Print(xfer string) {
	if len(t.Samples) == 0 {
		return
	}

	tab := tabulate.New(tabulate.Unicode)
	tab.Header("Op").SetAlign(tabulate.ML)
	tab.Header("Time").SetAlign(tabulate.MR)
	tab.Header("%").SetAlign(tabulate.MR)
	tab.Header("Xfer").SetAlign(tabulate.MR)

	total := t.Samples[len(t.Samples)-1].End.Sub(t.Start)
	for _, sample := range t.Samples {
		row := tab.Row()
		row.Column(sample.Label)

		duration := sample.End.Sub(sample.Start)
		row.Column(fmt.Sprintf("%s", duration.String()))
		row.Column(fmt.Sprintf("%.2f%%",
			float64(duration)/float64(total)*100))

		for _, col := range sample.Cols {
			row.Column(col)
		}

		for _, sub := range sample.Samples {
			row := tab.Row()
			row.Column(sub.Label).SetFormat(tabulate.FmtItalic)

			var d time.Duration
			if sub.Abs > 0 {
				d = sub.Abs
			} else {
				d = sub.End.Sub(sub.Start)
			}
			row.Column(d.String()).SetFormat(tabulate.FmtItalic)

			row.Column(
				fmt.Sprintf("%.2f%%", float64(d)/float64(duration)*100)).
				SetFormat(tabulate.FmtItalic)

		}
	}
	row := tab.Row()
	row.Column("Total").SetFormat(tabulate.FmtBold)
	row.Column(t.Samples[len(t.Samples)-1].End.Sub(t.Start).String()).
		SetFormat(tabulate.FmtBold)
	row.Column("").SetFormat(tabulate.FmtBold)
	row.Column(xfer).SetFormat(tabulate.FmtBold)

	tab.Print(os.Stdout)
}

type Sample struct {
	Label   string
	Start   time.Time
	End     time.Time
	Abs     time.Duration
	Cols    []string
	Samples []*Sample
}

func (s *Sample) SubSample(label string, end time.Time) {
	start := s.Start
	if len(s.Samples) > 0 {
		start = s.Samples[len(s.Samples)-1].End
	}
	s.Samples = append(s.Samples, &Sample{
		Label: label,
		Start: start,
		End:   end,
	})
}

func (s *Sample) AbsSubSample(label string, duration time.Duration) {
	s.Samples = append(s.Samples, &Sample{
		Label: label,
		Abs:   duration,
	})
}
