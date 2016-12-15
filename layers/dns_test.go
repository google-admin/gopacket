// Copyright 2012, Google, Inc. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package layers

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/google/gopacket"
)

// testPacketDNSRegression is the packet:
//   11:08:05.708342 IP 109.194.160.4.57766 > 95.211.92.14.53: 63000% [1au] A? picslife.ru. (40)
//      0x0000:  0022 19b6 7e22 000f 35bb 0b40 0800 4500  ."..~"..5..@..E.
//      0x0010:  0044 89c4 0000 3811 2f3d 6dc2 a004 5fd3  .D....8./=m..._.
//      0x0020:  5c0e e1a6 0035 0030 a597 f618 0010 0001  \....5.0........
//      0x0030:  0000 0000 0001 0870 6963 736c 6966 6502  .......picslife.
//      0x0040:  7275 0000 0100 0100 0029 1000 0000 8000  ru.......)......
//      0x0050:  0000                                     ..
var testPacketDNSRegression = []byte{
	0x00, 0x22, 0x19, 0xb6, 0x7e, 0x22, 0x00, 0x0f, 0x35, 0xbb, 0x0b, 0x40, 0x08, 0x00, 0x45, 0x00,
	0x00, 0x44, 0x89, 0xc4, 0x00, 0x00, 0x38, 0x11, 0x2f, 0x3d, 0x6d, 0xc2, 0xa0, 0x04, 0x5f, 0xd3,
	0x5c, 0x0e, 0xe1, 0xa6, 0x00, 0x35, 0x00, 0x30, 0xa5, 0x97, 0xf6, 0x18, 0x00, 0x10, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x08, 0x70, 0x69, 0x63, 0x73, 0x6c, 0x69, 0x66, 0x65, 0x02,
	0x72, 0x75, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x80, 0x00,
	0x00, 0x00,
}

func TestPacketDNSRegression(t *testing.T) {
	p := gopacket.NewPacket(testPacketDNSRegression, LinkTypeEthernet, testDecodeOptions)
	if p.ErrorLayer() != nil {
		t.Error("Failed to decode packet:", p.ErrorLayer().Error())
	}
	checkLayers(p, []gopacket.LayerType{LayerTypeEthernet, LayerTypeIPv4, LayerTypeUDP, LayerTypeDNS}, t)
}
func BenchmarkDecodePacketDNSRegression(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gopacket.NewPacket(testPacketDNSRegression, LinkTypeEthernet, gopacket.NoCopy)
	}
}

// response to `dig TXT google.com` over IPv4 link:
var testParseDNSTypeTXTValue = `v=spf1 include:_spf.google.com ~all`
var testParseDNSTypeTXT = []byte{
	0x02, 0x00, 0x00, 0x00, // PF_INET
	0x45, 0x00, 0x00, 0x73, 0x00, 0x00, 0x40, 0x00, 0x39, 0x11, 0x64, 0x98, 0xd0, 0x43, 0xde, 0xde,
	0x0a, 0xba, 0x23, 0x06, 0x00, 0x35, 0x81, 0xb2, 0x00, 0x5f, 0xdc, 0xb5, 0x98, 0x71, 0x81, 0x80,
	0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x06, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x03,
	0x63, 0x6f, 0x6d, 0x00, 0x00, 0x10, 0x00, 0x01, 0xc0, 0x0c, 0x00, 0x10, 0x00, 0x01, 0x00, 0x00,
	0x0e, 0x10, 0x00, 0x24, 0x23, 0x76, 0x3d, 0x73, 0x70, 0x66, 0x31, 0x20, 0x69, 0x6e, 0x63, 0x6c,
	0x75, 0x64, 0x65, 0x3a, 0x5f, 0x73, 0x70, 0x66, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x63, 0x6f, 0x6d, 0x20, 0x7e, 0x61, 0x6c, 0x6c, 0x00, 0x00, 0x29, 0x10, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00,
}

func TestParseDNSTypeTXT(t *testing.T) {
	p := gopacket.NewPacket(testParseDNSTypeTXT, LinkTypeNull, testDecodeOptions)
	if p.ErrorLayer() != nil {
		t.Error("Failed to decode packet:", p.ErrorLayer().Error())
	}
	checkLayers(p, []gopacket.LayerType{LayerTypeLoopback, LayerTypeIPv4, LayerTypeUDP, LayerTypeDNS}, t)
	answers := p.Layer(LayerTypeDNS).(*DNS).Answers
	if len(answers) != 1 {
		t.Error("Failed to parse 1 DNS answer")
	}
	if len(answers[0].TXTs) != 1 {
		t.Error("Failed to parse 1 TXT record")
	}
	txt := string(answers[0].TXTs[0])
	if txt != testParseDNSTypeTXTValue {
		t.Errorf("Incorrect TXT value, expected %q, got %q", testParseDNSTypeTXTValue, txt)
	}
}

func testQuestionEqual(t *testing.T, i int, exp, got DNSQuestion) {
	if !bytes.Equal(exp.Name, got.Name) {
		t.Errorf("expected Questions[%d].Name = %v, got %v", i, string(exp.Name), string(got.Name))
	}
	if exp.Type != got.Type {
		t.Errorf("expected Questions[%d].Type = %v, got %v", i, exp.Type, got.Type)
	}
	if exp.Class != got.Class {
		t.Errorf("expected Questions[%d].Class = %v, got %v", i, exp.Class, got.Class)
	}
}

func testResourceEqual(t *testing.T, i int, name string, exp, got DNSResourceRecord) {
	if !bytes.Equal(exp.Name, got.Name) {
		t.Errorf("expected %s[%d].Name = %v, got %v", name, i, string(exp.Name), string(got.Name))
	}

	if exp.Type != got.Type {
		t.Errorf("expected %s[%d].Type = %v, got %v", name, i, exp.Type, got.Type)
	}

	if exp.Class != got.Class {
		t.Errorf("expected %s[%d].Class = %v, got %v", name, i, exp.Class, got.Class)
	}

	if exp.TTL != got.TTL {
		t.Errorf("expected %s[%d].TTL = %v, got %v", name, i, exp.TTL, got.TTL)
	}
	if exp.DataLength != got.DataLength {
		t.Errorf("expected %s[%d].DataLength = %v, got %v", name, i, exp.DataLength, got.DataLength)
	}

	// we don't check .Data

	if !exp.IP.Equal(got.IP) {
		t.Errorf("expected %s[%d].IP = %v, got %v", name, i, exp.IP, got.IP)
	}
	if !bytes.Equal(exp.NS, got.NS) {
		t.Errorf("expected %s[%d].NS = %v, got %v", name, i, exp.NS, got.NS)
	}
	if !bytes.Equal(exp.CNAME, got.CNAME) {
		t.Errorf("expected %s[%d].CNAME = %v, got %v", name, i, exp.CNAME, got.CNAME)
	}
	if !bytes.Equal(exp.PTR, got.PTR) {
		t.Errorf("expected %s[%d].PTR = %v, got %v", name, i, exp.PTR, got.PTR)
	}
	if len(exp.TXTs) != len(got.TXTs) {
		t.Errorf("expected %s[%d].TXTs = %v, got %v", name, i, exp.TXTs, got.TXTs)
	}
	for j := range exp.TXTs {
		if !bytes.Equal(exp.TXTs[j], got.TXTs[j]) {
			t.Errorf("expected %s[%d].TXTs[%d] = %v, got %v", name, i, j, exp.TXTs[j], got.TXTs[j])
		}
	}

	// SOA
	if !bytes.Equal(exp.SOA.MName, got.SOA.MName) {
		t.Errorf("expected %s[%d].SOA.MName = %v, got %v", name, i, exp.SOA.MName, got.SOA.MName)
	}
	if !bytes.Equal(exp.SOA.RName, got.SOA.RName) {
		t.Errorf("expected %s[%d].SOA.RName = %v, got %v", name, i, exp.SOA.RName, got.SOA.RName)
	}
	if exp.SOA.Serial != got.SOA.Serial {
		t.Errorf("expected %s[%d].SOA.Serial = %v, got %v", name, i, exp.SOA.Serial, got.SOA.Serial)
	}
	if exp.SOA.Refresh != got.SOA.Refresh {
		t.Errorf("expected %s[%d].SOA.Refresh = %v, got %v", name, i, exp.SOA.Refresh, got.SOA.Refresh)
	}
	if exp.SOA.Retry != got.SOA.Retry {
		t.Errorf("expected %s[%d].SOA.Retry = %v, got %v", name, i, exp.SOA.Retry, got.SOA.Retry)
	}
	if exp.SOA.Expire != got.SOA.Expire {
		t.Errorf("expected %s[%d].SOA.Expire = %v, got %v", name, i, exp.SOA.Expire, got.SOA.Expire)
	}
	if exp.SOA.Minimum != got.SOA.Minimum {
		t.Errorf("expected %s[%d].SOA.Minimum = %v, got %v", name, i, exp.SOA.Minimum, got.SOA.Minimum)
	}

	// SRV
	if !bytes.Equal(exp.SRV.Name, got.SRV.Name) {
		t.Errorf("expected %s[%d].SRV.Name = %v, got %v", name, i, exp.SRV.Name, got.SRV.Name)
	}
	if exp.SRV.Weight != got.SRV.Weight {
		t.Errorf("expected %s[%d].SRV.Weight = %v, got %v", name, i, exp.SRV.Weight, got.SRV.Weight)
	}
	if exp.SRV.Port != got.SRV.Port {
		t.Errorf("expected %s[%d].SRV.Port = %v, got %v", name, i, exp.SRV.Port, got.SRV.Port)
	}
	// MX
	if !bytes.Equal(exp.MX.Name, got.MX.Name) {
		t.Errorf("expected %s[%d].MX.Name = %v, got %v", name, i, exp.MX.Name, got.MX.Name)
	}
	if exp.MX.Preference != got.MX.Preference {
		t.Errorf("expected %s[%d].MX.Preference = %v, got %v", name, i, exp.MX.Preference, got.MX.Preference)
	}
}

func testDNSEqual(t *testing.T, exp, got *DNS) {
	if exp.ID != got.ID {
		t.Errorf("expected ID = %v, got %v", exp.ID, got.ID)
	}
	if exp.AA != got.AA {
		t.Errorf("expected AA = %v, got %v", exp.AA, got.AA)
	}
	if exp.OpCode != got.OpCode {
		t.Errorf("expected OpCode = %v, got %v", exp.OpCode, got.OpCode)
	}
	if exp.AA != got.AA {
		t.Errorf("expected AA = %v, got %v", exp.AA, got.AA)
	}
	if exp.TC != got.TC {
		t.Errorf("expected TC = %v, got %v", exp.TC, got.TC)
	}
	if exp.RD != got.RD {
		t.Errorf("expected RD = %v, got %v", exp.RD, got.RD)
	}
	if exp.RA != got.RA {
		t.Errorf("expected RA = %v, got %v", exp.RA, got.RA)
	}
	if exp.Z != got.Z {
		t.Errorf("expected Z = %v, got %v", exp.Z, got.Z)
	}
	if exp.ResponseCode != got.ResponseCode {
		t.Errorf("expected ResponseCode = %v, got %v", exp.ResponseCode, got.ResponseCode)
	}
	if exp.QDCount != got.QDCount {
		t.Errorf("expected QDCount = %v, got %v", exp.QDCount, got.QDCount)
	}
	if exp.ANCount != got.ANCount {
		t.Errorf("expected ANCount = %v, got %v", exp.ANCount, got.ANCount)
	}
	if exp.ANCount != got.ANCount {
		t.Errorf("expected ANCount = %v, got %v", exp.ANCount, got.ANCount)
	}
	if exp.NSCount != got.NSCount {
		t.Errorf("expected NSCount = %v, got %v", exp.NSCount, got.NSCount)
	}
	if exp.ARCount != got.ARCount {
		t.Errorf("expected ARCount = %v, got %v", exp.ARCount, got.ARCount)
	}

	if len(exp.Questions) != len(got.Questions) {
		t.Errorf("expected %d Questions, got %d", len(exp.Questions), len(got.Questions))
	}
	for i := range exp.Questions {
		testQuestionEqual(t, i, exp.Questions[i], got.Questions[i])
	}

	if len(exp.Answers) != len(got.Answers) {
		t.Errorf("expected %d Answers, got %d", len(exp.Answers), len(got.Answers))
	}
	for i := range exp.Answers {
		testResourceEqual(t, i, "Answers", exp.Answers[i], got.Answers[i])
	}

	if len(exp.Authorities) != len(got.Authorities) {
		t.Errorf("expected %d Answers, got %d", len(exp.Authorities), len(got.Authorities))
	}
	for i := range exp.Authorities {
		testResourceEqual(t, i, "Authorities", exp.Authorities[i], got.Authorities[i])
	}

	if len(exp.Additionals) != len(got.Additionals) {
		t.Errorf("expected %d Additionals, got %d", len(exp.Additionals), len(got.Additionals))
	}
	for i := range exp.Additionals {
		testResourceEqual(t, i, "Additionals", exp.Additionals[i], got.Additionals[i])
	}
}

func TestDNSEncodeQuery(t *testing.T) {
	dns := &DNS{ID: 1234, OpCode: DNSOpCodeQuery, RD: true}
	dns.Questions = append(dns.Questions,
		DNSQuestion{
			Name:  []byte("example1.com"),
			Type:  DNSTypeA,
			Class: DNSClassIN,
		})

	dns.Questions = append(dns.Questions,
		DNSQuestion{
			Name:  []byte("example2.com"),
			Type:  DNSTypeA,
			Class: DNSClassIN,
		})

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buf, opts, dns)
	if err != nil {
		t.Fatal(err)
	}
	if int(dns.QDCount) != len(dns.Questions) {
		t.Errorf("fix lengths did not adjust QDCount, expected %d got %d", len(dns.Questions), dns.QDCount)
	}

	p2 := gopacket.NewPacket(buf.Bytes(), LayerTypeDNS, testDecodeOptions)
	dns2 := p2.Layer(LayerTypeDNS).(*DNS)
	testDNSEqual(t, dns, dns2)
}

func TestDNSEncodeResponse(t *testing.T) {
	dns := &DNS{ID: 1234, QR: true, OpCode: DNSOpCodeQuery,
		AA: true, RD: true, RA: true}
	dns.Questions = append(dns.Questions,
		DNSQuestion{
			Name:  []byte("example1.com"),
			Type:  DNSTypeA,
			Class: DNSClassIN,
		})
	dns.Questions = append(dns.Questions,
		DNSQuestion{
			Name:  []byte("www.example2.com"),
			Type:  DNSTypeAAAA,
			Class: DNSClassIN,
		})

	dns.Answers = append(dns.Answers,
		DNSResourceRecord{
			Name:  []byte("example1.com"),
			Type:  DNSTypeA,
			Class: DNSClassIN,
			TTL:   1024,
			IP:    net.IP([]byte{1, 2, 3, 4}),
		})

	dns.Answers = append(dns.Answers,
		DNSResourceRecord{
			Name:  []byte("www.example2.com"),
			Type:  DNSTypeAAAA,
			Class: DNSClassIN,
			TTL:   1024,
			IP:    net.IP([]byte{5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}),
		})

	dns.Answers = append(dns.Answers,
		DNSResourceRecord{
			Name:  []byte("www.example2.com"),
			Type:  DNSTypeCNAME,
			Class: DNSClassIN,
			TTL:   1024,
			CNAME: []byte("example2.com"),
		})

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err := gopacket.SerializeLayers(buf, opts, dns)
	if err != nil {
		t.Fatal(err)
	}
	if int(dns.ANCount) != len(dns.Answers) {
		t.Errorf("fix lengths did not adjust ANCount, expected %d got %d", len(dns.Answers), dns.ANCount)
	}
	for i, a := range dns.Answers {
		if a.DataLength == 0 {
			t.Errorf("fix lengths did not adjust Answers[%d].DataLength", i)
		}
	}

	p2 := gopacket.NewPacket(buf.Bytes(), LayerTypeDNS, testDecodeOptions)
	dns2 := p2.Layer(LayerTypeDNS).(*DNS)
	testDNSEqual(t, dns, dns2)
}

// testDNSMalformedPacket is the packet:
//   10:30:00.389666 IP 10.77.43.131.60718 > 10.1.0.17.53: 18245 updateD [b2&3=0x5420] [18516a] [12064q] [21584n] [12081au][|domain]
//   	0x0000:  0000 0101 0000 4e96 1476 afa1 0800 4500  ......N..v....E.
//   	0x0010:  0039 d431 0000 f311 b3a0 0a4d 2b83 0a01  .9.1.......M+...
//   	0x0020:  0011 ed2e 0035 0025 0832 4745 5420 2f20  .....5.%.2GET./.
//   	0x0030:  4854 5450 2f31 2e31 0d0a 486f 7374 3a20  HTTP/1.1..Host:.
//   	0x0040:  7777 770d 0a0d 0a                        www....
var testDNSMalformedPacket = []byte{
	0x00, 0x00, 0x01, 0x01, 0x00, 0x00, 0x4e, 0x96, 0x14, 0x76, 0xaf, 0xa1, 0x08, 0x00, 0x45, 0x00,
	0x00, 0x39, 0xd4, 0x31, 0x00, 0x00, 0xf3, 0x11, 0xb3, 0xa0, 0x0a, 0x4d, 0x2b, 0x83, 0x0a, 0x01,
	0x00, 0x11, 0xed, 0x2e, 0x00, 0x35, 0x00, 0x25, 0x08, 0x32, 0x47, 0x45, 0x54, 0x20, 0x2f, 0x20,
	0x48, 0x54, 0x54, 0x50, 0x2f, 0x31, 0x2e, 0x31, 0x0d, 0x0a, 0x48, 0x6f, 0x73, 0x74, 0x3a, 0x20,
	0x77, 0x77, 0x77, 0x0d, 0x0a, 0x0d, 0x0a,
}

func TestDNSMalformedPacket(t *testing.T) {
	p := gopacket.NewPacket(testDNSMalformedPacket, LinkTypeEthernet, testDecodeOptions)
	if errLayer := p.ErrorLayer(); errLayer == nil {
		t.Error("No error layer on invalid DNS name")
	} else if err := errLayer.Error(); !strings.Contains(err.Error(), "invalid index") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// testDNSMalformedPacket2 is the packet:
//   15:14:42.056054 IP 10.77.0.245.53 > 10.1.0.45.38769: 12625 zoneInit YXRRSet- [49833q],[|domain]
//   	0x0000:  0055 22af c637 0022 55ac deac 0800 4500  .U"..7."U.....E.
//   	0x0010:  0079 3767 4000 3911 f49d 0a4d 00f5 0a01  .y7g@.9....M....
//   	0x0020:  002d 0035 9771 0065 6377 3151 f057 c2a9  .-.5.q.ecw1Q.W..
//   	0x0030:  fc6e e86a beb0 f7d4 8599 373e b5f8 9db2  .n.j......7>....
//   	0x0040:  a399 21a1 9762 def1 def4 f5ab 5675 023e  ..!..b......Vu.>
//   	0x0050:  c9ca 304f 178a c2ad f2fc 677a 0e4c b892  ..0O......gz.L..
//   	0x0060:  ab71 09bb 1ea4 f7c4 fe47 7a39 868b 29a0  .q.......Gz9..).
//   	0x0070:  62c4 d184 5b4e 8817 4cc0 d1d0 d430 11d3  b...[N..L....0..
//   	0x0080:  d147 543f afc7 1a                        .GT?...
var testDNSMalformedPacket2 = []byte{
	0x00, 0x55, 0x22, 0xaf, 0xc6, 0x37, 0x00, 0x22, 0x55, 0xac, 0xde, 0xac, 0x08, 0x00, 0x45, 0x00,
	0x00, 0x79, 0x37, 0x67, 0x40, 0x00, 0x39, 0x11, 0xf4, 0x9d, 0x0a, 0x4d, 0x00, 0xf5, 0x0a, 0x01,
	0x00, 0x2d, 0x00, 0x35, 0x97, 0x71, 0x00, 0x65, 0x63, 0x77, 0x31, 0x51, 0xf0, 0x57, 0xc2, 0xa9,
	0xfc, 0x6e, 0xe8, 0x6a, 0xbe, 0xb0, 0xf7, 0xd4, 0x85, 0x99, 0x37, 0x3e, 0xb5, 0xf8, 0x9d, 0xb2,
	0xa3, 0x99, 0x21, 0xa1, 0x97, 0x62, 0xde, 0xf1, 0xde, 0xf4, 0xf5, 0xab, 0x56, 0x75, 0x02, 0x3e,
	0xc9, 0xca, 0x30, 0x4f, 0x17, 0x8a, 0xc2, 0xad, 0xf2, 0xfc, 0x67, 0x7a, 0x0e, 0x4c, 0xb8, 0x92,
	0xab, 0x71, 0x09, 0xbb, 0x1e, 0xa4, 0xf7, 0xc4, 0xfe, 0x47, 0x7a, 0x39, 0x86, 0x8b, 0x29, 0xa0,
	0x62, 0xc4, 0xd1, 0x84, 0x5b, 0x4e, 0x88, 0x17, 0x4c, 0xc0, 0xd1, 0xd0, 0xd4, 0x30, 0x11, 0xd3,
	0xd1, 0x47, 0x54, 0x3f, 0xaf, 0xc7, 0x1a,
}

func TestDNSMalformedPacket2(t *testing.T) {
	p := gopacket.NewPacket(testDNSMalformedPacket2, LinkTypeEthernet, testDecodeOptions)
	if errLayer := p.ErrorLayer(); errLayer == nil {
		t.Error("No error layer on invalid DNS name")
	} else if err := errLayer.Error(); !strings.Contains(err.Error(), "offset pointer too high") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// testMalformedRootQuery is the packet:
//   08:31:18.143065 IP 10.77.0.26.53 > 10.1.0.233.65071: 59508- 0/13/3 (508)
//   	0x0000:  0055 22af c637 0022 55ac deac 0800 4500  .U"..7."U.....E.
//   	0x0010:  0218 76b2 4000 7211 7ad2 0a4d 001a 0a01  ..v.@.r.z..M....
//   	0x0020:  00e9 0035 fe2f 0204 b8f5 e874 8100 0001  ...5./.....t....
//   	0x0030:  0000 000d 0003 0c61 786b 7663 6863 7063  .......axkvchcpc
//   	0x0040:  7073 6c0a 7878 7878 7878 7878 7878 036e  psl.xxxxxxxxxx.n
//   	0x0050:  6574 0000 0100 0100 0002 0001 0000 0e10  et..............
//   	0x0060:  0014 016d 0c72 6f6f 742d 7365 7276 6572  ...m.root-server
//   	0x0070:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0080:  0014 0161 0c72 6f6f 742d 7365 7276 6572  ...a.root-server
//   	0x0090:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x00a0:  0014 0169 0c72 6f6f 742d 7365 7276 6572  ...i.root-server
//   	0x00b0:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x00c0:  0014 0162 0c72 6f6f 742d 7365 7276 6572  ...b.root-server
//   	0x00d0:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x00e0:  0014 016c 0c72 6f6f 742d 7365 7276 6572  ...l.root-server
//   	0x00f0:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0100:  0014 0166 0c72 6f6f 742d 7365 7276 6572  ...f.root-server
//   	0x0110:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0120:  0014 0167 0c72 6f6f 742d 7365 7276 6572  ...g.root-server
//   	0x0130:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0140:  0014 0164 0c72 6f6f 742d 7365 7276 6572  ...d.root-server
//   	0x0150:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0160:  0014 0168 0c72 6f6f 742d 7365 7276 6572  ...h.root-server
//   	0x0170:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x0180:  0014 0165 0c72 6f6f 742d 7365 7276 6572  ...e.root-server
//   	0x0190:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x01a0:  0014 016a 0c72 6f6f 742d 7365 7276 6572  ...j.root-server
//   	0x01b0:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x01c0:  0014 016b 0c72 6f6f 742d 7365 7276 6572  ...k.root-server
//   	0x01d0:  7303 6e65 7400 c02d 0002 0001 0000 0e10  s.net..-........
//   	0x01e0:  0014 0163 0c72 6f6f 742d 7365 7276 6572  ...c.root-server
//   	0x01f0:  7303 6e65 7400 c038 0001 0001 0000 0e10  s.net..8........
//   	0x0200:  0004 ca0c 1b21 c058 0001 0001 0000 0e10  .....!.X........
//   	0x0210:  0004 c629 0004 c078 0001 0001 0000 0e10  ...)...x........
//   	0x0220:  0004 c024 9411                           ...$..
var testMalformedRootQuery = []byte{
	0x00, 0x55, 0x22, 0xaf, 0xc6, 0x37, 0x00, 0x22, 0x55, 0xac, 0xde, 0xac, 0x08, 0x00, 0x45, 0x00,
	0x02, 0x18, 0x76, 0xb2, 0x40, 0x00, 0x72, 0x11, 0x7a, 0xd2, 0x0a, 0x4d, 0x00, 0x1a, 0x0a, 0x01,
	0x00, 0xe9, 0x00, 0x35, 0xfe, 0x2f, 0x02, 0x04, 0xb8, 0xf5, 0xe8, 0x74, 0x81, 0x00, 0x00, 0x01,
	0x00, 0x00, 0x00, 0x0d, 0x00, 0x03, 0x0c, 0x61, 0x78, 0x6b, 0x76, 0x63, 0x68, 0x63, 0x70, 0x63,
	0x70, 0x73, 0x6c, 0x0a, 0x78, 0x78, 0x78, 0x78, 0x78, 0x78, 0x78, 0x78, 0x78, 0x78, 0x03, 0x6e,
	0x65, 0x74, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x6d, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x61, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x69, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x62, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x6c, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x66, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x67, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x64, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x68, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x65, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x6a, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x6b, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x2d, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x14, 0x01, 0x63, 0x0c, 0x72, 0x6f, 0x6f, 0x74, 0x2d, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x03, 0x6e, 0x65, 0x74, 0x00, 0xc0, 0x38, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x04, 0xca, 0x0c, 0x1b, 0x21, 0xc0, 0x58, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x04, 0xc6, 0x29, 0x00, 0x04, 0xc0, 0x78, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10,
	0x00, 0x04, 0xc0, 0x24, 0x94, 0x11,
}

func TestMalformedRootQuery(t *testing.T) {
	p := gopacket.NewPacket(testMalformedRootQuery, LinkTypeEthernet, testDecodeOptions)
	if errLayer := p.ErrorLayer(); errLayer == nil {
		t.Error("No error layer on invalid DNS name")
	} else if err := errLayer.Error(); !strings.Contains(err.Error(), "no dns data found") {
		t.Errorf("unexpected error message: %v", err)
	}
}
