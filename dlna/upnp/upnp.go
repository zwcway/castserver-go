package upnp

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var serviceURNRegexp *regexp.Regexp = regexp.MustCompile(`^urn:(.*):service:(\w+):(\d+)$`)

type ServiceURN struct {
	Auth    string
	Type    string
	Version uint64
}

func (me ServiceURN) String() string {
	return fmt.Sprintf("urn:%s:service:%s:%d", me.Auth, me.Type, me.Version)
}

func ParseServiceURN(s string) (ret ServiceURN, err error) {
	matches := serviceURNRegexp.FindStringSubmatch(s)
	if matches == nil || len(matches) != 4 {
		err = fmt.Errorf("unknown URN %s", s)
		return
	}
	ret.Auth = matches[1]
	ret.Type = matches[2]
	ret.Version, err = strconv.ParseUint(matches[3], 0, 32)
	return
}

type SoapAction struct {
	ServiceURN
	Action string
}

func ParseActionHTTPHeader(s string) (ret SoapAction, err error) {
	if len(s) < 3 {
		return
	}
	if s[0] != '"' || s[len(s)-1] != '"' {
		return
	}
	s = s[1 : len(s)-1]
	hashIndex := strings.LastIndex(s, "#")
	if hashIndex == -1 {
		return
	}
	ret.Action = s[hashIndex+1:]
	ret.ServiceURN, err = ParseServiceURN(s[:hashIndex])
	return
}

type SpecVersion struct {
	Major int `xml:"major"`
	Minor int `xml:"minor"`
}

type Icon struct {
	Mimetype string `xml:"mimetype"`
	Width    int    `xml:"width"`
	Height   int    `xml:"height"`
	Depth    int    `xml:"depth"`
	URL      string `xml:"url"`
}

type Service struct {
	XMLName     xml.Name `xml:"service"`
	ServiceType string   `xml:"serviceType"`
	ServiceId   string   `xml:"serviceId"`
	SCPDURL     string
	ControlURL  string `xml:"controlURL"`
	EventSubURL string `xml:"eventSubURL"`
}

type Device struct {
	DeviceType       string `xml:"deviceType"`
	FriendlyName     string `xml:"friendlyName"`
	Manufacturer     string `xml:"manufacturer"`
	ModelName        string `xml:"modelName"`
	ModelDescription string `xml:"modelDescription"`
	UDN              string
	ServiceList      []Service `xml:"serviceList>service"`
}

type DeviceDesc struct {
	XMLName     xml.Name    `xml:"urn:schemas-upnp-org:device-1-0 root"`
	NSDLNA      string      `xml:"xmlns:dlna,attr"`
	NSSEC       string      `xml:"xmlns:sec,attr"`
	SpecVersion SpecVersion `xml:"specVersion"`
	Device      Device      `xml:"device"`
}

type ServiceDesc struct {
	XMLName     xml.Name    `xml:"urn:schemas-upnp-org:service-1-0"`
	SpecVersion SpecVersion `xml:"specVersion"`
	ActionList  []Action    `xml:"actionList"`
}

type Error struct {
	XMLName xml.Name `xml:"urn:schemas-upnp-org:control-1-0 Error"`
	Code    uint     `xml:"errorCode"`
	Desc    string   `xml:"errorDescription"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d %s", e.Code, e.Desc)
}

const (
	InvalidActionErrorCode        = 401
	ActionFailedErrorCode         = 501
	ArgumentValueInvalidErrorCode = 600
)

var (
	InvalidActionError        = upnpNewErrorf(401, "Invalid Action")
	ArgumentValueInvalidError = upnpNewErrorf(600, "The argument value is invalid")
)

func upnpNewErrorf(code uint, tpl string, args ...interface{}) *Error {
	return &Error{Code: code, Desc: fmt.Sprintf(tpl, args...)}
}

func upnpNewError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return upnpNewErrorf(ActionFailedErrorCode, err.Error())
}

type Action struct {
	Name      string     `xml:"name"`
	Arguments []Argument `xml:"argumentList"`
}

type Argument struct {
	Name            string `xml:"name"`
	Direction       string `xml:"direction"`
	RelatedStateVar string `xml:"relatedStateVariable"`
}

type SCPD struct {
	XMLName           xml.Name        `xml:"urn:schemas-upnp-org:service-1-0 scpd"`
	SpecVersion       SpecVersion     `xml:"specVersion"`
	ActionList        []Action        `xml:"actionList>action"`
	ServiceStateTable []StateVariable `xml:"serviceStateTable>stateVariable"`
}

type StateVariable struct {
	SendEvents    string    `xml:"sendEvents,attr"`
	Name          string    `xml:"name"`
	DataType      string    `xml:"dataType"`
	AllowedValues *[]string `xml:"allowedValueList>allowedValue,omitempty"`
}
