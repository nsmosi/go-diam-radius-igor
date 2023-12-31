package cdrwriter

import (
	"fmt"
	"strings"
	"time"

	"github.com/francistor/igor/core"

	"golang.org/x/exp/slices"
)

type LivingstoneFormat struct {
	positiveFilter      []string
	negativeFilter      []string
	headDateFormat      string
	attributeDateFormat string
}

// Creates a new instance of a Livinstone Writer
func NewLivingstoneFormat(positiveFilter []string, negativeFilter []string, headDateFormat string, attributeDateFormat string) *LivingstoneFormat {
	lw := LivingstoneFormat{
		positiveFilter:      positiveFilter,
		negativeFilter:      negativeFilter,
		headDateFormat:      headDateFormat,
		attributeDateFormat: attributeDateFormat,
	}

	return &lw
}

// Not implemented
func (w *LivingstoneFormat) GetDiameterCDRString(dm *core.DiameterMessage) string {
	panic("GetDiameterCDRString is not implemented by CSVWriter")
}

func (w *LivingstoneFormat) GetRadiusCDRString(rp *core.RadiusPacket) string {
	var builder strings.Builder

	// Write header
	builder.WriteString(time.Now().Format(w.headDateFormat))
	builder.WriteString("\n")

	// Write AVPs
	for i := range rp.AVPs {

		// Apply filters
		if w.positiveFilter != nil && !slices.Contains(w.positiveFilter, rp.AVPs[i].Name) {
			continue
		} else if w.negativeFilter != nil && slices.Contains(w.negativeFilter, rp.AVPs[i].Name) {
			continue
		}

		builder.WriteString(rp.AVPs[i].Name)

		switch rp.AVPs[i].DictItem.RadiusType {

		case core.RadiusTypeNone, core.RadiusTypeOctets, core.RadiusTypeString, core.RadiusTypeInterfaceId, core.RadiusTypeAddress, core.RadiusTypeIPv6Address, core.RadiusTypeIPv6Prefix:
			// Write as a string
			builder.WriteString("=\"")
			builder.WriteString(rp.AVPs[i].GetTaggedString())
			builder.WriteString("\"\n")

		case core.RadiusTypeInteger:
			// Try dictionary, if not found use integer value
			var intValue, _ = rp.AVPs[i].Value.(int64)
			if stringValue, ok := rp.AVPs[i].DictItem.EnumCodes[int(intValue)]; ok {
				builder.WriteString("=\"")
				builder.WriteString(stringValue)
				builder.WriteString("\"\n")
			} else {
				builder.WriteString("=")
				builder.WriteString(fmt.Sprintf("%d", intValue))
				builder.WriteString("\n")
			}
		case core.RadiusTypeTime:
			// Format date
			builder.WriteString("=\"")
			builder.WriteString(rp.AVPs[i].GetDate().Format(w.attributeDateFormat))
			builder.WriteString("\"\n")
		}

	}

	builder.WriteString("\n")

	return builder.String()
}

/*
https://pkg.go.dev/time#pkg-constants

ANSIC       = "Mon Jan _2 15:04:05 2006"s
UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
RFC822      = "02 Jan 06 15:04 MST"
RFC822Z     = "02 Jan 06 15:04 -0700"
RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700"
RFC3339     = "2006-01-02T15:04:05Z07:00"
RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
*/
