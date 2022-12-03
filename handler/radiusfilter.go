package handler

import (
	"encoding/json"
	"fmt"

	"github.com/francistor/igor/config"
	"github.com/francistor/igor/radiuscodec"
)

/////////////////////////////////////////////////////////////////////////////
// Radius Packet Attribute Filter
/////////////////////////////////////////////////////////////////////////////

// Entry in AVPFilter file
type AVPFilter struct {
	Allow  []string    // List of attributes to allow
	Remove []string    // List of attributes to remove. Makes sense either Allow or Remove, but not both
	Force  [][2]string // List of attributes to set a specific value. Contents of the list are 2 element arrays (attribute, value)
}

// Contents of the AVPFilters file. Set of AVPFilters by key
type AVPFilters map[string]*AVPFilter

// Creates a copy of the radius packet with the attributes filtered as specified in the filter for the passed key
func (fs AVPFilters) FilteredPacket(packet *radiuscodec.RadiusPacket, key string) (*radiuscodec.RadiusPacket, error) {
	if filter, ok := fs[key]; !ok {
		return &radiuscodec.RadiusPacket{}, fmt.Errorf("%s filter not found", key)
	} else {
		return filter.FilteredPacket(packet), nil
	}
}

// Copy the radius packet with the attributes modified as defined in the specified filter
func (f *AVPFilter) FilteredPacket(packet *radiuscodec.RadiusPacket) *radiuscodec.RadiusPacket {
	var rp *radiuscodec.RadiusPacket
	if len(f.Allow) > 0 {
		rp = packet.Copy(f.Allow, nil)
	} else if len(f.Remove) > 0 {
		rp = packet.Copy(nil, f.Remove)
	} else {
		rp = packet.Copy(nil, nil)
	}

	for _, forceSpec := range f.Force {
		if len(forceSpec) == 2 {
			rp.DeleteAllAVP(forceSpec[0])
			rp.Add(forceSpec[0], forceSpec[1])
		}
	}

	return rp
}

// Returns an object representing the configured AVPFilters
func NewAVPFilters(configObjectName string, ci *config.PolicyConfigurationManager) (AVPFilters, error) {

	filters := make(AVPFilters)

	// If we pass nil as last parameter, use the default
	var myCi *config.PolicyConfigurationManager
	if ci == nil {
		myCi = config.GetPolicyConfig()
	} else {
		myCi = ci
	}

	// Read the configuration object
	fs, err := myCi.CM.GetBytesConfigObject(configObjectName)
	if err != nil {
		return filters, err
	}

	if err = json.Unmarshal(fs, &filters); err != nil {
		return filters, fmt.Errorf(err.Error())
	}

	return filters, nil
}
