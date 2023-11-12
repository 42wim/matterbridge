package multiaddr

// You **MUST** register your multicodecs with
// https://github.com/multiformats/multicodec before adding them here.
const (
	P_IP4               = 4
	P_TCP               = 6
	P_DNS               = 53 // 4 or 6
	P_DNS4              = 54
	P_DNS6              = 55
	P_DNSADDR           = 56
	P_UDP               = 273
	P_DCCP              = 33
	P_IP6               = 41
	P_IP6ZONE           = 42
	P_IPCIDR            = 43
	P_QUIC              = 460
	P_QUIC_V1           = 461
	P_WEBTRANSPORT      = 465
	P_CERTHASH          = 466
	P_SCTP              = 132
	P_CIRCUIT           = 290
	P_UDT               = 301
	P_UTP               = 302
	P_UNIX              = 400
	P_P2P               = 421
	P_IPFS              = P_P2P // alias for backwards compatibility
	P_HTTP              = 480
	P_HTTPS             = 443 // deprecated alias for /tls/http
	P_ONION             = 444 // also for backwards compatibility
	P_ONION3            = 445
	P_GARLIC64          = 446
	P_GARLIC32          = 447
	P_P2P_WEBRTC_DIRECT = 276 // Deprecated. use webrtc-direct instead
	P_TLS               = 448
	P_SNI               = 449
	P_NOISE             = 454
	P_WS                = 477
	P_WSS               = 478 // deprecated alias for /tls/ws
	P_PLAINTEXTV2       = 7367777
	P_WEBRTC_DIRECT     = 280
	P_WEBRTC            = 281
)

var (
	protoIP4 = Protocol{
		Name:       "ip4",
		Code:       P_IP4,
		VCode:      CodeToVarint(P_IP4),
		Size:       32,
		Path:       false,
		Transcoder: TranscoderIP4,
	}
	protoTCP = Protocol{
		Name:       "tcp",
		Code:       P_TCP,
		VCode:      CodeToVarint(P_TCP),
		Size:       16,
		Path:       false,
		Transcoder: TranscoderPort,
	}
	protoDNS = Protocol{
		Code:       P_DNS,
		Size:       LengthPrefixedVarSize,
		Name:       "dns",
		VCode:      CodeToVarint(P_DNS),
		Transcoder: TranscoderDns,
	}
	protoDNS4 = Protocol{
		Code:       P_DNS4,
		Size:       LengthPrefixedVarSize,
		Name:       "dns4",
		VCode:      CodeToVarint(P_DNS4),
		Transcoder: TranscoderDns,
	}
	protoDNS6 = Protocol{
		Code:       P_DNS6,
		Size:       LengthPrefixedVarSize,
		Name:       "dns6",
		VCode:      CodeToVarint(P_DNS6),
		Transcoder: TranscoderDns,
	}
	protoDNSADDR = Protocol{
		Code:       P_DNSADDR,
		Size:       LengthPrefixedVarSize,
		Name:       "dnsaddr",
		VCode:      CodeToVarint(P_DNSADDR),
		Transcoder: TranscoderDns,
	}
	protoUDP = Protocol{
		Name:       "udp",
		Code:       P_UDP,
		VCode:      CodeToVarint(P_UDP),
		Size:       16,
		Path:       false,
		Transcoder: TranscoderPort,
	}
	protoDCCP = Protocol{
		Name:       "dccp",
		Code:       P_DCCP,
		VCode:      CodeToVarint(P_DCCP),
		Size:       16,
		Path:       false,
		Transcoder: TranscoderPort,
	}
	protoIP6 = Protocol{
		Name:       "ip6",
		Code:       P_IP6,
		VCode:      CodeToVarint(P_IP6),
		Size:       128,
		Transcoder: TranscoderIP6,
	}
	protoIPCIDR = Protocol{
		Name:       "ipcidr",
		Code:       P_IPCIDR,
		VCode:      CodeToVarint(P_IPCIDR),
		Size:       8,
		Transcoder: TranscoderIPCIDR,
	}
	// these require varint
	protoIP6ZONE = Protocol{
		Name:       "ip6zone",
		Code:       P_IP6ZONE,
		VCode:      CodeToVarint(P_IP6ZONE),
		Size:       LengthPrefixedVarSize,
		Path:       false,
		Transcoder: TranscoderIP6Zone,
	}
	protoSCTP = Protocol{
		Name:       "sctp",
		Code:       P_SCTP,
		VCode:      CodeToVarint(P_SCTP),
		Size:       16,
		Transcoder: TranscoderPort,
	}

	protoCIRCUIT = Protocol{
		Code:  P_CIRCUIT,
		Size:  0,
		Name:  "p2p-circuit",
		VCode: CodeToVarint(P_CIRCUIT),
	}

	protoONION2 = Protocol{
		Name:       "onion",
		Code:       P_ONION,
		VCode:      CodeToVarint(P_ONION),
		Size:       96,
		Transcoder: TranscoderOnion,
	}
	protoONION3 = Protocol{
		Name:       "onion3",
		Code:       P_ONION3,
		VCode:      CodeToVarint(P_ONION3),
		Size:       296,
		Transcoder: TranscoderOnion3,
	}
	protoGARLIC64 = Protocol{
		Name:       "garlic64",
		Code:       P_GARLIC64,
		VCode:      CodeToVarint(P_GARLIC64),
		Size:       LengthPrefixedVarSize,
		Transcoder: TranscoderGarlic64,
	}
	protoGARLIC32 = Protocol{
		Name:       "garlic32",
		Code:       P_GARLIC32,
		VCode:      CodeToVarint(P_GARLIC32),
		Size:       LengthPrefixedVarSize,
		Transcoder: TranscoderGarlic32,
	}
	protoUTP = Protocol{
		Name:  "utp",
		Code:  P_UTP,
		VCode: CodeToVarint(P_UTP),
	}
	protoUDT = Protocol{
		Name:  "udt",
		Code:  P_UDT,
		VCode: CodeToVarint(P_UDT),
	}
	protoQUIC = Protocol{
		Name:  "quic",
		Code:  P_QUIC,
		VCode: CodeToVarint(P_QUIC),
	}
	protoQUICV1 = Protocol{
		Name:  "quic-v1",
		Code:  P_QUIC_V1,
		VCode: CodeToVarint(P_QUIC_V1),
	}
	protoWEBTRANSPORT = Protocol{
		Name:  "webtransport",
		Code:  P_WEBTRANSPORT,
		VCode: CodeToVarint(P_WEBTRANSPORT),
	}
	protoCERTHASH = Protocol{
		Name:       "certhash",
		Code:       P_CERTHASH,
		VCode:      CodeToVarint(P_CERTHASH),
		Size:       LengthPrefixedVarSize,
		Transcoder: TranscoderCertHash,
	}
	protoHTTP = Protocol{
		Name:  "http",
		Code:  P_HTTP,
		VCode: CodeToVarint(P_HTTP),
	}
	protoHTTPS = Protocol{
		Name:  "https",
		Code:  P_HTTPS,
		VCode: CodeToVarint(P_HTTPS),
	}
	protoP2P = Protocol{
		Name:       "p2p",
		Code:       P_P2P,
		VCode:      CodeToVarint(P_P2P),
		Size:       LengthPrefixedVarSize,
		Transcoder: TranscoderP2P,
	}
	protoUNIX = Protocol{
		Name:       "unix",
		Code:       P_UNIX,
		VCode:      CodeToVarint(P_UNIX),
		Size:       LengthPrefixedVarSize,
		Path:       true,
		Transcoder: TranscoderUnix,
	}
	protoP2P_WEBRTC_DIRECT = Protocol{
		Name:  "p2p-webrtc-direct",
		Code:  P_P2P_WEBRTC_DIRECT,
		VCode: CodeToVarint(P_P2P_WEBRTC_DIRECT),
	}
	protoTLS = Protocol{
		Name:  "tls",
		Code:  P_TLS,
		VCode: CodeToVarint(P_TLS),
	}
	protoSNI = Protocol{
		Name:       "sni",
		Size:       LengthPrefixedVarSize,
		Code:       P_SNI,
		VCode:      CodeToVarint(P_SNI),
		Transcoder: TranscoderDns,
	}
	protoNOISE = Protocol{
		Name:  "noise",
		Code:  P_NOISE,
		VCode: CodeToVarint(P_NOISE),
	}
	protoPlaintextV2 = Protocol{
		Name:  "plaintextv2",
		Code:  P_PLAINTEXTV2,
		VCode: CodeToVarint(P_PLAINTEXTV2),
	}
	protoWS = Protocol{
		Name:  "ws",
		Code:  P_WS,
		VCode: CodeToVarint(P_WS),
	}
	protoWSS = Protocol{
		Name:  "wss",
		Code:  P_WSS,
		VCode: CodeToVarint(P_WSS),
	}
	protoWebRTCDirect = Protocol{
		Name:  "webrtc-direct",
		Code:  P_WEBRTC_DIRECT,
		VCode: CodeToVarint(P_WEBRTC_DIRECT),
	}
	protoWebRTC = Protocol{
		Name:  "webrtc",
		Code:  P_WEBRTC,
		VCode: CodeToVarint(P_WEBRTC),
	}
)

func init() {
	for _, p := range []Protocol{
		protoIP4,
		protoTCP,
		protoDNS,
		protoDNS4,
		protoDNS6,
		protoDNSADDR,
		protoUDP,
		protoDCCP,
		protoIP6,
		protoIP6ZONE,
		protoIPCIDR,
		protoSCTP,
		protoCIRCUIT,
		protoONION2,
		protoONION3,
		protoGARLIC64,
		protoGARLIC32,
		protoUTP,
		protoUDT,
		protoQUIC,
		protoQUICV1,
		protoWEBTRANSPORT,
		protoCERTHASH,
		protoHTTP,
		protoHTTPS,
		protoP2P,
		protoUNIX,
		protoP2P_WEBRTC_DIRECT,
		protoTLS,
		protoSNI,
		protoNOISE,
		protoWS,
		protoWSS,
		protoPlaintextV2,
		protoWebRTCDirect,
		protoWebRTC,
	} {
		if err := AddProtocol(p); err != nil {
			panic(err)
		}
	}

	// explicitly set both of these
	protocolsByName["p2p"] = protoP2P
	protocolsByName["ipfs"] = protoP2P
}
