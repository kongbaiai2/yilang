package main

import (
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	device       string = "en0"
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
	buffer       gopacket.SerializeBuffer
	options      gopacket.SerializeOptions
)

func main() {
	// Open device

	// sendPacketsUDPTest()
	sendPacketsTCPTest()
}

func sendPacketsUDPTest() {
	// UDP报文，比TCP简单，对端能收到
	//先写出包结构
	eth := layers.Ethernet{
		SrcMAC: net.HardwareAddr{0xC4, 0xB3, 0x01, 0xAB, 0x6E, 0x22}, //c4:b3:01:ab:6e:22
		DstMAC: net.HardwareAddr{0x0C, 0x4B, 0x54, 0x95, 0xB6, 0xB2},

		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    net.IP{192, 168, 0, 125},
		DstIP:    net.IP{183, 232, 144, 150},
		Version:  4,
		TTL:      255,
		Protocol: layers.IPProtocolUDP,
	}
	udp := layers.UDP{
		SrcPort:  layers.UDPPort(43210),
		DstPort:  layers.UDPPort(25),
		Checksum: 0,
	}
	//这个一定要加上，不管是基于tcp的还是udp的
	udp.SetNetworkLayerForChecksum(&ip4)

	//有的时候应用层协议gopacket并不是全都有，所以gopacket支持自定义数据包，但是假如说要用到的应用层包
	//比较多的时候还是比较麻烦 会写很多的代码，同时gopacket可以发送原始二进制数据，这种使用起来就比较方便
	//了，不需要你进行自定义新的数据包，只需要你把原始数据放在这里就行了，比如：
	rawBytes := []byte{
		0x16, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x36,
		0x01, 0x00, 0x00, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a,
		0xfe, 0xfd,
		0x00, 0x00, 0x00, 0x00, 0x7c, 0x77, 0x40, 0x1e, 0x8a, 0xc8, 0x22, 0xa0, 0xa0, 0x18, 0xff, 0x93,
		0x08, 0xca, 0xac, 0x0a, 0x64, 0x2f, 0xc9, 0x22, 0x64, 0xbc, 0x08, 0xa8, 0x16, 0x89, 0x19, 0x3f,
		0x00, 0x00,
		0x00, 0x02, 0x00, 0x2f,
		0x01, 0x00,
	}
	//参数ComputeChecksums，如果为true的话会自动帮你计算检验和
	serializeOptions := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true, //是否在序列化阶段要重新计算检验和
	}

	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, serializeOptions, &eth, &ip4, &udp, gopacket.Payload(rawBytes)); err != nil {
		log.Fatal(err)
	}
	//发送数据
	handle.WritePacketData(buf.Bytes())
}

func sendPacketsTCPTest() {
	// 测试能发到对端，并能回报文。构建tcp时，要按实际报文加参数。
	// 如Flags [SEW] 参数为ECE: true, CWR: true, SYN: true,
	//先写出包结构
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	rawBytes := []byte{
		0x16, 0xfe, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x36,
		0x01, 0x00, 0x00, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x2a,
		0xfe, 0xfd,
		0x00, 0x00, 0x00, 0x00, 0x7c, 0x77, 0x40, 0x1e, 0x8a, 0xc8, 0x22, 0xa0, 0xa0, 0x18, 0xff, 0x93,
		0x08, 0xca, 0xac, 0x0a, 0x64, 0x2f, 0xc9, 0x22, 0x64, 0xbc, 0x08, 0xa8, 0x16, 0x89, 0x19, 0x3f,
		0x00, 0x00,
		0x00, 0x02, 0x00, 0x2f,
		0x01, 0x00,
	}

	// This time lets fill out some information
	ipLayer := &layers.IPv4{
		SrcIP:    net.IP{192, 168, 0, 125},
		DstIP:    net.IP{183, 232, 144, 150}, //183.232.144.150
		Version:  4,
		TTL:      255,
		Protocol: layers.IPProtocolTCP,
		Flags:    layers.IPv4DontFragment,
	}
	ethernetLayer := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0xC4, 0xB3, 0x01, 0xAB, 0x6E, 0x22}, //c4:b3:01:ab:6e:22
		DstMAC:       net.HardwareAddr{0x0C, 0x4B, 0x54, 0x95, 0xB6, 0xB2}, //0c:4b:54:95:b6:b2
		EthernetType: layers.EthernetTypeIPv4,
	}
	tcpLayer := &layers.TCP{
		SrcPort:  layers.TCPPort(4321),
		DstPort:  layers.TCPPort(25),
		Checksum: 0,
		// FIN:      true,
		Seq:     1242974465,
		Ack:     1547000864,
		Options: []layers.TCPOption{},
		ECE:     true,
		CWR:     true,
		SYN:     true,
		//[nop,nop,TS val 287787641 ecr 1515364428],
	}
	//参数ComputeChecksums，如果为true的话会自动帮你计算检验和
	options = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true, //是否在序列化阶段要重新计算检验和
	}

	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	// And create the packet with the layers
	buffer := gopacket.NewSerializeBuffer()
	gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		ipLayer,
		tcpLayer,
		gopacket.Payload(rawBytes),
	)
	// outgoingPacket = buffer.Bytes()

	// Send our packet
	err = handle.WritePacketData(buffer.Bytes())
	if err != nil {
		log.Fatal("write2", err)
	}

}

/*
sudo tcpdump -vvnni en0 host 183.232.144.150
Password:
tcpdump: listening on en0, link-type EN10MB (Ethernet), capture size 262144 bytes


23:23:13.084084 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 64)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [SEW], cksum 0x1e0d (correct), seq 2676501422, win 65535, options [mss 1460,nop,wscale 6,nop,nop,TS val 289106595 ecr 0,sackOK,eol], length 0
23:23:13.139973 IP (tos 0x0, ttl 53, id 0, offset 0, flags [DF], proto TCP (6), length 60)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [S.E], cksum 0xc740 (correct), seq 168736458, ack 2676501423, win 28960, options [mss 1412,sackOK,TS val 1516755982 ecr 289106595,nop,wscale 7], length 0
23:23:13.140063 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x5efd (correct), seq 1, ack 1, win 2056, options [nop,nop,TS val 289106651 ecr 1516755982], length 0
23:23:13.140695 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 73)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x9d37 (correct), seq 1:22, ack 1, win 2056, options [nop,nop,TS val 289106651 ecr 1516755982], length 21
23:23:13.184351 IP (tos 0x0, ttl 53, id 27376, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x65e0 (correct), seq 1, ack 22, win 227, options [nop,nop,TS val 1516756027 ecr 289106651], length 0
23:23:13.197566 IP (tos 0x2,ECT(0), ttl 53, id 27377, offset 0, flags [DF], proto TCP (6), length 73)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0xa20d (correct), seq 1:22, ack 22, win 227, options [nop,nop,TS val 1516756040 ecr 289106651], length 21
23:23:13.197652 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x5e61 (correct), seq 22, ack 22, win 2055, options [nop,nop,TS val 289106708 ecr 1516756040], length 0
23:23:13.199110 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 1444)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x1cc5 (correct), seq 22:1414, ack 22, win 2055, options [nop,nop,TS val 289106709 ecr 1516756040], length 1392
23:23:13.242256 IP (tos 0x2,ECT(0), ttl 53, id 27378, offset 0, flags [DF], proto TCP (6), length 1332)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x45cd (correct), seq 22:1302, ack 22, win 227, options [nop,nop,TS val 1516756084 ecr 289106708], length 1280
23:23:13.242370 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x53ad (correct), seq 1414, ack 1302, win 2035, options [nop,nop,TS val 289106752 ecr 1516756084], length 0
23:23:13.283512 IP (tos 0x0, ttl 53, id 27379, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x5aa8 (correct), seq 1302, ack 1414, win 249, options [nop,nop,TS val 1516756126 ecr 289106709], length 0
23:23:13.283584 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 100)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x5a21 (correct), seq 1414:1462, ack 1302, win 2048, options [nop,nop,TS val 289106792 ecr 1516756126], length 48
23:23:13.328978 IP (tos 0x0, ttl 53, id 27380, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x59f8 (correct), seq 1302, ack 1462, win 249, options [nop,nop,TS val 1516756171 ecr 289106792], length 0
23:23:13.340828 IP (tos 0x2,ECT(0), ttl 53, id 27381, offset 0, flags [DF], proto TCP (6), length 416)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x9d4d (correct), seq 1302:1666, ack 1462, win 249, options [nop,nop,TS val 1516756183 ecr 289106792], length 364
23:23:13.343184 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x5147 (correct), seq 1462, ack 1666, win 2042, options [nop,nop,TS val 289106848 ecr 1516756183], length 0
23:23:13.350004 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 68)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x46ff (correct), seq 1462:1478, ack 1666, win 2048, options [nop,nop,TS val 289106857 ecr 1516756183], length 16
23:23:13.436421 IP (tos 0x0, ttl 53, id 27382, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x57d1 (correct), seq 1666, ack 1478, win 249, options [nop,nop,TS val 1516756277 ecr 289106857], length 0
23:23:13.436507 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 96)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x4859 (correct), seq 1478:1522, ack 1666, win 2048, options [nop,nop,TS val 289106943 ecr 1516756277], length 44
23:23:13.480242 IP (tos 0x0, ttl 53, id 27383, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x5721 (correct), seq 1666, ack 1522, win 249, options [nop,nop,TS val 1516756323 ecr 289106943], length 0
23:23:13.480700 IP (tos 0x2,ECT(0), ttl 53, id 27384, offset 0, flags [DF], proto TCP (6), length 96)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x0a37 (correct), seq 1666:1710, ack 1522, win 249, options [nop,nop,TS val 1516756323 ecr 289106943], length 44
23:23:13.480774 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x4fc4 (correct), seq 1522, ack 1710, win 2047, options [nop,nop,TS val 289106986 ecr 1516756323], length 0
23:23:13.480902 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 112)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x05ac (correct), seq 1522:1582, ack 1710, win 2048, options [nop,nop,TS val 289106986 ecr 1516756323], length 60
23:23:13.567939 IP (tos 0x0, ttl 53, id 27385, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x563a (correct), seq 1710, ack 1582, win 249, options [nop,nop,TS val 1516756407 ecr 289106986], length 0
23:23:13.571084 IP (tos 0x2,ECT(0), ttl 53, id 27386, offset 0, flags [DF], proto TCP (6), length 136)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0xec51 (correct), seq 1710:1794, ack 1582, win 249, options [nop,nop,TS val 1516756413 ecr 289106986], length 84
23:23:13.571156 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x4e81 (correct), seq 1582, ack 1794, win 2046, options [nop,nop,TS val 289107076 ecr 1516756413], length 0
23:23:13.571359 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 424)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x00c1 (correct), seq 1582:1954, ack 1794, win 2048, options [nop,nop,TS val 289107076 ecr 1516756413], length 372
23:23:13.615168 IP (tos 0x0, ttl 53, id 27387, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x53d0 (correct), seq 1794, ack 1954, win 271, options [nop,nop,TS val 1516756457 ecr 289107076], length 0
23:23:13.616820 IP (tos 0x2,ECT(0), ttl 53, id 27388, offset 0, flags [DF], proto TCP (6), length 384)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0xbc01 (correct), seq 1794:2126, ack 1954, win 271, options [nop,nop,TS val 1516756459 ecr 289107076], length 332
23:23:13.616896 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x4b6a (correct), seq 1954, ack 2126, win 2042, options [nop,nop,TS val 289107121 ecr 1516756459], length 0
23:23:13.621212 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 704)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0xeafa (correct), seq 1954:2606, ack 2126, win 2048, options [nop,nop,TS val 289107125 ecr 1516756459], length 652
23:23:13.672911 IP (tos 0x2,ECT(0), ttl 53, id 27389, offset 0, flags [DF], proto TCP (6), length 80)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x1d90 (correct), seq 2126:2154, ack 2606, win 293, options [nop,nop,TS val 1516756515 ecr 289107125], length 28
23:23:13.672994 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x484e (correct), seq 2606, ack 2154, win 2047, options [nop,nop,TS val 289107176 ecr 1516756515], length 0
23:23:13.673337 IP (tos 0x2,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 164)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0xb6dc (correct), seq 2606:2718, ack 2154, win 2048, options [nop,nop,TS val 289107176 ecr 1516756515], length 112
23:23:13.757173 IP (tos 0x0, ttl 53, id 27390, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x4e63 (correct), seq 2154, ack 2718, win 293, options [nop,nop,TS val 1516756600 ecr 289107176], length 0
23:23:14.565213 IP (tos 0x2,ECT(0), ttl 53, id 27391, offset 0, flags [DF], proto TCP (6), length 552)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x5584 (correct), seq 2154:2654, ack 2718, win 293, options [nop,nop,TS val 1516757307 ecr 289107176], length 500
23:23:14.565302 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3f60 (correct), seq 2718, ack 2654, win 2040, options [nop,nop,TS val 289108065 ecr 1516757307], length 0
23:23:14.609477 IP (tos 0x2,ECT(0), ttl 53, id 27392, offset 0, flags [DF], proto TCP (6), length 96)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x0fa8 (correct), seq 2654:2698, ack 2718, win 293, options [nop,nop,TS val 1516757452 ecr 289108065], length 44
23:23:14.609548 IP (tos 0x0, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3e70 (correct), seq 2718, ack 2698, win 2047, options [nop,nop,TS val 289108109 ecr 1516757452], length 0
23:23:14.609994 IP (tos 0x4a,ECT(0), ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 504)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [P.], cksum 0x3494 (correct), seq 2718:3170, ack 2698, win 2048, options [nop,nop,TS val 289108109 ecr 1516757452], length 452
23:23:14.653632 IP (tos 0x0, ttl 53, id 27393, offset 0, flags [DF], proto TCP (6), length 52)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [.], cksum 0x4344 (correct), seq 2698, ack 3170, win 315, options [nop,nop,TS val 1516757496 ecr 289108109], length 0
23:23:14.664299 IP (tos 0x2,ECT(0), ttl 53, id 27394, offset 0, flags [DF], proto TCP (6), length 160)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x3f74 (correct), seq 2698:2806, ack 3170, win 315, options [nop,nop,TS val 1516757505 ecr 289108109], length 108
23:23:14.664394 IP (tos 0x48, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3bd7 (correct), seq 3170, ack 2806, win 2046, options [nop,nop,TS val 289108162 ecr 1516757505], length 0
23:23:14.671856 IP (tos 0x2,ECT(0), ttl 53, id 27395, offset 0, flags [DF], proto TCP (6), length 296)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x2fb1 (correct), seq 2806:3050, ack 3170, win 315, options [nop,nop,TS val 1516757513 ecr 289108109], length 244
23:23:14.671952 IP (tos 0x48, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3ad6 (correct), seq 3170, ack 3050, win 2044, options [nop,nop,TS val 289108169 ecr 1516757513], length 0
23:23:14.705868 IP (tos 0x2,ECT(0), ttl 53, id 27396, offset 0, flags [DF], proto TCP (6), length 112)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x459f (correct), seq 3050:3110, ack 3170, win 315, options [nop,nop,TS val 1516757549 ecr 289108109], length 60
23:23:14.705967 IP (tos 0x48, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3a51 (correct), seq 3170, ack 3110, win 2047, options [nop,nop,TS val 289108203 ecr 1516757549], length 0
23:23:14.707298 IP (tos 0x2,ECT(0), ttl 53, id 27397, offset 0, flags [DF], proto TCP (6), length 96)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0x7bbc (correct), seq 3110:3154, ack 3170, win 315, options [nop,nop,TS val 1516757549 ecr 289108109], length 44
23:23:14.707304 IP (tos 0x2,ECT(0), ttl 53, id 27398, offset 0, flags [DF], proto TCP (6), length 112)
    183.232.144.150.22 > 192.168.0.125.51039: Flags [P.], cksum 0xac0f (correct), seq 3154:3214, ack 3170, win 315, options [nop,nop,TS val 1516757549 ecr 289108109], length 60
23:23:14.707431 IP (tos 0x48, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x3a24 (correct), seq 3170, ack 3154, win 2047, options [nop,nop,TS val 289108204 ecr 1516757549], length 0
23:23:14.707431 IP (tos 0x48, ttl 64, id 0, offset 0, flags [DF], proto TCP (6), length 52)
    192.168.0.125.51039 > 183.232.144.150.22: Flags [.], cksum 0x39e9 (correct), seq 3170, ack 3214, win 2046, options [nop,nop,TS val 289108204 ecr 1516757549], length 0
*/

/*
构造的报文回报如下
IP (tos 0x74, ttl 242, id 0, offset 0, flags [DF], proto TCP (6), length 43)
    111.194.219.50.35984 > 183.232.144.150.25: Flags [SEW], cksum 0x1b81 (correct), seq 1242974465:1242974468, win 0, length 3: SMTP, length: 3

IP (tos 0x74, ttl 66, id 4390, offset 0, flags [DF], proto TCP (6), length 40)
    183.232.144.150.25 > 111.194.219.50.35984: Flags [R.], cksum 0xfc97 (correct), seq 0, ack 1242974469, win 0, length 0
*/
