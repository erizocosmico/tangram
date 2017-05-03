package report

import (
	"github.com/erizocosmico/elmo/source"
	"github.com/erizocosmico/elmo/token"
)

// Reporter is in charge of reporting the diagnostics occurred during any of
// the compilation steps to the user.
type Reporter struct {
	cm      *source.CodeMap
	emitter Emitter
	reports map[string][]Report
}

// NewReporter creates a new reporter.
func NewReporter(cm *source.CodeMap, emitter Emitter) *Reporter {
	return &Reporter{cm, emitter, make(map[string][]Report)}
}

// IsOK returns true if there are no diagnostics yet.
func (r *Reporter) IsOK() bool {
	return len(r.reports) == 0
}

func (r *Reporter) Reports(path string) []Report {
	return r.reports[path]
}

// Emit writes all the reports using the reporter's emitter.
func (r *Reporter) Emit() error {
	for file, reports := range r.reports {
		var ds = make([]*Diagnostic, 0, len(reports))
		for _, report := range reports {
			d, err := r.makeDiagnostic(file, report)
			if err != nil {
				return err
			}

			ds = append(ds, d)
		}

		if err := r.emitter.Emit(file, ds); err != nil {
			return err
		}
	}
	return nil
}

// Report adds a new report occurred at some path.
func (r *Reporter) Report(path string, report Report) {
	r.reports[path] = append(r.reports[path], report)
}

// makeDiagnostic transforms a report into a diagnostic, with the affected
// snippet of code, if there is any.
func (r *Reporter) makeDiagnostic(path string, report Report) (*Diagnostic, error) {
	if report.Pos() == token.NoPos {
		return &Diagnostic{
			Type:    report.Type(),
			Message: report.Message(),
		}, nil
	}

	src := r.cm.Source(path)
	pos, err := src.LinePos(report.Pos())
	if err != nil {
		return nil, err
	}
	region := report.Region()

	var snippet *source.Snippet
	if region != nil {
		snippet, err = src.Region(region.Start, region.End)
		if err != nil {
			return nil, err
		}
	}

	return &Diagnostic{
		Type:    report.Type(),
		Message: report.Message(),
		Pos:     pos,
		Region:  snippet,
	}, nil
}
