package calendar

import (
	"bufio"
	"strings"
	"time"
)

// ICalEvent represents a parsed VEVENT from an iCal feed.
type ICalEvent struct {
	UID         string
	Summary     string
	Description string
	Start       time.Time
	End         time.Time
}

// ParseICal parses an iCalendar (RFC 5545) feed and returns the list of events.
func ParseICal(data string) ([]ICalEvent, error) {
	var events []ICalEvent
	var current *ICalEvent
	inEvent := false

	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		if line == "BEGIN:VEVENT" {
			inEvent = true
			current = &ICalEvent{}
			continue
		}
		if line == "END:VEVENT" {
			if current != nil && !current.Start.IsZero() && !current.End.IsZero() {
				events = append(events, *current)
			}
			inEvent = false
			current = nil
			continue
		}

		if !inEvent || current == nil {
			continue
		}

		key, value := parseICalLine(line)
		switch {
		case key == "UID":
			current.UID = value
		case key == "SUMMARY":
			current.Summary = unescapeICalText(value)
		case key == "DESCRIPTION":
			current.Description = unescapeICalText(value)
		case key == "DTSTART" || strings.HasPrefix(key, "DTSTART;"):
			if t, err := parseICalDateTime(value); err == nil {
				current.Start = t
			}
		case key == "DTEND" || strings.HasPrefix(key, "DTEND;"):
			if t, err := parseICalDateTime(value); err == nil {
				current.End = t
			}
		}
	}

	return events, scanner.Err()
}

func parseICalLine(line string) (string, string) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return line, ""
	}
	return line[:idx], line[idx+1:]
}

func parseICalDateTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	// 20060102T150405Z — UTC
	if strings.HasSuffix(value, "Z") {
		return time.Parse("20060102T150405Z", value)
	}

	// 20060102T150405 — local time (treat as UTC)
	if len(value) == 15 && strings.Contains(value, "T") {
		return time.Parse("20060102T150405", value)
	}

	// 20060102 — all-day event (treat as full day UTC)
	if len(value) == 8 {
		return time.Parse("20060102", value)
	}

	return time.Time{}, &time.ParseError{Message: "unsupported date format: " + value}
}

func unescapeICalText(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\,", ",")
	s = strings.ReplaceAll(s, "\\;", ";")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
