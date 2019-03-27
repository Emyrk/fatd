package factom

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Valid Test Addresses generated by factom-walletd
	// OBVIOUSLY NEVER USE THESE FOR ANY FUNDS!
	FAAddressStr = "FA2PdKfzGP5XwoSbeW1k9QunCHwC8DY6d8xgEdfm57qfR31nTueb"
	FsAddressStr = "Fs1ipNRjEXcWj8RUn1GRLMJYVoPFBL1yw9rn6sCxWGcxciC4HdPd"
	ECAddressStr = "EC2Pawhv7uAiKFQeLgaqfRhzk5o9uPVY8Ehjh8DnLXENosvYTT26"
	EsAddressStr = "Es2tFRhAqHnydaygVAR6zbpWTQXUDaXy1JHWJugQXnYavS8ssQQE"
)

type addressUnmarshalJSONTest struct {
	Name   string
	Adr    Address
	ExpAdr Address
	Data   string
	Err    string
}

var addressUnmarshalJSONTests = []addressUnmarshalJSONTest{{
	Name: "valid FA",
	Data: fmt.Sprintf("%#v", FAAddressStr),
	Adr:  new(FAAddress),
	ExpAdr: func() *FAAddress {
		adr, _ := NewFsAddress(FsAddressStr)
		pub := adr.FAAddress()
		return &pub
	}(),
}, {
	Name: "valid Fs",
	Data: fmt.Sprintf("%#v", FsAddressStr),
	Adr:  new(FsAddress),
	ExpAdr: func() *FsAddress {
		adr, _ := NewFsAddress(FsAddressStr)
		return &adr
	}(),
}, {
	Name: "valid EC",
	Data: fmt.Sprintf("%#v", ECAddressStr),
	Adr:  new(ECAddress),
	ExpAdr: func() *ECAddress {
		adr, _ := NewEsAddress(EsAddressStr)
		pub := adr.ECAddress()
		return &pub
	}(),
}, {
	Name: "valid Es",
	Data: fmt.Sprintf("%#v", EsAddressStr),
	Adr:  new(EsAddress),
	ExpAdr: func() *EsAddress {
		adr, _ := NewEsAddress(EsAddressStr)
		return &adr
	}(),
}, {
	Name: "invalid type",
	Data: `{}`,
	Err:  "json: cannot unmarshal object into Go value of type string",
}, {
	Name: "invalid type",
	Data: `5.5`,
	Err:  "json: cannot unmarshal number into Go value of type string",
}, {
	Name: "invalid type",
	Data: `["hello"]`,
	Err:  "json: cannot unmarshal array into Go value of type string",
}, {
	Name: "invalid length",
	Data: fmt.Sprintf("%#v", FAAddressStr[0:len(FAAddressStr)-1]),
	Err:  "invalid length",
}, {
	Name: "invalid length",
	Data: fmt.Sprintf("%#v", FAAddressStr+"Q"),
	Err:  "invalid length",
}, {
	Name: "invalid prefix",
	Data: fmt.Sprintf("%#v", func() string {
		adr, _ := NewFAAddress(FAAddressStr)
		return adr.payload().StringPrefix([2]byte{0x50, 0x50})
	}()),
	Err: "invalid prefix",
}, {
	Name: "invalid prefix",
	Data: fmt.Sprintf("%#v", FsAddressStr),
	Err:  "invalid prefix",
}, {
	Name:   "invalid symbol/FA",
	Data:   fmt.Sprintf("%#v", FAAddressStr[0:len(FAAddressStr)-1]+"0"),
	Err:    "invalid format: version and/or checksum bytes missing",
	Adr:    new(FAAddress),
	ExpAdr: new(FAAddress),
}, {
	Name:   "invalid checksum",
	Data:   fmt.Sprintf("%#v", FAAddressStr[0:len(FAAddressStr)-1]+"e"),
	Err:    "checksum error",
	Adr:    new(FAAddress),
	ExpAdr: new(FAAddress),
}}

func testAddressUnmarshalJSON(t *testing.T, test addressUnmarshalJSONTest) {
	err := json.Unmarshal([]byte(test.Data), test.Adr)
	assert := assert.New(t)
	if len(test.Err) > 0 {
		assert.EqualError(err, test.Err)
		return
	}
	assert.Equal(test.ExpAdr, test.Adr)
}

func TestAddress(t *testing.T) {
	t.Run("UnmarshalJSON", func(t *testing.T) {
		for _, test := range addressUnmarshalJSONTests {
			t.Run(test.Name, func(t *testing.T) {
				if test.Adr != nil {
					testAddressUnmarshalJSON(t, test)
					return
				}
				test.ExpAdr, test.Adr = &FAAddress{}, &FAAddress{}
				t.Run("FA", func(t *testing.T) {
					testAddressUnmarshalJSON(t, test)
				})
				test.ExpAdr, test.Adr = &ECAddress{}, &ECAddress{}
				t.Run("EC", func(t *testing.T) {
					testAddressUnmarshalJSON(t, test)
				})
			})
		}
	})
	fa, _ := NewFAAddress(FAAddressStr)
	fs, _ := NewFsAddress(FsAddressStr)
	ec, _ := NewECAddress(ECAddressStr)
	es, _ := NewEsAddress(EsAddressStr)
	strToAdr := map[string]Address{FAAddressStr: fa, FsAddressStr: fs,
		ECAddressStr: ec, EsAddressStr: es}
	for adrStr, adr := range strToAdr {
		t.Run("MarshalJSON/"+adr.PrefixString(), func(t *testing.T) {
			data, err := json.Marshal(adr)
			assert := assert.New(t)
			assert.NoError(err)
			assert.Equal(fmt.Sprintf("%#v", adrStr), string(data))
		})
		t.Run("Payload/"+adr.PrefixString(), func(t *testing.T) {
			assert.EqualValues(t, adr, adr.Payload())
		})
	}

	t.Run("FsAddress", func(t *testing.T) {
		pub, _ := NewFAAddress(FAAddressStr)
		priv, _ := NewFsAddress(FsAddressStr)
		assert := assert.New(t)
		assert.Equal(pub, priv.FAAddress())
		assert.Equal(pub.PublicAddress(), priv.PublicAddress())
		assert.Equal(pub.RCDHash(), priv.RCDHash(), "RCDHash")
	})
	t.Run("EsAddress", func(t *testing.T) {
		pub, _ := NewECAddress(ECAddressStr)
		priv, _ := NewEsAddress(EsAddressStr)
		assert := assert.New(t)
		assert.Equal(pub, priv.ECAddress())
		assert.Equal(pub.PublicAddress(), priv.PublicAddress())
	})

	t.Run("New", func(t *testing.T) {
		for _, adrStr := range []string{FAAddressStr, FsAddressStr,
			ECAddressStr, EsAddressStr} {
			t.Run(adrStr, func(t *testing.T) {
				assert := assert.New(t)
				adr, err := NewAddress(adrStr)
				assert.NoError(err)
				assert.Equal(adrStr, fmt.Sprintf("%v", adr))
			})
			t.Run("Public/"+adrStr, func(t *testing.T) {
				assert := assert.New(t)
				adr, err := NewPublicAddress(adrStr)
				if adrStr[1] == 's' {
					assert.EqualError(err, "invalid prefix")
					return
				}
				assert.NoError(err)
				assert.Equal(adrStr, fmt.Sprintf("%v", adr))
			})
			t.Run("Private/"+adrStr, func(t *testing.T) {
				assert := assert.New(t)
				adr, err := NewPrivateAddress(adrStr)
				if adrStr[1] != 's' {
					assert.EqualError(err, "invalid prefix")
					return
				}
				assert.NoError(err)
				assert.Equal(adrStr, fmt.Sprintf("%v", adr))
			})
		}

		t.Run("invalid length", func(t *testing.T) {
			assert := assert.New(t)

			_, err := NewAddress("too short")
			assert.EqualError(err, "invalid length")

			_, err = NewPrivateAddress("too short")
			assert.EqualError(err, "invalid length")

			_, err = NewPublicAddress("too short")
			assert.EqualError(err, "invalid length")
		})

		t.Run("unrecognized prefix", func(t *testing.T) {
			adr, _ := NewFAAddress(FAAddressStr)
			adrStr := adr.payload().StringPrefix([2]byte{0x50, 0x50})
			assert := assert.New(t)

			_, err := NewAddress(adrStr)
			assert.EqualError(err, "unrecognized prefix")

			_, err = NewPrivateAddress(adrStr)
			assert.EqualError(err, "unrecognized prefix")

			_, err = NewPublicAddress(adrStr)
			assert.EqualError(err, "unrecognized prefix")
		})
	})

	t.Run("Generate/Fs", func(t *testing.T) {
		var err error
		fs, err = GenerateFsAddress()
		assert.NoError(t, err)
	})
	t.Run("Generate/Es", func(t *testing.T) {
		var err error
		es, err = GenerateEsAddress()
		assert.NoError(t, err)
	})

	c := NewClient()
	t.Run("Save/Fs", func(t *testing.T) {
		err := fs.Save(c)
		assert.NoError(t, err)
	})
	t.Run("Save/Es", func(t *testing.T) {
		err := es.Save(c)
		assert.NoError(t, err)
	})

	t.Run("GetPrivateAddress/Fs", func(t *testing.T) {
		assert := assert.New(t)
		_, err := fs.GetPrivateAddress(nil)
		assert.NoError(err)

		fa = fs.FAAddress()
		newFs, err := fa.GetPrivateAddress(c)
		assert.NoError(err)
		assert.Equal(fs, newFs)
	})
	t.Run("GetPrivateAddress/Es", func(t *testing.T) {
		assert := assert.New(t)
		_, err := es.GetPrivateAddress(nil)
		assert.NoError(err)

		ec = es.ECAddress()
		newEs, err := ec.GetPrivateAddress(c)
		assert.NoError(err)
		assert.Equal(es, newEs)
		assert.Equal(ec.PublicKey(), es.PublicKey())
	})

	t.Run("GetAllAddresses", func(t *testing.T) {
		adrs, err := c.GetAllAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})
	t.Run("GetAllPrivateAddresses", func(t *testing.T) {
		adrs, err := c.GetAllPrivateAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})
	t.Run("GetAllFAAddresses", func(t *testing.T) {
		adrs, err := c.GetAllFAAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})
	t.Run("GetAllFsAddresses", func(t *testing.T) {
		adrs, err := c.GetAllFsAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})
	t.Run("GetAllECAddresses", func(t *testing.T) {
		adrs, err := c.GetAllECAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})
	t.Run("GetAllEsAddresses", func(t *testing.T) {
		adrs, err := c.GetAllEsAddresses()
		assert := assert.New(t)
		assert.NoError(err)
		assert.NotEmpty(adrs)
	})

	t.Run("Remove/Fs", func(t *testing.T) {
		err := fs.Remove(c)
		assert.NoError(t, err)
	})
	t.Run("Remove/Es", func(t *testing.T) {
		err := es.Remove(c)
		assert.NoError(t, err)
	})
}
