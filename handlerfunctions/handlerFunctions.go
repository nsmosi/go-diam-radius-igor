package handlerfunctions

import (
	"encoding/json"
	"fmt"
	"igor/config"
	"igor/diamcodec"
	"igor/instrumentation"
	"igor/radiuscodec"

	"go.uber.org/zap/zapcore"
)

// The most basic handler ever. Returns an empty response to the received message
func EmptyDiameterHandler(request *diamcodec.DiameterMessage) (*diamcodec.DiameterMessage, error) {
	logLines := make(instrumentation.LogLines, 0)

	defer func(lines []instrumentation.LogLine) {
		logLines.WriteWLog()
	}(logLines)

	logLines.WLogEntry(zapcore.InfoLevel, "%s", "Starting EmptyDiameterHandler")
	logLines.WLogEntry(zapcore.InfoLevel, "%s %s", "request", request)

	response := diamcodec.NewDiameterAnswer(request)
	response.Add("Result-Code", diamcodec.DIAMETER_SUCCESS)

	logLines.WLogEntry(zapcore.InfoLevel, "%s %s", "response", request)

	return response, nil
}

// The most basic handler ever. Returns an empty response to the received message
func EmptyRadiusHandler(request *radiuscodec.RadiusPacket) (*radiuscodec.RadiusPacket, error) {
	resp := radiuscodec.NewRadiusResponse(request, true)

	return resp, nil
}

func TestRadiusAttributesHandler(request *radiuscodec.RadiusPacket) (*radiuscodec.RadiusPacket, error) {

	logger := config.GetLogger()

	// Print the password
	pwd := request.GetPasswordStringAVP("User-Password")

	logger.Infof("Password: <%s>", pwd)

	// Print all received attributes
	for _, avp := range request.AVPs {
		logger.Infof("%s -> %s", avp.Name, avp.GetTaggedString())
	}

	// Reply with one attribute of each type
	// The Igor-SaltedOctetsAttribute contains the length as the first byte, since
	// in Nokia AAA this VSA is "salted-password" type
	jAVPs := `
				[
					{"Igor-OctetsAttribute": "0102030405060708090a0b"},
					{"Igor-StringAttribute": "stringvalue"},
					{"Igor-IntegerAttribute": "Zero"},
					{"Igor-IntegerAttribute": "1"},
					{"Igor-IntegerAttribute": 1},
					{"Igor-AddressAttribute": "127.0.0.1"},
					{"Igor-TimeAttribute": "1966-11-26T03:34:08 UTC"},
					{"Igor-IPv6AddressAttribute": "bebe:cafe::0"},
					{"Igor-IPv6PrefixAttribute": "bebe:cafe:cccc::0/64"},
					{"Igor-InterfaceIdAttribute": "00aabbccddeeff11"},
					{"Igor-TaggedStringAttribute": "mystring:1"},
					{"Igor-Integer64Attribute": 999999999999},
					{"Igor-SaltedOctetsAttribute": "0F313233343536373839616263646566"},
					{"Igor-TaggedSaltedOctetsAttribute": "0F313233343536373839616263646566:1"},
					{"User-Name":"MyUserName"}
				]
				`

	resp := radiuscodec.NewRadiusResponse(request, true)

	var responseAVPs []radiuscodec.RadiusAVP
	err := json.Unmarshal([]byte(jAVPs), &responseAVPs)
	if err != nil {
		logger.Errorf("%s", err)
		return nil, fmt.Errorf(err.Error())
	}

	for _, avp := range responseAVPs {
		resp.AddAVP(&avp)
	}

	return resp, nil
}
