// Copyright 2012 Google, Inc. All rights reserved.
// Copyright 2009-2011 Andreas Krennmair. All rights reserved.
//
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file in the root of the source
// tree.

package layers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/google/gopacket"
)

// Based on RFC 4861

type ICMPv6Opt uint8

const (
	_ ICMPv6Opt = iota
	ICMPv6OptSourceAddress
	ICMPv6OptTargetAddress
	ICMPv6OptPrefixInfo
	ICMPv6OptRedirectedHeader
	ICMPv6OptMTU
)

type ICMPv6RouterSolicitation struct {
	BaseLayer
	Options ICMPv6Options
}

type ICMPv6RouterAdvertisement struct {
	BaseLayer
	HopLimit       uint8
	Flags          uint8
	RouterLifetime uint16
	ReachableTime  uint32
	RetransTimer   uint32
	Options        ICMPv6Options
}

type ICMPv6NeighborSolicitation struct {
	BaseLayer
	TargetAddress net.IP
	Options       ICMPv6Options
}

type ICMPv6NeighborAdvertisement struct {
	BaseLayer
	Flags         uint8
	TargetAddress net.IP
	Options       ICMPv6Options
}

type ICMPv6Redirect struct {
	BaseLayer
	TargetAddress      net.IP
	DestinationAddress net.IP
	Options            ICMPv6Options
}

type ICMPv6Option struct {
	Type   ICMPv6Opt
	Length int // Length of option, in bytes
	Data   []byte
}

type ICMPv6Options []ICMPv6Option

func (i *ICMPv6RouterSolicitation) LayerType() gopacket.LayerType {
	return LayerTypeICMPv6RouterSolicitation
}

func (i *ICMPv6RouterSolicitation) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (i *ICMPv6RouterSolicitation) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	// first 4 bytes are reserved followed by options
	if len(data) < 4 {
		df.SetTruncated()
		return errors.New("ICMP layer less then 4 bytes for ICMPv6 router solicitation")
	}

	// truncate old options
	i.Options = i.Options[:0]

	return i.Options.DecodeFromBytes(data[4:], df)
}

func (i *ICMPv6RouterSolicitation) SerializeTo(bytes []byte) {
	// TODO
}

func (i *ICMPv6RouterAdvertisement) LayerType() gopacket.LayerType {
	return LayerTypeICMPv6RouterAdvertisement
}

func (i *ICMPv6RouterAdvertisement) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (i *ICMPv6RouterAdvertisement) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 12 {
		df.SetTruncated()
		return errors.New("ICMP layer less then 12 bytes for ICMPv6 router advertisement")
	}

	i.HopLimit = uint8(data[0])
	// M, O bit followed by 6 reserved bits
	i.Flags = uint8(data[1])
	i.RouterLifetime = binary.BigEndian.Uint16(data[2:4])
	i.ReachableTime = binary.BigEndian.Uint32(data[4:8])
	i.RetransTimer = binary.BigEndian.Uint32(data[8:12])

	// truncate old options
	i.Options = i.Options[:0]

	return i.Options.DecodeFromBytes(data[12:], df)
}

func (i *ICMPv6RouterAdvertisement) SerializeTo(bytes []byte) {
	// TODO
}

func (i *ICMPv6RouterAdvertisement) ManagedAddressConfig() bool {
	return i.Flags&0x80 != 1
}

func (i *ICMPv6RouterAdvertisement) OtherConfig() bool {
	return i.Flags&0x40 != 1
}

func (i *ICMPv6NeighborSolicitation) LayerType() gopacket.LayerType {
	return LayerTypeICMPv6NeighborSolicitation
}

func (i *ICMPv6NeighborSolicitation) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (i *ICMPv6NeighborSolicitation) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 20 {
		df.SetTruncated()
		return errors.New("ICMP layer less then 20 bytes for ICMPv6 neighbor solicitation")
	}

	i.TargetAddress = net.IP(data[4:20])

	// truncate old options
	i.Options = i.Options[:0]

	return i.Options.DecodeFromBytes(data[20:], df)
}

func (i *ICMPv6NeighborSolicitation) SerializeTo(bytes []byte) {
	// TODO
}

func (i *ICMPv6NeighborAdvertisement) LayerType() gopacket.LayerType {
	return LayerTypeICMPv6NeighborAdvertisement
}

func (i *ICMPv6NeighborAdvertisement) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (i *ICMPv6NeighborAdvertisement) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 20 {
		df.SetTruncated()
		return errors.New("ICMP layer less then 20 bytes for ICMPv6 neighbor advertisement")
	}

	i.Flags = uint8(data[0])
	i.TargetAddress = net.IP(data[4:20])

	// truncate old options
	i.Options = i.Options[:0]

	return i.Options.DecodeFromBytes(data[20:], df)
}

func (i *ICMPv6NeighborAdvertisement) SerializeTo(bytes []byte) {
	// TODO
}

func (i *ICMPv6NeighborAdvertisement) Router() bool {
	return i.Flags&0x80 != 0
}

func (i *ICMPv6NeighborAdvertisement) Solicited() bool {
	return i.Flags&0x40 != 0
}

func (i *ICMPv6NeighborAdvertisement) Override() bool {
	return i.Flags&0x20 != 0
}

func (i *ICMPv6Redirect) LayerType() gopacket.LayerType {
	return LayerTypeICMPv6Redirect
}

func (i *ICMPv6Redirect) NextLayerType() gopacket.LayerType {
	return gopacket.LayerTypePayload
}

func (i *ICMPv6Redirect) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	if len(data) < 36 {
		df.SetTruncated()
		return errors.New("ICMP layer less then 36 bytes for ICMPv6 redirect")
	}

	i.TargetAddress = net.IP(data[4:20])
	i.DestinationAddress = net.IP(data[20:36])

	// truncate old options
	i.Options = i.Options[:0]

	return i.Options.DecodeFromBytes(data[36:], df)
}

func (i *ICMPv6Redirect) SerializeTo(bytes []byte) {
	// TODO
}

func (i *ICMPv6Options) DecodeFromBytes(data []byte, df gopacket.DecodeFeedback) error {
	for len(data) > 0 {
		if len(data) < 2 {
			df.SetTruncated()
			return errors.New("ICMP layer less then 2 bytes for ICMPv6 message option")
		}

		o := ICMPv6Option{
			Type: ICMPv6Opt(data[0]),
			// unit of Length is 8 octets, convert to bytes
			Length: int(data[1]) * 8,
		}

		if len(data) < int(o.Length) {
			df.SetTruncated()
			return fmt.Errorf("ICMP layer only %v bytes for ICMPv6 message option with length %v", len(data), o.Length)
		}

		o.Data = data[2:o.Length]
		// chop off option we just consumed
		data = data[o.Length:]

		*i = append(*i, o)
	}

	return nil
}

func (i *ICMPv6Options) SerializeTo(bytes []byte) {
	// TODO
}

func decodeICMPv6RouterSolicitation(data []byte, p gopacket.PacketBuilder) error {
	i := &ICMPv6RouterSolicitation{}
	return decodingLayerDecoder(i, data, p)
}

func decodeICMPv6RouterAdvertisement(data []byte, p gopacket.PacketBuilder) error {
	i := &ICMPv6RouterAdvertisement{}
	return decodingLayerDecoder(i, data, p)
}

func decodeICMPv6NeighborSolicitation(data []byte, p gopacket.PacketBuilder) error {
	i := &ICMPv6NeighborSolicitation{}
	return decodingLayerDecoder(i, data, p)
}

func decodeICMPv6NeighborAdvertisement(data []byte, p gopacket.PacketBuilder) error {
	i := &ICMPv6NeighborAdvertisement{}
	return decodingLayerDecoder(i, data, p)
}

func decodeICMPv6Redirect(data []byte, p gopacket.PacketBuilder) error {
	i := &ICMPv6Redirect{}
	return decodingLayerDecoder(i, data, p)
}
