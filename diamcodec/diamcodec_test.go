package diamcodec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"igor/config"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Initialization
var bootstrapFile = "resources/searchRules.json"
var instanceName = "testClient"

// Initializer of the test suite.
func TestMain(m *testing.M) {
	config.InitPolicyConfigInstance(bootstrapFile, instanceName, true)

	// Execute the tests and exit
	os.Exit(m.Run())
}

func TestAVPNotFound(t *testing.T) {
	var _, err = NewAVP("Unknown AVP", []byte("hello, world!"))
	if err == nil {
		t.Errorf("Unknown AVP was created")
	}
}

func TestOctetsAVP(t *testing.T) {

	var password = "'my-password!"

	// Create avp
	avp, err := NewAVP("User-Password", []byte(password))
	if err != nil {
		t.Errorf("error creating Octets AVP: %v", err)
		return
	}
	if avp.GetString() != fmt.Sprintf("%x", password) {
		t.Errorf("Octets AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != fmt.Sprintf("%x", password) {
		t.Errorf("Octets AVP not properly encoded after unmarshalling. Got %s", rebuiltAVP.GetString())
	}
	if !reflect.DeepEqual(rebuiltAVP.GetOctets(), []byte(password)) {
		t.Errorf("Octets AVP not properly encoded after unmarshalling. Got %v instead of %v", rebuiltAVP.GetOctets(), []byte(password))
	}
}

func TestUTF8StringAVP(t *testing.T) {

	var theString = "%Hola España. 'Quiero €"

	// Create avp
	avp, err := NewAVP("User-Name", theString)
	if err != nil {
		t.Errorf("error creating UTFString AVP %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("UTF8String AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != theString {
		t.Errorf("UTF8String AVP not properly encoded after unmarshalling. Got %s", rebuiltAVP.GetString())
	}
}

func TestInt32AVP(t *testing.T) {

	var theInt int32 = -65535*16384 - 1000 // 2^31 - 1000

	// Create avp
	avp, err := NewAVP("Igor-myInteger32", theInt)
	if err != nil {
		t.Errorf("error creating Int32 AVP %v", err)
		return
	}
	if avp.GetInt() != int64(theInt) {
		t.Errorf("Int32 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != fmt.Sprint(theInt) {
		t.Errorf("Integer32 AVP not properly encoded after unmarshalling (string value). Got %s", rebuiltAVP.GetString())
	}
	if rebuiltAVP.GetInt() != int64(theInt) {
		t.Errorf("Integer32 AVP not properly encoded after unmarshalling (long value). Got %d", rebuiltAVP.GetInt())
	}
	if rebuiltAVP.GetInt() >= 0 {
		t.Errorf("Integer32 should be negative. Got %d", rebuiltAVP.GetInt())
	}
}

func TestInt64AVP(t *testing.T) {

	var theInt int64 = -65535*65535*65534*16384 - 999 // - 2 ^ 62 - 999
	// Create avp
	avp, err := NewAVP("Igor-myInteger64", theInt)
	if err != nil {
		t.Errorf("error creating Int64 AVP %v", err)
		return
	}
	if avp.GetInt() != int64(theInt) {
		t.Errorf("Int64 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != fmt.Sprint(theInt) {
		t.Errorf("Integer64 AVP not properly encoded after unmarshalling (string value). Got %s", rebuiltAVP.GetString())
	}
	if rebuiltAVP.GetInt() != int64(theInt) {
		t.Errorf("Integer64 AVP not properly encoded after unmarshalling (long value). Got %d", rebuiltAVP.GetInt())
	}
	if rebuiltAVP.GetInt() >= 0 {

		t.Errorf("Integer64 should be negative. Got %d", rebuiltAVP.GetInt())
	}
}

func TestUnsignedInt32AVP(t *testing.T) {

	var theInt uint32 = 65535 * 40001

	// Create avp
	avp, err := NewAVP("Igor-myUnsigned32", int64(theInt))
	if err != nil {
		t.Errorf("error creating UInt32 AVP %v", err)
		return
	}
	if avp.GetInt() != int64(theInt) {
		t.Errorf("UInt32 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != fmt.Sprint(theInt) {
		t.Errorf("UnsignedInteger32 AVP not properly encoded after unmarshalling (string value). Got %s", rebuiltAVP.GetString())
	}
	if rebuiltAVP.GetInt() != int64(theInt) {
		t.Errorf("UnsignedInteger32 AVP not properly encoded after unmarshalling (long value). Got %d", rebuiltAVP.GetInt())
	}
	if rebuiltAVP.GetInt() < 0 {
		t.Errorf("Unsigned Integer32 should be positive. Got %d", rebuiltAVP.GetInt())
	}
}

func TestUnsignedInt64AVP(t *testing.T) {

	// Due to a limitaton of the implementation, it is inernally stored as a signed int64
	var theInt int64 = 65535 * 65535 * 65535 * 16001

	// Create avp
	avp, err := NewAVP("Igor-myUnsigned64", theInt)
	if err != nil {
		t.Errorf("error creating UInt64 AVP %v", err)
		return
	}
	if avp.GetInt() != int64(theInt) {
		t.Errorf("Unsigned Int64 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	rebuiltAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if rebuiltAVP.GetString() != fmt.Sprint(theInt) {
		t.Errorf("Unsigned Integer64 AVP not properly encoded after unmarshalling (string value). Got %s", rebuiltAVP.GetString())
	}
	if rebuiltAVP.GetInt() != int64(theInt) {
		t.Errorf("Unsigned Integer64 AVP not properly encoded after unmarshalling (long value). Got %d", rebuiltAVP.GetInt())
	}
	if rebuiltAVP.GetInt() < 0 {
		t.Errorf("Unsigned Integer64 should be positive. Got %d", rebuiltAVP.GetInt())
	}
}

func TestFloat32AVP(t *testing.T) {

	var theFloat float32 = 6.03e23

	// Create avp
	avp, err := NewAVP("Igor-myFloat32", theFloat)
	if err != nil {
		t.Errorf("error creating Float32 AVP %v", err)
		return
	}
	if avp.GetFloat() != float64(theFloat) {
		t.Errorf("Float32 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != fmt.Sprintf("%f", theFloat) {
		t.Errorf("Float32 AVP not properly encoded after unmarshalling (string value). Got %s", recoveredAVP.GetString())
	}
	if recoveredAVP.GetFloat() != float64(theFloat) {
		t.Errorf("Float32 AVP not properly encoded after unmarshalling (long value). Got %f", recoveredAVP.GetFloat())
	}
}

func TestFloat64AVP(t *testing.T) {

	var theFloat float64 = 6.03e23

	// Create avp
	avp, err := NewAVP("Igor-myFloat64", float64(theFloat))
	if err != nil {
		t.Errorf("error creating Float64 AVP %v", err)
		return
	}
	if avp.GetFloat() != float64(theFloat) {
		t.Errorf("Float64 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != fmt.Sprintf("%f", theFloat) {
		t.Errorf("Float64 AVP not properly encoded after unmarshalling (string value). Got %s", recoveredAVP.GetString())
	}
	if recoveredAVP.GetFloat() != float64(theFloat) {
		t.Errorf("Float64 AVP not properly encoded after unmarshalling (long value). Got %f", recoveredAVP.GetFloat())
	}
}

func TestAddressAVP(t *testing.T) {

	var ipv4Address = "1.2.3.4"
	var ipv6Address = "bebe:cafe::0"

	// Using strings as values

	// IPv4
	// Create avp
	avp, err := NewAVP("Igor-myAddress", ipv4Address)
	if err != nil {
		t.Errorf("error creating IPv4 Address AVP: %v", err)
		return
	}
	if avp.GetString() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP not properly encoded after unmarshalling (string value). Got %s %s", recoveredAVP.GetString(), net.ParseIP(ipv4Address).String())
	}

	// IPv6
	// Create avp
	avp, err = NewAVP("Igor-myAddress", ipv6Address)
	if err != nil {
		t.Errorf("error creating IPv6 Address AVP: %v", err)
	}
	if avp.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ = avp.MarshalBinary()
	recoveredAVP, _, _ = DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP not properly encoded after unmarshalling (string value). Got %s %s", recoveredAVP.GetString(), net.ParseIP(ipv6Address).String())
	}

	// Using IP addresses as value
	avp, _ = NewAVP("Igor-myAddress", net.ParseIP(ipv4Address))
	if avp.GetString() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP does not match value (created as ipaddr) %s %s", avp.GetString(), net.ParseIP(ipv4Address).String())
	}

	avp, _ = NewAVP("Igor-myAddress", net.ParseIP(ipv6Address))
	if avp.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP does not match value (created as ipaddr) %s %s", avp.GetString(), net.ParseIP(ipv6Address).String())
	}
}

func TestIPv4Address(t *testing.T) {

	var ipv4Address = "1.2.3.4"

	// Create avp from string
	avp, err := NewAVP("Igor-myIPv4Address", ipv4Address)
	if err != nil {
		t.Errorf("error creating IPv4 Address AVP %v", err)
		return
	}
	if avp.GetString() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP not properly encoded after unmarshalling (string value). Got %s", recoveredAVP.GetString())
	}

	// Create avp from address
	avp, _ = NewAVP("Igor-myIPv4Address", net.ParseIP(ipv4Address))
	if avp.GetIPAddress().String() != net.ParseIP(ipv4Address).String() {
		t.Errorf("IPv4 AVP does not match value (created as ipaddr) %s", avp.GetIPAddress())
	}
}

func TestIPv6Address(t *testing.T) {
	var ipv6Address = "bebe:cafe::0"

	// Create avp from string
	avp, err := NewAVP("Igor-myIPv6Address", ipv6Address)
	if err != nil {
		t.Errorf("error creating IPv6 Address AVP %v", err)
		return
	}
	if avp.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP not properly encoded after unmarshalling (string value). Got %s", recoveredAVP.GetString())
	}

	// Create avp from IP address
	avp, _ = NewAVP("Igor-myIPv6Address", net.ParseIP(ipv6Address))
	if avp.GetString() != net.ParseIP(ipv6Address).String() {
		t.Errorf("IPv6 AVP does not match value (created as ipaddr) %s", avp.GetString())
	}
}

func TestTimeAVP(t *testing.T) {
	var theTime, _ = time.Parse("02/01/2006 15:04:05 UTC", "26/11/1966 03:21:54 UTC")
	var theStringTime = "1966-11-26T03:21:54 UTC"

	// Create avp from string
	avp, err := NewAVP("Igor-myTime", theStringTime)
	if err != nil {
		t.Errorf("error creating Time Address AVP %v", err)
		return
	}
	if avp.GetDate() != theTime {
		t.Errorf("Time AVP does not match value")
	}
}

func TestDiamIdentAVP(t *testing.T) {

	var theString = "domain.name"

	// Create avp
	avp, err := NewAVP("Igor-myDiameterIdentity", theString)
	if err != nil {
		t.Errorf("error creating Diameter Identity AVP %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("Diamident AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != theString {
		t.Errorf("Diameter Identity AVP not properly encoded after unmarshalling. Got %s", recoveredAVP.GetString())
	}
}

func TestDiamURIAVP(t *testing.T) {

	var theString = "domain.name"

	// Create avp
	avp, err := NewAVP("Igor-myDiameterURI", theString)
	if err != nil {
		t.Errorf("error creating Diameter URI AVP %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("Diamident AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != theString {
		t.Errorf("Diameter URI AVP not properly encoded after unmarshalling. Got %s", recoveredAVP.GetString())
	}
}

func TestIPFilterRuleIAVP(t *testing.T) {

	var theString = "deny 1.2.3.4"

	// Create avp
	avp, err := NewAVP("Igor-myIPFilterRule", theString)
	if err != nil {
		t.Errorf("error creating IP Filter Rule AVP %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("IP Filter Rule AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != theString {
		t.Errorf("IP Filter Rule AVP not properly encoded after unmarshalling. Got %s", recoveredAVP.GetString())
	}
}

func TestIPv6PrefixAVP(t *testing.T) {

	var thePrefix = "bebe:cafe::/16"

	// Create avp
	avp, err := NewAVP("Igor-myIPv6Prefix", thePrefix)
	if err != nil {
		t.Errorf("error creating IPv6 prefix AVP %v", err)
		return
	}
	if avp.GetString() != thePrefix {
		t.Errorf("IPv6 Prefix AVP does not match value")
	}

	// Serialize and unserialize
	binaryAVP, _ := avp.MarshalBinary()
	recoveredAVP, _, _ := DiameterAVPFromBytes(binaryAVP)
	if recoveredAVP.GetString() != thePrefix {
		t.Errorf("IPv6 Prefix AVP not properly encoded after unmarshalling. Got %s", recoveredAVP.GetString())
	}
}

func TestEnumeratedAVP(t *testing.T) {

	var theString = "zero"
	var theNumber int64 = 0

	avp, err := NewAVP("Igor-myEnumerated", "zero")
	if err != nil {
		t.Errorf("error creating Enumerated AVP: %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("Enumerated AVP does not match string value")
	}
	if avp.GetInt() != theNumber {
		t.Errorf("Enumerated AVP does not match number value")
	}

	avp, err = NewAVP("Igor-myEnumerated", theNumber)
	if err != nil {
		t.Errorf("error creating Enumerated AVP: %v", err)
		return
	}
	if avp.GetString() != theString {
		t.Errorf("Enumerated AVP does not match string value")
	}
	if avp.GetInt() != theNumber {
		t.Errorf("Enumerated AVP does not match number value")
	}
}

func TestGroupedAVP(t *testing.T) {

	var theInt int64 = 99
	var theString = "theString"

	// Create grouped AVP
	avpl0, _ := NewAVP("Igor-myGroupedInGrouped", nil)
	avpl1, _ := NewAVP("Igor-myGrouped", nil)

	avpInt, _ := NewAVP("Igor-myInteger32", theInt)
	avpString, _ := NewAVP("Igor-myString", theString)

	avpl1.AddAVP(*avpInt).AddAVP(*avpString)
	avpl0.AddAVP(*avpl1)

	// Serialize and unserialize
	binaryAVP, _ := avpl0.MarshalBinary()
	recoveredAVPl0, _, _ := DiameterAVPFromBytes(binaryAVP)

	// Navigate to the values
	recoveredAVPl1 := recoveredAVPl0.GetAllAVP("Igor-myGrouped")[0]
	newInt, _ := recoveredAVPl1.GetAVP("Igor-myInteger32")
	if newInt.GetInt() != theInt {
		t.Error("Integer value does not match or not found in Group")
	}
	newString, _ := recoveredAVPl1.GetAVP("Igor-myString")
	if newString.GetString() != theString {
		t.Error("String value does not match or not found in Group")
	}

	// Non existing AVP
	_, err := recoveredAVPl1.GetAVP("non-existing")
	if err == nil {
		t.Error("No error when trying to find a non existing AVP")
	}

	// Printed avp
	var targetString = "{Igor-myGrouped={Igor-myInteger32=99,Igor-myString=theString}}"
	stringRepresentation := recoveredAVPl0.GetString()
	if stringRepresentation != targetString {
		t.Errorf("Grouped string representation does not match %s", stringRepresentation)
	}
}

func TestSerializationError(t *testing.T) {

	// Generate an AVP
	avp, err := NewAVP("Igor-myOctetString", "0A0B0C0c765654")
	theBytes, _ := avp.MarshalBinary()

	if err != nil {
		t.Errorf("error creating octectstring from string: %s", err)
		return
	}

	// Change the vendorId to something not existing in the dict
	var theBytesUnknown []byte
	theBytesUnknown = append(theBytesUnknown, theBytes...)
	copy(theBytesUnknown[8:12], []byte{11, 12, 13, 14})

	// Simulate we read an AVP not in the dictionary
	// It should create an AVP with name UNKNOWN
	newavp, _, _ := DiameterAVPFromBytes(theBytesUnknown)
	if newavp.VendorId != 11*256*256*256+12*256*256+13*256+14 {
		t.Errorf("unknown vendor Id was not unmarshalled")
	}
	if newavp.DictItem.Name != "UNKNOWN" {
		t.Errorf("unknown AVP not named UNKNOWN")
	}

	// We should be able to serialize the unknown AVP
	// The vendorId should be the same
	otherBytes, marshalError := newavp.MarshalBinary()
	if marshalError != nil {
		t.Errorf("error serializing unknown avp: %s", marshalError)
	}
	if !reflect.DeepEqual([]byte{11, 12, 13, 14}, otherBytes[8:12]) {
		t.Errorf("error serializing unknown avp. Vendor Id does not match: %s", marshalError)
	}

	// Force unmarshalling error. Size is some big number

	copy(theBytesUnknown[5:8], []byte{100, 100, 100})
	_, _, e := DiameterAVPFromBytes(theBytesUnknown)
	if e == nil {
		t.Error("bad bytes should have reported error")
	}

}

func TestJSONAVP(t *testing.T) {

	var javp = `{
		"Igor-myTestAllGrouped": [
			{"Igor-myOctetString": "0102030405060708090a0b"},
			{"Igor-myInteger32": -99},
			{"Igor-myInteger64": -99},
			{"Igor-myUnsigned32": 99},
			{"Igor-myUnsigned64": 99},
			{"Igor-myFloat32": 99.9},
			{"Igor-myFloat64": 99.9},
			{"Igor-myAddress": "1.2.3.4"},
			{"Igor-myTime": "1966-11-26T03:34:08 UTC"},
			{"Igor-myString": "Hello, world!"},
			{"Igor-myDiameterIdentity": "Diameter@identity"},
			{"Igor-myDiameterURI": "Diameter@URI"},
			{"Igor-myIPFilterRule": "allow all"},
			{"Igor-myIPv4Address": "4.5.6.7"},
			{"Igor-myIPv6Address": "bebe:cafe::0"},
			{"Igor-myIPv6Prefix": "bebe:cafe::0/128"},
			{"Igor-myEnumerated": "two"}
		]
	}`

	// Read JSON to AVP
	var avp DiameterAVP
	err := json.Unmarshal([]byte(javp), &avp)
	if err != nil {
		t.Errorf("unmarshal error for avp: %s", err)
	}
	// Check the contents of the unmarshalled avp
	if avp.Name != "Igor-myTestAllGrouped" {
		t.Errorf("unmarshalled avp has the wrong name: %s", avp.Name)
	}
	if v, _ := avp.GetAVP("Igor-myEnumerated"); v.GetInt() != 2 {
		t.Errorf("unmarshalled avp has the wrong name: %s", avp.Name)
	}
	v, _ := avp.GetAVP("Igor-myTime")
	vv, _ := time.Parse(timeFormatString, "1966-11-26T03:34:08 UTC")
	if v.GetDate() != vv {
		t.Errorf("unmarshalled avp has the wrong date value: %s", v.String())
	}

	// Marshal again
	jNewAVP, _ := json.Marshal(&avp)
	if !strings.Contains(string(jNewAVP), "bebe:cafe::0/128") {
		t.Errorf("part of the expected JSON content was not found")
	}

	/*
		var jBytes bytes.Buffer
		if err := json.Indent(&jBytes, []byte(jRecovered), "", "    "); err != nil {
			t.Errorf("prettyprint error %s", err)
		}

		fmt.Println(jBytes.String())
		fmt.Println(avp.String())
	*/
}

/////////////////////////////////////////////////////////////////////////////////////

func TestDiameterMessage(t *testing.T) {

	var ci = config.GetPolicyConfig()

	diameterMessage, err := NewDiameterRequest("TestApplication", "TestRequest")
	diameterMessage.AddOriginAVPs(ci)
	if err != nil {
		t.Errorf("could not create diameter request for application TestAppliciaton and command TestRequest")
		return
	}
	sessionIdAVP, _ := NewAVP("Session-Id", "my-session-id")
	originHostAVP, _ := NewAVP("Origin-Host", "server.igorserver")
	originRealmAVP, _ := NewAVP("Origin-Realm", "igorserver")
	destinationHostAVP, _ := NewAVP("Destination-Host", "server.igorserver")
	destinationRealmAVP, _ := NewAVP("Destination-Realm", "igorserver")
	groupedInGroupedAVP, _ := NewAVP("Igor-myGroupedInGrouped", nil)
	groupedAVP, _ := NewAVP("Igor-myGrouped", nil)
	intAVP, _ := NewAVP("Igor-myInteger32", 1)
	stringAVP, _ := NewAVP("Igor-myString", "hello")
	groupedAVP.AddAVP(*intAVP)
	groupedAVP.AddAVP(*stringAVP)
	groupedInGroupedAVP.AddAVP(*groupedAVP)
	groupedInGroupedAVP.AddAVP(*intAVP)
	groupedInGroupedAVP.AddAVP(*stringAVP)

	diameterMessage.AddAVP(sessionIdAVP)
	diameterMessage.AddAVP(originHostAVP)
	diameterMessage.AddAVP(originRealmAVP)
	diameterMessage.AddAVP(destinationHostAVP)
	diameterMessage.AddAVP(destinationRealmAVP)
	diameterMessage.AddAVP(groupedInGroupedAVP)

	diameterMessage.Add("Igor-myUnsigned32", 8)
	diameterMessage.Add("Igor-myUnsigned32", 9)

	// Serialize
	theBytes, err := diameterMessage.MarshalBinary()
	if err != nil {
		t.Errorf("could not serialize diameter message %s", err)
		return
	}

	// Unserialize
	recoveredMessage, _, err := DiameterMessageFromBytes(theBytes)
	if err != nil {
		t.Errorf("could not unserialize diameter message %s", err)
		return
	}

	// Get and check the values of simple AVP
	unsignedAVPs := recoveredMessage.GetAllAVP("Igor-myUnsigned32")
	if len(unsignedAVPs) != 2 {
		t.Errorf("did not get two unsigned32 avps in Diameter message")
	}
	for _, avp := range unsignedAVPs {
		value := avp.GetInt()
		if value != 8 && value != 9 {
			t.Errorf("incorrect value")
		}
	}

	// Delete the avp
	recoveredMessage.DeleteAllAVP("Igor-myUnsigned32")
	unsignedAVPs = recoveredMessage.GetAllAVP("Igor-myUnsigned32")
	if len(unsignedAVPs) != 0 {
		t.Errorf("avp still there after being deleted")
	}

	// Get and check the value of a grouped AVP
	gig, err := recoveredMessage.GetAVP("Igor-myGroupedInGrouped")
	if err != nil {
		t.Errorf("could not retrieve groupedingrouped avp: %s", err)
		return
	}
	g, err := gig.GetAVP("Igor-myGrouped")
	if err != nil {
		t.Errorf("could not retrieve grouped avp: %s", err)
		return
	}
	s, err := g.GetAVP("Igor-myString")
	if err != nil {
		t.Errorf("could not retrieve string avp: %s", err)
		return
	}
	if s.GetString() != "hello" {
		t.Errorf("got incorrect value for string avp: %s instead of <hello>", err)
	}

	// Generate reply message
	replyMessage := NewDiameterAnswer(&recoveredMessage)
	replyMessage.AddOriginAVPs(ci)
	if replyMessage.IsRequest {
		t.Errorf("reply message is a request")
	}

	// TODO:
	// Cuando se hace return de un item de un slice ¿Es una copia?
	// Cuando se añade un AVP ¿es una copia o se puede modificar el orgiginal?
}

func TestDiameterMessageAllAttributeTypes(t *testing.T) {

	jDiameterMessage := `
	{
		"IsRequest": true,
		"IsProxyable": false,
		"IsError": false,
		"IsRetransmission": false,
		"CommandCode": 2000,
		"ApplicationId": 1000,
		"avps":[
			{
			  "Igor-myTestAllGrouped": [
  				{"Igor-myOctetString": "0102030405060708090a0b"},
  				{"Igor-myInteger32": -99},
  				{"Igor-myInteger64": -99},
  				{"Igor-myUnsigned32": 99},
  				{"Igor-myUnsigned64": 99},
  				{"Igor-myFloat32": 99.9},
  				{"Igor-myFloat64": 99.9},
  				{"Igor-myAddress": "1.2.3.4"},
  				{"Igor-myTime": "1966-11-26T03:34:08 UTC"},
  				{"Igor-myString": "Hello, world!"},
  				{"Igor-myDiameterIdentity": "Diameter@identity"},
  				{"Igor-myDiameterURI": "Diameter@URI"},
  				{"Igor-myIPFilterRule": "allow all"},
  				{"Igor-myIPv4Address": "4.5.6.7"},
  				{"Igor-myIPv6Address": "bebe:cafe::0"},
  				{"Igor-myIPv6Prefix": "bebe:cafe::0/128"},
  				{"Igor-myEnumerated": "two"}
			  ]
			}
		]
	}
	`

	// Read JSON to DiameterMessage
	var diameterMessage DiameterMessage
	err := json.Unmarshal([]byte(jDiameterMessage), &diameterMessage)
	if err != nil {
		t.Errorf("unmarshal error for diameter message: %s", err)
	}
	diameterMessage.Tidy()

	// Write message to buffer
	messageBytes, error := diameterMessage.MarshalBinary()
	if error != nil {
		t.Fatal("Marshal error")
	}

	// Recover message from buffer
	recoveredMessage := DiameterMessage{}
	_, err = recoveredMessage.ReadFrom(bytes.NewReader(messageBytes))
	if err != nil {
		t.Fatalf("Error recovering DiameterMessage from bytes: %s", err)
	}

	if recoveredMessage.GetStringAVP("Igor-myTestAllGrouped.Igor-myAddress") != "1.2.3.4" {
		t.Errorf("Error recovering IP address. Got <%s> instead of 1.2.3.4", recoveredMessage.GetStringAVP("Igor-myTestAllGrouped.Igor-myAddress"))
	}
	if recoveredMessage.GetStringAVP("Igor-myTestAllGrouped.Igor-myEnumerated") != "two" {
		t.Errorf("Error recovering Enumerated. Got <%s> instead of <two>", recoveredMessage.GetStringAVP("Igor-myTestAllGrouped.Igor-myEnumerated"))
	}
	targetTime, _ := time.Parse("2006-01-02T15:04:05 UTC", "1966-11-26T03:34:08 UTC")
	if recoveredMessage.GetDateAVP("Igor-myTestAllGrouped.Igor-myTime") != targetTime {
		t.Errorf("Error recovering date. Got <%v> instead of <1966-11-26T03:34:08 UTC>", recoveredMessage.GetDateAVP("Igor-myTestAllGrouped.Igor-myTime"))
	}
	if recoveredMessage.GetIntAVP("Igor-myTestAllGrouped.Igor-myInteger32") != -99 {
		t.Errorf("Error recovering int. Got <%d> instead of -99", recoveredMessage.GetIntAVP("Igor-myTestAllGrouped.Igor-myInteger32"))
	}
	targetIPAddress := net.ParseIP("4.5.6.7")
	if !recoveredMessage.GetIPAddressAVP("Igor-myTestAllGrouped.Igor-myIPv4Address").Equal(targetIPAddress) {
		t.Errorf("Error recovering IPv4Address. Got <%v> instead of <4.5.6.7>", recoveredMessage.GetIPAddressAVP("Igor-myTestAllGrouped.Igor-myIPv4Address"))
	}
}

func TestDiameterMessageJSON(t *testing.T) {
	jDiameterMessage := `
	{
		"IsRequest": true,
		"IsProxyable": false,
		"IsError": false,
		"IsRetransmission": false,
		"CommandCode": 2000,
		"ApplicationId": 1000,
		"avps":[
			{
			  "Igor-myTestAllGrouped": [
  				{"Igor-myOctetString": "0102030405060708090a0b"},
  				{"Igor-myInteger32": -99},
  				{"Igor-myInteger64": -99},
  				{"Igor-myUnsigned32": 99},
  				{"Igor-myUnsigned64": 99},
  				{"Igor-myFloat32": 99.9},
  				{"Igor-myFloat64": 99.9},
  				{"Igor-myAddress": "1.2.3.4"},
  				{"Igor-myTime": "1966-11-26T03:34:08 UTC"},
  				{"Igor-myString": "Hello, world!"},
  				{"Igor-myDiameterIdentity": "Diameter@identity"},
  				{"Igor-myDiameterURI": "Diameter@URI"},
  				{"Igor-myIPFilterRule": "allow all"},
  				{"Igor-myIPv4Address": "4.5.6.7"},
  				{"Igor-myIPv6Address": "bebe:cafe::0"},
  				{"Igor-myIPv6Prefix": "bebe:cafe::0/128"},
  				{"Igor-myEnumerated": "two"}
			  ]
			}
		]
	}
	`

	// Read JSON to DiameterMessage
	var diameterMessage DiameterMessage
	err := json.Unmarshal([]byte(jDiameterMessage), &diameterMessage)
	if err != nil {
		t.Errorf("unmarshal error for diameter message: %s", err)
	}
	diameterMessage.Tidy()

	// Write Diameter message as JSON
	jNewDiameterMessage, _ := json.Marshal(&diameterMessage)
	if !strings.Contains(string(jNewDiameterMessage), "TestApplication") || !strings.Contains(string(jNewDiameterMessage), "TestRequest") {
		t.Errorf("marshalled json does not contain the tidied attributes")
	}

	var jBytes bytes.Buffer
	if err := json.Indent(&jBytes, []byte(jNewDiameterMessage), "", "    "); err != nil {
		t.Errorf("prettyprint error %s", err)
	}

	// Uncoment this to see the result
	// fmt.Println(jBytes.String())
}

func TestCopyDiameterMessage(t *testing.T) {

	jDiameterMessage := `
	{
		"IsRequest": true,
		"IsProxyable": false,
		"IsError": false,
		"IsRetransmission": false,
		"CommandCode": 2000,
		"ApplicationId": 1000,
		"avps":[
			{"Session-Id":"session-id"},
			{"Destination-Realm":"igorsuperserver"},
			{"Auth-Application-Id":1000},
			{"Vendor-Id": 1001},
			{"Subscription-Id":[
				{"Subscription-Id-Type": "EndUserE164"},
				{"Subscription-Id-Data": "the-subscription-id"}
				]
			},
			{"User-Name":"francisco"},
			{"Framed-IP-Address":"1.1.1.1"},
			{"Igor-Command": "Echo"},
			{
			  "Igor-myTestAllGrouped": [
  				{"Igor-myOctetString": "0102030405060708090a0b"},
  				{"Igor-myInteger32": -99},
  				{"Igor-myInteger64": -99},
  				{"Igor-myUnsigned32": 99},
  				{"Igor-myUnsigned64": 99},
  				{"Igor-myFloat32": 99.9},
  				{"Igor-myFloat64": 99.9},
  				{"Igor-myAddress": "1.2.3.4"},
  				{"Igor-myTime": "1966-11-26T03:34:08 UTC"},
  				{"Igor-myString": "Hello, world!"},
  				{"Igor-myDiameterIdentity": "Diameter@identity"},
  				{"Igor-myDiameterURI": "Diameter@URI"},
  				{"Igor-myIPFilterRule": "allow all"},
  				{"Igor-myIPv4Address": "4.5.6.7"},
  				{"Igor-myIPv6Address": "bebe:cafe::0"},
  				{"Igor-myIPv6Prefix": "bebe:cafe::0/128"},
  				{"Igor-myEnumerated": "two"}
			  ]
			}
		]
	}`

	// Read JSON to DiameterMessage
	var diameterMessage DiameterMessage
	err := json.Unmarshal([]byte(jDiameterMessage), &diameterMessage)
	if err != nil {
		t.Errorf("unmarshal error for diameter message: %s", err)
	}
	diameterMessage.Tidy()

	positiveCopy := diameterMessage.Copy([]string{"Igor-myTestAllGrouped"}, nil)
	embeddedAttribute, err := positiveCopy.GetAVPFromPath("Igor-myTestAllGrouped.Igor-myEnumerated")
	if err != nil {
		t.Fatalf("could not get embedded attribute after positive copy: %s", err)
	}
	if embeddedAttribute.GetInt() != 2 {
		t.Fatal("bad balue for emvedded attribute after positive copy")
	}

	negativeCopy := diameterMessage.Copy(nil, []string{"Session-Id"})
	if negativeCopy.GetStringAVP("Session-Id") != "" {
		t.Fatal("Session-Id found after negative copy")
	}
	if negativeCopy.GetIntAVP("Vendor-Id") != 1001 {
		t.Fatal("Attribute not found after negative copy")
	}
}

func TestCheckDiameterMessage(t *testing.T) {

	jDiameterMessage := `
	{
		"IsRequest": true,
		"IsProxyable": false,
		"IsError": false,
		"IsRetransmission": false,
		"CommandCode": 2000,
		"ApplicationId": 1000,
		"avps":[
			{"Session-Id":"session-id"},
			{"Destination-Realm":"igorsuperserver"},
			{"Auth-Application-Id":1000},
			{"Vendor-Id": 1001},
			{"Subscription-Id":[
				{"Subscription-Id-Type": "EndUserE164"},
				{"Subscription-Id-Data": "the-subscription-id"}
				]
			},
			{"Framed-IP-Address":"1.1.1.1"},
			{"Igor-Command": "Echo"},
			{
			  "Igor-myTestAllGrouped": [
  				{"Igor-myOctetString": "0102030405060708090a0b"},
  				{"Igor-myInteger32": -99},
  				{"Igor-myInteger64": -99},
  				{"Igor-myUnsigned32": 99},
  				{"Igor-myUnsigned64": 99},
  				{"Igor-myFloat32": 99.9},
  				{"Igor-myFloat64": 99.9},
  				{"Igor-myAddress": "1.2.3.4"},
  				{"Igor-myTime": "1966-11-26T03:34:08 UTC"},
  				{"Igor-myString": "Hello, world!"},
  				{"Igor-myDiameterIdentity": "Diameter@identity"},
  				{"Igor-myDiameterURI": "Diameter@URI"},
  				{"Igor-myIPFilterRule": "allow all"},
  				{"Igor-myIPv4Address": "4.5.6.7"},
  				{"Igor-myIPv6Address": "bebe:cafe::0"},
  				{"Igor-myIPv6Prefix": "bebe:cafe::0/128"},
  				{"Igor-myEnumerated": "two"}
			  ]
			}
		]
	}`

	// Read JSON to DiameterMessage
	var diameterMessage DiameterMessage
	err := json.Unmarshal([]byte(jDiameterMessage), &diameterMessage)
	if err != nil {
		t.Fatalf("unmarshal error for diameter message: %s", err)
	}
	diameterMessage.Tidy()
	diameterMessage.AddOriginAVPs(config.GetPolicyConfigInstance("testClient"))

	// Initially, the message is valid
	err = diameterMessage.CheckAttributes()
	if err != nil {
		t.Fatalf("Check error: %s", err)
	}

	// Add an attribute not in the spec
	diameterMessage.Add("Igor-myOctetString", "00112233")
	err = diameterMessage.CheckAttributes()
	if err == nil {
		t.Fatal("unspecified attribute not detected afther Check()")
	}
	// Remove the attribute and delete another one which has minoccurs: 1
	diameterMessage.
		DeleteAllAVP("Igor-myOctetString").
		DeleteAllAVP("Vendor-Id")

	err = diameterMessage.CheckAttributes()
	if err == nil {
		t.Fatal("missing attribute not detected afther CheckAttributes()")
	}

	// Check error in grouped attribute. the Subscription-Id-Type will be missing
	diameterMessage.DeleteAllAVP("Subscription-Id")
	sidData, _ := NewAVP("Subscription-Id-Data", "the subscriptionId")
	savp, _ := NewAVP("Subscription-Id", []DiameterAVP{*sidData})
	err = savp.Check()
	if err == nil {
		t.Fatal("missing attribute in Group not detected after Check()")
	}
	diameterMessage.AddAVP(savp)
	err = diameterMessage.CheckAttributes()
	if err == nil {
		t.Fatal("missing attribute in Message not detected after CheckAttributes()")
	}

	// Add missing attribute
	sidType, _ := NewAVP("Subscription-Id-Type", "EndUserE164")
	savp, _ = NewAVP("Subscription-Id", []DiameterAVP{*sidData, *sidType})
	err = savp.Check()
	if err != nil {
		t.Fatal("error checking subscription-id grouped attribute")
	}
	diameterMessage.DeleteAllAVP("Subscription-Id").AddAVP(savp)
	err = diameterMessage.CheckAttributes()
	if err == nil {
		t.Fatal("error in CheckAttributes() in well-formed message")
	}

	// Too many session ids
	diameterMessage.Add("Session-Id", "another-session")
	err = diameterMessage.CheckAttributes()
	if err == nil {
		t.Fatal("undetected duplicate Session-Id")
	}
}