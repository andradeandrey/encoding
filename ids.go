package ebml

var (
	EBMLID               = []byte{0x1a, 0x45, 0xdf, 0xa3}
	EBMLVersionID        = []byte{0x42, 0x86}
	EBMLReadVersionID    = []byte{0x42, 0xf7}
	EBMLMaxIDLengthID    = []byte{0x42, 0xf2}
	EBMLMaxSizeLengthID  = []byte{0x42, 0xf3}
	DocTypeID            = []byte{0x42, 0x82}
	DocTypeVersionID     = []byte{0x42, 0x87}
	DocTypeReadVersionID = []byte{0x42, 0x85}

	CRC32ID              = []byte{0xbf}
	VoidID               = []byte{0xec}
	SignatureSlotID      = []byte{0x1b, 0x53, 0x86, 0x67}
	SignatureAlgoID      = []byte{0x7e, 0x8a}
	SignatureHashID      = []byte{0x7e, 0x9a}
	SignaturePublicKey   = []byte{0x7e, 0xa5}
	Signature            = []byte{0x7e, 0xb5}
	SignatureElements    = []byte{0x7e, 0x5b}
	SignatureElementList = []byte{0x7e, 0x7b}
	SignedElemnt         = []byte{0x65, 0x32}
)

const (
	EBMLIDUint                = 0x1a45dfa3
	EBMLVersionIDUint         = 0x4286
	EBMLReadVersionIDUint     = 0x42f7
	EBMLMaxIDUintLengthIDUint = 0x42f2
	EBMLMaxSizeLengthIDUint   = 0x42f3
	DocTypeIDUint             = 0x4282
	DocTypeVersionIDUint      = 0x4287
	DocTypeReadVersionIDUint  = 0x4285

	CRC32IDUint              = 0xbf
	VoidIDUint               = 0xec
	SignatureSlotIDUint      = 0x1b538667
	SignatureAlgoIDUint      = 0x7e8a
	SignatureHashIDUint      = 0x7e9a
	SignaturePublicKeyUint   = 0x7ea5
	SignatureUint            = 0x7eb5
	SignatureElementsUint    = 0x7e5b
	SignatureElementListUint = 0x7e7b
	SignedElemntUint         = 0x6532
)
