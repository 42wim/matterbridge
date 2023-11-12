package qrcode

const (
	_VERSION_COUNT       = 40  // (40 versions)
	_VERSIONS_ITEM_COUNT = 160 // (40 versions x 4 error correction level)
)

// versions contains information about each QR Code version.
// NOTICE: item in version array MUST keep sorted according to
// QR Version sequential as the first key and Error Correction Level as the second (ASC).
var versions = []version{
	{
		Ver:     1,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      41,
			AlphaNumeric: 25,
			Byte:         17,
			JP:           10,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     19,
				ECBlockwordsPerBlock: 7,
			},
		},
	},
	{
		Ver:     1,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      34,
			AlphaNumeric: 20,
			Byte:         14,
			JP:           8,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 10,
			},
		},
	},
	{
		Ver:     1,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      27,
			AlphaNumeric: 16,
			Byte:         11,
			JP:           7,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 13,
			},
		},
	},
	{
		Ver:     1,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      17,
			AlphaNumeric: 10,
			Byte:         7,
			JP:           4,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     9,
				ECBlockwordsPerBlock: 17,
			},
		},
	},
	{
		Ver:     2,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      77,
			AlphaNumeric: 47,
			Byte:         32,
			JP:           20,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     34,
				ECBlockwordsPerBlock: 10,
			},
		},
	},
	{
		Ver:     2,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      63,
			AlphaNumeric: 38,
			Byte:         26,
			JP:           16,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     28,
				ECBlockwordsPerBlock: 16,
			},
		},
	},
	{
		Ver:     2,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      48,
			AlphaNumeric: 29,
			Byte:         20,
			JP:           12,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     2,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      34,
			AlphaNumeric: 20,
			Byte:         14,
			JP:           8,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     3,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      127,
			AlphaNumeric: 77,
			Byte:         53,
			JP:           32,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     55,
				ECBlockwordsPerBlock: 15,
			},
		},
	},
	{
		Ver:     3,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      101,
			AlphaNumeric: 61,
			Byte:         42,
			JP:           26,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     44,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     3,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      77,
			AlphaNumeric: 47,
			Byte:         32,
			JP:           20,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     3,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      58,
			AlphaNumeric: 35,
			Byte:         24,
			JP:           15,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     4,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      187,
			AlphaNumeric: 114,
			Byte:         78,
			JP:           48,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     80,
				ECBlockwordsPerBlock: 20,
			},
		},
	},
	{
		Ver:     4,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      149,
			AlphaNumeric: 90,
			Byte:         62,
			JP:           38,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     32,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     4,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      111,
			AlphaNumeric: 67,
			Byte:         46,
			JP:           28,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     4,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      82,
			AlphaNumeric: 50,
			Byte:         34,
			JP:           21,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     9,
				ECBlockwordsPerBlock: 16,
			},
		},
	},
	{
		Ver:     5,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      255,
			AlphaNumeric: 154,
			Byte:         106,
			JP:           65,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     108,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     5,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      202,
			AlphaNumeric: 122,
			Byte:         84,
			JP:           52,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     43,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     5,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      144,
			AlphaNumeric: 87,
			Byte:         60,
			JP:           37,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 18,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     5,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      106,
			AlphaNumeric: 64,
			Byte:         44,
			JP:           27,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     11,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     6,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      322,
			AlphaNumeric: 195,
			Byte:         134,
			JP:           82,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     68,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     6,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      255,
			AlphaNumeric: 154,
			Byte:         106,
			JP:           65,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     27,
				ECBlockwordsPerBlock: 16,
			},
		},
	},
	{
		Ver:     6,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      178,
			AlphaNumeric: 108,
			Byte:         74,
			JP:           45,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     19,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     6,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      139,
			AlphaNumeric: 84,
			Byte:         58,
			JP:           36,
		},
		RemainderBits: 7,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     7,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      370,
			AlphaNumeric: 224,
			Byte:         154,
			JP:           95,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     78,
				ECBlockwordsPerBlock: 20,
			},
		},
	},
	{
		Ver:     7,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      293,
			AlphaNumeric: 178,
			Byte:         122,
			JP:           75,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     31,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     7,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      207,
			AlphaNumeric: 125,
			Byte:         86,
			JP:           53,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 18,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     7,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      154,
			AlphaNumeric: 93,
			Byte:         64,
			JP:           39,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     8,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      461,
			AlphaNumeric: 279,
			Byte:         192,
			JP:           118,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     97,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     8,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      365,
			AlphaNumeric: 221,
			Byte:         152,
			JP:           93,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     38,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     39,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     8,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      259,
			AlphaNumeric: 157,
			Byte:         108,
			JP:           66,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     18,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     19,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     8,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      202,
			AlphaNumeric: 122,
			Byte:         84,
			JP:           52,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     9,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      552,
			AlphaNumeric: 335,
			Byte:         230,
			JP:           141,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     9,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      432,
			AlphaNumeric: 262,
			Byte:         180,
			JP:           111,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     36,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     37,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     9,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      312,
			AlphaNumeric: 189,
			Byte:         130,
			JP:           80,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 20,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 20,
			},
		},
	},
	{
		Ver:     9,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      235,
			AlphaNumeric: 143,
			Byte:         98,
			JP:           60,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     10,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      652,
			AlphaNumeric: 395,
			Byte:         271,
			JP:           167,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     68,
				ECBlockwordsPerBlock: 18,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     69,
				ECBlockwordsPerBlock: 18,
			},
		},
	},
	{
		Ver:     10,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      513,
			AlphaNumeric: 311,
			Byte:         213,
			JP:           131,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     43,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     44,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     10,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      364,
			AlphaNumeric: 221,
			Byte:         151,
			JP:           93,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     19,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     20,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     10,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      288,
			AlphaNumeric: 174,
			Byte:         119,
			JP:           74,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     11,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      772,
			AlphaNumeric: 468,
			Byte:         321,
			JP:           198,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     81,
				ECBlockwordsPerBlock: 20,
			},
		},
	},
	{
		Ver:     11,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      604,
			AlphaNumeric: 366,
			Byte:         251,
			JP:           155,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     50,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     51,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     11,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      427,
			AlphaNumeric: 259,
			Byte:         177,
			JP:           109,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     11,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      331,
			AlphaNumeric: 200,
			Byte:         137,
			JP:           85,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            8,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     12,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      883,
			AlphaNumeric: 535,
			Byte:         367,
			JP:           226,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     92,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     93,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     12,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      691,
			AlphaNumeric: 419,
			Byte:         287,
			JP:           177,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     36,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     37,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     12,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      489,
			AlphaNumeric: 296,
			Byte:         203,
			JP:           125,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     20,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     21,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     12,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      374,
			AlphaNumeric: 227,
			Byte:         155,
			JP:           96,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            7,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     13,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1022,
			AlphaNumeric: 619,
			Byte:         425,
			JP:           262,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     107,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     13,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      796,
			AlphaNumeric: 483,
			Byte:         331,
			JP:           204,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     37,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     38,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     13,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      580,
			AlphaNumeric: 352,
			Byte:         241,
			JP:           149,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     20,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     21,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     13,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      427,
			AlphaNumeric: 259,
			Byte:         177,
			JP:           109,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            12,
				NumDataCodewords:     11,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     14,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1101,
			AlphaNumeric: 667,
			Byte:         458,
			JP:           282,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     14,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      871,
			AlphaNumeric: 528,
			Byte:         362,
			JP:           223,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     40,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     41,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     14,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      621,
			AlphaNumeric: 376,
			Byte:         258,
			JP:           159,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 20,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 20,
			},
		},
	},
	{
		Ver:     14,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      468,
			AlphaNumeric: 283,
			Byte:         194,
			JP:           120,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     15,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1250,
			AlphaNumeric: 758,
			Byte:         520,
			JP:           320,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     87,
				ECBlockwordsPerBlock: 22,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     88,
				ECBlockwordsPerBlock: 22,
			},
		},
	},
	{
		Ver:     15,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      991,
			AlphaNumeric: 600,
			Byte:         412,
			JP:           254,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     41,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     42,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     15,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      703,
			AlphaNumeric: 426,
			Byte:         292,
			JP:           180,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     15,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      530,
			AlphaNumeric: 321,
			Byte:         220,
			JP:           136,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     12,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     16,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1408,
			AlphaNumeric: 854,
			Byte:         586,
			JP:           361,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     98,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     99,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     16,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1082,
			AlphaNumeric: 656,
			Byte:         450,
			JP:           277,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            7,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            3,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     16,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      775,
			AlphaNumeric: 470,
			Byte:         322,
			JP:           198,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            15,
				NumDataCodewords:     19,
				ECBlockwordsPerBlock: 24,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     20,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     16,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      602,
			AlphaNumeric: 365,
			Byte:         250,
			JP:           154,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            13,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     17,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1548,
			AlphaNumeric: 938,
			Byte:         644,
			JP:           397,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     107,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     108,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     17,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1212,
			AlphaNumeric: 734,
			Byte:         504,
			JP:           310,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            10,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     17,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      876,
			AlphaNumeric: 531,
			Byte:         364,
			JP:           224,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            15,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     17,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      674,
			AlphaNumeric: 408,
			Byte:         280,
			JP:           173,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            17,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     18,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1725,
			AlphaNumeric: 1046,
			Byte:         718,
			JP:           442,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     120,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     121,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     18,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1346,
			AlphaNumeric: 816,
			Byte:         560,
			JP:           345,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            9,
				NumDataCodewords:     43,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     44,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     18,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      948,
			AlphaNumeric: 574,
			Byte:         394,
			JP:           243,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     18,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      746,
			AlphaNumeric: 452,
			Byte:         310,
			JP:           191,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            19,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     19,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      1903,
			AlphaNumeric: 1153,
			Byte:         792,
			JP:           488,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     113,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     114,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     19,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1500,
			AlphaNumeric: 909,
			Byte:         624,
			JP:           384,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     44,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            11,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     19,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1063,
			AlphaNumeric: 644,
			Byte:         442,
			JP:           272,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     21,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     19,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      813,
			AlphaNumeric: 493,
			Byte:         338,
			JP:           208,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            9,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            16,
				NumDataCodewords:     14,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     20,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      2061,
			AlphaNumeric: 1249,
			Byte:         858,
			JP:           528,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     107,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     108,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     20,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1600,
			AlphaNumeric: 970,
			Byte:         666,
			JP:           410,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     41,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            13,
				NumDataCodewords:     42,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     20,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1159,
			AlphaNumeric: 702,
			Byte:         482,
			JP:           297,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            15,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     20,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      919,
			AlphaNumeric: 557,
			Byte:         382,
			JP:           235,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            15,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     21,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      2232,
			AlphaNumeric: 1352,
			Byte:         929,
			JP:           572,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     117,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     21,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1708,
			AlphaNumeric: 1035,
			Byte:         711,
			JP:           438,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     42,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     21,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1224,
			AlphaNumeric: 742,
			Byte:         509,
			JP:           314,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     21,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      969,
			AlphaNumeric: 587,
			Byte:         403,
			JP:           248,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     22,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      2409,
			AlphaNumeric: 1460,
			Byte:         1003,
			JP:           618,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     111,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     112,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     22,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      1872,
			AlphaNumeric: 1134,
			Byte:         779,
			JP:           480,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     22,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1358,
			AlphaNumeric: 823,
			Byte:         565,
			JP:           348,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            7,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            16,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     22,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1056,
			AlphaNumeric: 640,
			Byte:         439,
			JP:           270,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            34,
				NumDataCodewords:     13,
				ECBlockwordsPerBlock: 24,
			},
		},
	},
	{
		Ver:     23,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      2620,
			AlphaNumeric: 1588,
			Byte:         1091,
			JP:           672,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     121,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            5,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     23,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2059,
			AlphaNumeric: 1248,
			Byte:         857,
			JP:           528,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     23,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1468,
			AlphaNumeric: 890,
			Byte:         611,
			JP:           376,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     23,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1108,
			AlphaNumeric: 672,
			Byte:         461,
			JP:           284,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            16,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     24,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      2812,
			AlphaNumeric: 1704,
			Byte:         1171,
			JP:           721,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     117,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     118,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     24,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2188,
			AlphaNumeric: 1326,
			Byte:         911,
			JP:           561,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     24,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1588,
			AlphaNumeric: 963,
			Byte:         661,
			JP:           407,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            16,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     24,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1228,
			AlphaNumeric: 744,
			Byte:         511,
			JP:           315,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            30,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     25,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      3057,
			AlphaNumeric: 1853,
			Byte:         1273,
			JP:           784,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     106,
				ECBlockwordsPerBlock: 26,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     107,
				ECBlockwordsPerBlock: 26,
			},
		},
	},
	{
		Ver:     25,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2395,
			AlphaNumeric: 1451,
			Byte:         997,
			JP:           614,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            13,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     25,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1718,
			AlphaNumeric: 1041,
			Byte:         715,
			JP:           440,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            7,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            22,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     25,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1286,
			AlphaNumeric: 779,
			Byte:         535,
			JP:           330,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            22,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            13,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     26,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      3283,
			AlphaNumeric: 1990,
			Byte:         1367,
			JP:           842,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            10,
				NumDataCodewords:     114,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            2,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     26,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2544,
			AlphaNumeric: 1542,
			Byte:         1059,
			JP:           652,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     26,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1804,
			AlphaNumeric: 1094,
			Byte:         751,
			JP:           462,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            28,
				NumDataCodewords:     22,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     26,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1425,
			AlphaNumeric: 864,
			Byte:         593,
			JP:           365,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            33,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     27,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      3517,
			AlphaNumeric: 2132,
			Byte:         1465,
			JP:           902,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     123,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     27,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2701,
			AlphaNumeric: 1637,
			Byte:         1125,
			JP:           692,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            22,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            3,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     27,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      1933,
			AlphaNumeric: 1172,
			Byte:         805,
			JP:           496,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            8,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            26,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     27,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1501,
			AlphaNumeric: 910,
			Byte:         625,
			JP:           385,
		},
		RemainderBits: 4,
		Groups: []group{
			{
				NumBlocks:            12,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            28,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     28,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      3669,
			AlphaNumeric: 2223,
			Byte:         1528,
			JP:           940,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     117,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     118,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     28,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      2857,
			AlphaNumeric: 1732,
			Byte:         1190,
			JP:           732,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            3,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            23,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     28,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2085,
			AlphaNumeric: 1263,
			Byte:         868,
			JP:           534,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            31,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     28,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1581,
			AlphaNumeric: 958,
			Byte:         658,
			JP:           405,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            31,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     29,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      3909,
			AlphaNumeric: 2369,
			Byte:         1628,
			JP:           1002,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            7,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     117,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     29,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      3035,
			AlphaNumeric: 1839,
			Byte:         1264,
			JP:           778,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            21,
				NumDataCodewords:     45,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     29,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2181,
			AlphaNumeric: 1322,
			Byte:         908,
			JP:           559,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            1,
				NumDataCodewords:     23,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            37,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     29,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1677,
			AlphaNumeric: 1016,
			Byte:         698,
			JP:           430,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            26,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     30,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      4158,
			AlphaNumeric: 2520,
			Byte:         1732,
			JP:           1066,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            5,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     30,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      3289,
			AlphaNumeric: 1994,
			Byte:         1370,
			JP:           843,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     30,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2358,
			AlphaNumeric: 1429,
			Byte:         982,
			JP:           604,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            15,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            25,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     30,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1782,
			AlphaNumeric: 1080,
			Byte:         742,
			JP:           457,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            23,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            25,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     31,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      4417,
			AlphaNumeric: 2677,
			Byte:         1840,
			JP:           1132,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            13,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            3,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     31,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      3486,
			AlphaNumeric: 2113,
			Byte:         1452,
			JP:           894,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            29,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     31,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2473,
			AlphaNumeric: 1499,
			Byte:         1030,
			JP:           634,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            42,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     31,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      1897,
			AlphaNumeric: 1150,
			Byte:         790,
			JP:           486,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            23,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            28,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     32,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      4686,
			AlphaNumeric: 2840,
			Byte:         1952,
			JP:           1201,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     32,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      3693,
			AlphaNumeric: 2238,
			Byte:         1538,
			JP:           947,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            10,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            23,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     32,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2670,
			AlphaNumeric: 1618,
			Byte:         1112,
			JP:           684,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            10,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            35,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     32,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2022,
			AlphaNumeric: 1226,
			Byte:         842,
			JP:           518,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            35,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     33,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      4965,
			AlphaNumeric: 3009,
			Byte:         2068,
			JP:           1273,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     33,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      3909,
			AlphaNumeric: 2369,
			Byte:         1628,
			JP:           1002,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            14,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            21,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     33,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2805,
			AlphaNumeric: 1700,
			Byte:         1168,
			JP:           719,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            29,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            19,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     33,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2157,
			AlphaNumeric: 1307,
			Byte:         898,
			JP:           553,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            11,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            46,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     34,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      5253,
			AlphaNumeric: 3183,
			Byte:         2188,
			JP:           1347,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            13,
				NumDataCodewords:     115,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     116,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     34,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      4134,
			AlphaNumeric: 2506,
			Byte:         1722,
			JP:           1060,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            14,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            23,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     34,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      2949,
			AlphaNumeric: 1787,
			Byte:         1228,
			JP:           756,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            44,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     34,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2301,
			AlphaNumeric: 1394,
			Byte:         958,
			JP:           590,
		},
		RemainderBits: 3,
		Groups: []group{
			{
				NumBlocks:            59,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            1,
				NumDataCodewords:     17,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     35,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      5529,
			AlphaNumeric: 3351,
			Byte:         2303,
			JP:           1417,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            12,
				NumDataCodewords:     121,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     35,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      4343,
			AlphaNumeric: 2632,
			Byte:         1809,
			JP:           1113,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            12,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            26,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     35,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3081,
			AlphaNumeric: 1867,
			Byte:         1283,
			JP:           790,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            39,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     35,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2361,
			AlphaNumeric: 1431,
			Byte:         983,
			JP:           605,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            22,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            41,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     36,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      5836,
			AlphaNumeric: 3537,
			Byte:         2431,
			JP:           1496,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     121,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     36,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      4588,
			AlphaNumeric: 2780,
			Byte:         1911,
			JP:           1176,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            6,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            34,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     36,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3244,
			AlphaNumeric: 1966,
			Byte:         1351,
			JP:           832,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            46,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     36,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2524,
			AlphaNumeric: 1530,
			Byte:         1051,
			JP:           647,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            2,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            64,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     37,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      6153,
			AlphaNumeric: 3729,
			Byte:         2563,
			JP:           1577,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            17,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     123,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     37,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      4775,
			AlphaNumeric: 2894,
			Byte:         1989,
			JP:           1224,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            29,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     37,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3417,
			AlphaNumeric: 2071,
			Byte:         1423,
			JP:           876,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            49,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            10,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     37,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2625,
			AlphaNumeric: 1591,
			Byte:         1093,
			JP:           673,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            24,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            46,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     38,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      6479,
			AlphaNumeric: 3927,
			Byte:         2699,
			JP:           1661,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            4,
				NumDataCodewords:     122,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            18,
				NumDataCodewords:     123,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     38,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      5039,
			AlphaNumeric: 3054,
			Byte:         2099,
			JP:           1292,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            13,
				NumDataCodewords:     46,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            32,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     38,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3599,
			AlphaNumeric: 2181,
			Byte:         1499,
			JP:           923,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            48,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            14,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     38,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2735,
			AlphaNumeric: 1658,
			Byte:         1139,
			JP:           701,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            42,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            32,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     39,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      6743,
			AlphaNumeric: 4087,
			Byte:         2809,
			JP:           1729,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            20,
				NumDataCodewords:     117,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            4,
				NumDataCodewords:     118,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     39,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      5313,
			AlphaNumeric: 3220,
			Byte:         2213,
			JP:           1362,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            40,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            7,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     39,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3791,
			AlphaNumeric: 2298,
			Byte:         1579,
			JP:           972,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            43,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            22,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     39,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      2927,
			AlphaNumeric: 1774,
			Byte:         1219,
			JP:           750,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            10,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            67,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     40,
		ECLevel: 1,
		Cap: capacity{
			Numeric:      7089,
			AlphaNumeric: 4296,
			Byte:         2953,
			JP:           1817,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            19,
				NumDataCodewords:     118,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            6,
				NumDataCodewords:     119,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     40,
		ECLevel: 2,
		Cap: capacity{
			Numeric:      5596,
			AlphaNumeric: 3391,
			Byte:         2331,
			JP:           1435,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            18,
				NumDataCodewords:     47,
				ECBlockwordsPerBlock: 28,
			},
			{
				NumBlocks:            31,
				NumDataCodewords:     48,
				ECBlockwordsPerBlock: 28,
			},
		},
	},
	{
		Ver:     40,
		ECLevel: 3,
		Cap: capacity{
			Numeric:      3993,
			AlphaNumeric: 2420,
			Byte:         1663,
			JP:           1024,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            34,
				NumDataCodewords:     24,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            34,
				NumDataCodewords:     25,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
	{
		Ver:     40,
		ECLevel: 4,
		Cap: capacity{
			Numeric:      3057,
			AlphaNumeric: 1852,
			Byte:         1273,
			JP:           784,
		},
		RemainderBits: 0,
		Groups: []group{
			{
				NumBlocks:            20,
				NumDataCodewords:     15,
				ECBlockwordsPerBlock: 30,
			},
			{
				NumBlocks:            61,
				NumDataCodewords:     16,
				ECBlockwordsPerBlock: 30,
			},
		},
	},
}
