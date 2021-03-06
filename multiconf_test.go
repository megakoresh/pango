package pango

import (
	"encoding/xml"
	"testing"

	"github.com/PaloAltoNetworks/pango/objs/addr"
)

const invalidMultiConfigResp = `
<response status="error" code="13">
    <response status="success" code="20" id="1109"><msg>command succeeded</msg></response>
    <response status="success" code="7"><msg>Object doesn't exist</msg></response>
    <response status="error" code="12" id="1110"><msg><line><![CDATA[ test-new unexpected here]]></line></msg></response>
</response>
`

const okMultiConfigResp = `
<response status="success" code="20">
    <response status="success" code="20" id="1109"><msg>command succeeded</msg></response>
    <response status="success" code="7"><msg>Object doesn't exist</msg></response>
    <response status="success" code="20" id="1110"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
</response>
`

func TestPrepareMultiConfigure(t *testing.T) {
	var err error
	fw := &Firewall{Client: Client{
		rb: [][]byte{
			[]byte(okMultiConfigResp),
		},
	}}
	if err = fw.Initialize(); err != nil {
		t.Errorf("Initialize failed: %s", err)
		return
	}

	fw.PrepareMultiConfigure(2)
	if len(fw.MultiConfigure.Reqs) != 0 {
		t.Errorf("len(%d) not 0", len(fw.MultiConfigure.Reqs))
	} else if cap(fw.MultiConfigure.Reqs) != 2 {
		t.Errorf("cap(%d) not 2", cap(fw.MultiConfigure.Reqs))
	}
}

func TestOkMultiConfig(t *testing.T) {
	var err error
	fw := &Firewall{Client: Client{
		rb: [][]byte{
			[]byte(okMultiConfigResp),
		},
		MultiConfigure: &MultiConfigure{
			Reqs: make([]MultiConfigureRequest, 0, 1),
		},
	}}
	if err = fw.Initialize(); err != nil {
		t.Errorf("Initialize failed: %s", err)
		return
	}

	o := addr.Entry{
		Name:        "foo",
		Value:       "1.2.3.4",
		Type:        "ip-netmask",
		Description: "Don't mention it",
	}

	if err = fw.Objects.Address.Set("vsys2", o); err != nil {
		t.Errorf("Failed to set address object: %s", err)
	} else if len(fw.MultiConfigure.Reqs) != 1 {
		t.Errorf("Didn't save request into multi-configure: %#v", fw.MultiConfigure)
	} else {
		resp, err := fw.SendMultiConfigure()
		if err != nil {
			t.Errorf("Failed send multiconfigure: %s", err)
		} else if !resp.Ok() {
			t.Errorf("Response is failed: %s", resp.Error())
		} else if len(fw.rp) == 0 {
			t.Errorf("No url values sent..?")
		} else {
			vals := fw.rp[0]
			if vals.Get("action") != "multi-config" {
				t.Errorf("Action is %q, not 'multi-config'", vals.Get("action"))
			} else if vals.Get("type") != "config" {
				t.Errorf("Type is %q, not 'config'", vals.Get("type"))
			} else if vals.Get("element") == "" {
				t.Errorf("Element is unset it seems")
			}
		}
	}
}

func TestUnmarshalOkMultiConfigureResponse(t *testing.T) {
	r := MultiConfigureResponse{}
	if err := xml.Unmarshal([]byte(okMultiConfigResp), &r); err != nil {
		t.Errorf("Failed unmarshal: %s", err)
	}

	if !r.Ok() {
		t.Errorf("response is %t, not ok", r.Ok())
	} else if len(r.Results) != 6 {
		t.Errorf("results are len %d, not 6", len(r.Results))
	} else if r.Results[0].Id != 1109 {
		t.Errorf("r0 id is %d, not 1109", r.Results[0].Id)
	} else if r.Results[1].Code != 7 {
		t.Errorf("r1 code is %d, not 7", r.Results[1].Code)
	} else if r.Results[1].Id != 0 {
		t.Errorf("r1 id is %d, not 0", r.Results[1].Id)
	} else if !r.Results[1].Ok() {
		t.Errorf("r1 is %t, not ok", r.Results[1].Ok())
	} else if r.Error() != "" {
		t.Errorf("response error(%q), is not empty string", r.Error())
	}
}

func TestUnmarshalInvalidMultiConfigureResponse(t *testing.T) {
	r := MultiConfigureResponse{}
	if err := xml.Unmarshal([]byte(invalidMultiConfigResp), &r); err != nil {
		t.Errorf("Failed unmarshal: %s", err)
	}

	if r.Ok() {
		t.Errorf("Response is ok, should be failed")
	} else if r.Code != 13 {
		t.Errorf("Expected code(13), got code(%d)", r.Code)
	} else if len(r.Results) != 3 {
		t.Errorf("Have %d results, expected 3", len(r.Results))
	} else if r.Results[0].Code != 20 {
		t.Errorf("R0 has code %d, not 20", r.Results[0].Code)
	} else if r.Results[0].Status != "success" {
		t.Errorf("R0 has status %q, not success", r.Results[0].Status)
	} else if r.Results[0].Message() != "command succeeded" {
		t.Errorf("R0 has message %q, not 'command succeeded'", r.Results[0].Message())
	} else if r.Results[1].Code != 7 {
		t.Errorf("R1 has code %d, not 7", r.Results[1].Code)
	} else if r.Results[1].Status != "success" {
		t.Errorf("R1 has status %q, not success", r.Results[1].Status)
	} else if r.Results[1].Message() != "Object doesn't exist" {
		t.Errorf("R1 has message %q, not 'Object doesn't exist'", r.Results[1].Message())
	} else if r.Results[2].Code != 12 {
		t.Errorf("R2 has code %d, not 12", r.Results[2].Code)
	} else if r.Results[2].Status != "error" {
		t.Errorf("R2 has status %q, not 'error'", r.Results[2].Status)
	} else if r.Results[2].Message() != "test-new unexpected here" {
		t.Errorf("R2 has message %q, not 'test-new unexpected here'", r.Results[2].Message())
	} else if r.Error() != "test-new unexpected here" {
		t.Errorf("response has message %q, not 'test-new unexpected here'", r.Results[2].Message())
	}
}

/*
$ cat invalid-req.xml
<multi-configure-request>
 <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test-shared-new']" id="1109">
    <entry name="test-shared-new">
      <ip-netmask>51.5.5.5</ip-netmask>
    </entry>
  </request>
  <request cmd="delete" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test2']"/>
  <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test-new']" id="1110">
    <entry name="test-new">
    </entry>
  </request>
  <request cmd="set" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/9']">
      <comment>mshen test comment</comment>
    <tap/>
  </request>
  <request cmd="set" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/8']">
    <virtual-wire/>
  </request>
  <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/1']">
    <entry name="ethernet1/1">
      <tap/>
      <comment>mshen test comment</comment>
    </entry>
  </request>
</multi-configure-request>

$curl -k -X POST -u admin:Paloalto1 'https://10.5.104.123/api/' -d 'type=config' -d 'action=multi-config' --data-urlencode element@invalid-req.xml
<response status="error" code="13">
    <response status="success" code="20" id="1109"><msg>command succeeded</msg></response>
    <response status="success" code="7"><msg>Object doesn't exist</msg></response>
    <response status="error" code="12" id="1110"><msg><line><![CDATA[ test-new unexpected here]]></line></msg></response>
</response>


good-req
<multi-configure-request>
  <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test-shared-new']" id="1109">
    <entry name="test-shared-new">
      <ip-netmask>51.5.5.55</ip-netmask>
    </entry>
  </request>
  <request cmd="delete" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test2']"/>
  <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address/entry[@name='test-new']" id="1110">
    <entry name="test-new">
     <fqdn>Iamgood.com</fqdn>
    </entry>
  </request>
  <request cmd="set" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/9']">
      <comment>mshen test comment</comment>
    <tap/>
  </request>
  <request cmd="set" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/8']">
    <virtual-wire/>
  </request>
  <request cmd="edit" obj="/config/devices/entry[@name='localhost.localdomain']/network/interface/ethernet/entry[@name='ethernet1/1']">
    <entry name="ethernet1/1">
      <tap/>
      <comment>mshen test comment</comment>
    </entry>
  </request>
<multi-configure-request>
after running this: response:

<response status="success" code="20">
    <response status="success" code="20" id="1109"><msg>command succeeded</msg></response>
    <response status="success" code="7"><msg>Object doesn't exist</msg></response>
    <response status="success" code="20" id="1110"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
    <response status="success" code="20"><msg>command succeeded</msg></response>
</response>

and UI shows the change with config logs
*/
