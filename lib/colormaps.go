// Copyright (c) 2023 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the ISC license

package lib

// there is a slight mismatch between the 16 basic IRC colors and the ANSI colors;
// this is Dan's mapping that works nicely with mainstream terminal color schemes
var ircColorToAnsiForeground = [16]uint8{
	97, // 00 white -> white, high intensity
	30, // 01 black
	34, // 02 blue
	32, // 03 green
	91, // 04 red -> red, high intensity
	31, // 05 brown -> red, normal intensity
	35, // 06 magenta
	33, // 07 orange -> yellow, normal intensity
	93, // 08 yellow -> yellow, high intensity
	92, // 09 light green -> green, high intensity
	36, // 10 cyan
	96, // 11 light cyan -> cyan, high intensity
	94, // 12 light blue -> blue, high intensity
	95, // 13 pink -> magenta, high intensity
	90, // 14 gray -> black, high intensity
	37, // 15 light gray -> white, normal intensity
}

var ircColorToAnsi256 = map[uint8]uint8{
	// overrides for the 16-color palette
	5: 94,  // brown
	7: 208, // orange
	// i considered mapping red to a "true red" but not sure how that would
	// interact with unusual remappings of the 16 colors, such as Solarized
	// 4: 196,

	// https://modern.ircdocs.horse/formatting.html#colors-16-98
	16: 52,
	17: 94,
	18: 100,
	19: 58,
	20: 22,
	21: 29,
	22: 23,
	23: 24,
	24: 17,
	25: 54,
	26: 53,
	27: 89,
	28: 88,
	29: 130,
	30: 142,
	31: 64,
	32: 28,
	33: 35,
	34: 30,
	35: 25,
	36: 18,
	37: 91,
	38: 90,
	39: 125,
	40: 124,
	41: 166,
	42: 184,
	43: 106,
	44: 34,
	45: 49,
	46: 37,
	47: 33,
	48: 19,
	49: 129,
	50: 127,
	51: 161,
	52: 196,
	53: 208,
	54: 226,
	55: 154,
	56: 46,
	57: 86,
	58: 51,
	59: 75,
	60: 21,
	61: 171,
	62: 201,
	63: 198,
	64: 203,
	65: 215,
	66: 227,
	67: 191,
	68: 83,
	69: 122,
	70: 87,
	71: 111,
	72: 63,
	73: 177,
	74: 207,
	75: 205,
	76: 217,
	77: 223,
	78: 229,
	79: 193,
	80: 157,
	81: 158,
	82: 159,
	83: 153,
	84: 147,
	85: 183,
	86: 219,
	87: 212,
	88: 16,
	89: 233,
	90: 235,
	91: 237,
	92: 239,
	93: 241,
	94: 244,
	95: 247,
	96: 250,
	97: 254,
	98: 231,
}
