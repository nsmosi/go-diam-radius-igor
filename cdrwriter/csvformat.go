package cdrwriter

import (
	"fmt"
	"strings"
	"time"

	"github.com/francistor/igor/diamcodec"
	"github.com/francistor/igor/radiuscodec"
	"github.com/francistor/igor/radiusdict"
)

type CSVWriter struct {
	// The attribute names of the fields to write
	fields []string
	// Separator for the fields
	fieldSeparator string
	// If multiple instances are found, separator to use
	attributeSeparator string
	// To format dates
	attributeDateFormat string
	// Whether to surround strings by quotes
	quoteStrings bool
	// Whether to parse ints as strings
	parseInts bool
}

// Creates a new instance of a Livinstone Writer
func NewCSVWriter(fields []string, fieldSeparator string, attributeSeparator string, attributeDateFormat string, quoteStrings bool, parseInts bool) *CSVWriter {
	lw := CSVWriter{
		fields:              fields,
		fieldSeparator:      fieldSeparator,
		attributeSeparator:  attributeSeparator,
		attributeDateFormat: attributeDateFormat,
		quoteStrings:        quoteStrings,
		parseInts:           parseInts,
	}

	return &lw
}

// Not implemented
func (w *CSVWriter) GetDiameterCDRString(dm *diamcodec.DiameterMessage) string {
	panic("GetDiameterCDRString is not implemented by CSVWriter")
}

// Write CDR as list with separators
// Special field names:
// * %Timestamp% -> Datetime of CDR generation
func (w *CSVWriter) GetRadiusCDRString(rp *radiuscodec.RadiusPacket) string {
	var builder strings.Builder

	// Iterate through the fields in the spec
	for i, field := range w.fields {

		if field == "%Timestamp%" {
			// Write as string
			if w.quoteStrings {
				builder.WriteString("\"")
			}
			builder.WriteString(time.Now().Format(w.attributeDateFormat))
			// Write as string
			if w.quoteStrings {
				builder.WriteString("\"")
			}
		}

		// Get all the attributes for that name
		avps := rp.GetAllAVP(field)

		// Do not write quotes if no attributes found
		if len(avps) > 0 {

			radiusType := avps[0].DictItem.RadiusType
			if (radiusType == radiusdict.Integer || radiusType == radiusdict.Integer64) && !w.parseInts {
				// Write as integer
				for j := range avps {
					builder.WriteString(fmt.Sprintf("%d", avps[j].GetInt()))
					if j < len(avps)-1 {
						builder.WriteString(w.attributeSeparator)
					}
				}
			} else if radiusType == radiusdict.Time {
				// Write as string
				if w.quoteStrings {
					builder.WriteString("\"")
				}
				for j := range avps {
					builder.WriteString(avps[j].GetDate().Format(w.attributeDateFormat))
					if j < len(avps)-1 {
						builder.WriteString(w.attributeSeparator)
					}
				}
				if w.quoteStrings {
					builder.WriteString("\"")
				}
			} else {
				// Write as string
				if w.quoteStrings {
					builder.WriteString("\"")
				}
				for j := range avps {
					builder.WriteString(avps[j].GetTaggedString())
					if j < len(avps)-1 {
						builder.WriteString(w.attributeSeparator)
					}
				}
				if w.quoteStrings {
					builder.WriteString("\"")
				}
			}
		}

		// If not the last, write separator
		if i < len(w.fields)-1 {
			builder.WriteString(w.fieldSeparator)
		}
	}

	builder.WriteString("\n")

	return builder.String()
}
