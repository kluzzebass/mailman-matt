package main

import (
	ics "github.com/arran4/golang-ical"
	"github.com/google/uuid"
)

type CalendarBuilder struct {
	cfg Config
}

func NewCalendarBuilder(cfg Config) *CalendarBuilder {
	return &CalendarBuilder{
		cfg: cfg,
	}
}

func (c *CalendarBuilder) buildCalendar(sch schedule) *ics.Calendar {
	cal := ics.NewCalendar()
	cal.SetProductId(c.cfg.ProductID)
	cal.SetName(c.cfg.Name)
	cal.SetXWRCalName(c.cfg.Name)
	cal.SetTimezoneId(c.cfg.Timezone)
	cal.SetXWRTimezone(c.cfg.Timezone)

	for _, d := range sch {
		e := cal.AddEvent(uuid.New().String())
		e.SetSequence(0)
		e.SetDtStampTime(d)
		e.SetProperty("DTSTART;VALUE=DATE", d.Format("20060102"))
		e.SetProperty("X-MICROSOFT-CDO-ALLDAYEVENT", "TRUE")
		e.SetProperty("X-MICROSOFT-MSNCALENDAR-ALLDAYEVENT", "TRUE")
		e.SetSummary(c.cfg.Summary)
	}

	return cal
}
