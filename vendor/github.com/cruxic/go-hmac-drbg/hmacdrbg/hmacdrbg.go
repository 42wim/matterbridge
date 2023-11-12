package hmacdrbg

//Ported from: https://github.com/fpgaminer/python-hmac-drbg
import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

/**937 bytes (~7500 bits) as per the spec.*/
const MaxBytesPerGenerate = 937  // ~ 7500bits/8

/**Entropy for NewHmacDrbg() and Reseed() must never exceed this number of bytes.*/
const MaxEntropyBytes = 125      // = 1000bits

const MaxPersonalizationBytes = 32 // = 256bits

type HmacDrbg struct {
	/**The effective security level (eg 128 bits) which this generator was instantiated with.*/
	SecurityLevelBits int
	
	k, v []byte
	
	reseedCounter int
}

/**Read from an arbitrary number of bytes from HmacDrbg efficiently.
Internally it generates blocks of MaxBytesPerGenerate.  It then
serves these out through the standard `Read` function.  Read returns
an error if reseed becomes is necessary.
*/
type HmacDrbgReader struct {
	Drbg *HmacDrbg
	buffer []byte //size MaxBytesPerGenerate
	offset int
}

/**Create a new DRBG.
desiredSecurityLevelBits must be one of 112, 128, 192, 256.

entropy length (in bits) must be at least 1.5 times securityLevelBits.
entropy byte length cannot exceed MaxEntropyBytes.

The personalization can be nil.  If non-nil, it's byte length cannot exceed MaxPersonalizationBytes.

If any of the parameters are out-of-range this function will panic.
*/
func NewHmacDrbg(securityLevelBits int, entropy, personalization []byte) *HmacDrbg {
	if securityLevelBits != 112 && 
		securityLevelBits != 128 &&
		securityLevelBits != 192 &&
		securityLevelBits != 256 {
		
		panic("Illegal desiredSecurityLevelBits")
	}
	
	if len(entropy) > MaxEntropyBytes {
		panic("Input entropy too large")
	}
	
	if (len(entropy) * 8 * 2) < (securityLevelBits * 3) {
		panic("Insufficient entropy for security level")
	}
	
	if personalization != nil && len(personalization) > MaxPersonalizationBytes {
		panic("Personalization too long")
	}
	
	self := &HmacDrbg{
		SecurityLevelBits: securityLevelBits,
		k: make([]byte, 32),
		v: make([]byte, 32),
		reseedCounter: 1,
	}
	
	//Instantiate
	//k already holds 0x00.
	//Fill v with 0x01.
	for i := range self.v {
		self.v[i] = 0x01
	}
	
	nPers := 0
	if personalization != nil {
		nPers = len(personalization)
	}
	seed := make([]byte, len(entropy) + nPers)
	copy(seed, entropy)
	if personalization != nil {
		copy(seed[len(entropy):], personalization)
	}
	
	self.update(seed)
	
	return self
}

func (self *HmacDrbg) _hmac(key, message []byte) []byte {
	hm := hmac.New(sha256.New, key)
	hm.Write(message)
	return hm.Sum(nil)
}

func (self *HmacDrbg) update(providedData []byte) {
	nProvided := 0
	if providedData != nil {
		nProvided = len(providedData)
	}		

	msg := make([]byte, len(self.v) + 1 + nProvided)
	copy(msg, self.v)
	//leave hole with 0x00 at msg[len(self.v)]
	if (providedData != nil) {
		copy(msg[len(self.v)+1:], providedData)
	}

	self.k = self._hmac(self.k, msg)
	self.v = self._hmac(self.k, self.v)

	if providedData != nil {
		copy(msg, self.v)
		msg[len(self.v)] = 0x01
		copy(msg[len(self.v)+1:], providedData)
		
		self.k = self._hmac(self.k, msg)
		self.v = self._hmac(self.k, self.v)
	}
}

func (self *HmacDrbg) Reseed(entropy []byte) error {
	if len(entropy) * 8 < self.SecurityLevelBits {
		return errors.New("Reseed entropy is less than security-level")
	}
	
	if len(entropy) > MaxEntropyBytes {
		return errors.New("Reseed entropy exceeds MaxEntropyBytes")
	}
	
	self.update(entropy)
	self.reseedCounter = 1
	
	return nil
}

/**Fill the given byte array with random bytes.
Returns false if a reseed is necessary first.
This function will panic if the array is larger than MaxBytesPerGenerate.*/
func (self *HmacDrbg) Generate(outputBytes []byte) bool {
	nWanted := len(outputBytes)
	if nWanted > MaxBytesPerGenerate {
		panic("HmacDrbg: generate request too large.")
	}
	
	if self.reseedCounter >= 10000 {
		//set all bytes to zero, just to be clear
		for i := range outputBytes {
			outputBytes[i] = 0
		}
		return false
	}

	nGen := 0
	var n int
	for nGen < nWanted {
		self.v = self._hmac(self.k, self.v)
		
		n = nWanted - nGen
		if n > len(self.v) {
			n = len(self.v)
		}
		copy(outputBytes[nGen:], self.v[0:n])
		nGen += n
	}

	self.update(nil)
	self.reseedCounter++
	
	return true
} 

func NewHmacDrbgReader(drbg *HmacDrbg) *HmacDrbgReader {
	return &HmacDrbgReader{
		Drbg: drbg,
		buffer: make([]byte, MaxBytesPerGenerate),
		offset: MaxBytesPerGenerate,
	}
}

func (self *HmacDrbgReader) Read(b []byte) (n int, err error) {
	nRead := 0
	nWanted := len(b)
	for nRead < nWanted {
		if self.offset >= MaxBytesPerGenerate {
			if !self.Drbg.Generate(self.buffer) {
				return nRead, errors.New("MUST_RESEED")
			}
			self.offset = 0
		}
		
		b[nRead] = self.buffer[self.offset]
		nRead++
		self.offset++
	}
	
	return nRead, nil
}
