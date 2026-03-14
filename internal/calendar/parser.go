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

	lines := unfoldICalLines(data)

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

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

	return events, nil
}

// unfoldICalLines handles RFC 5545 line folding: lines starting with a space
// or tab are continuations of the previous line.
func unfoldICalLines(data string) []string {
	scanner := bufio.NewScanner(strings.NewReader(data))
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') && len(lines) > 0 {
			// Continuation line: append to previous line (strip leading whitespace char)
			lines[len(lines)-1] += line[1:]
		} else {
			lines = append(lines, line)
		}
	}
	return lines
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
