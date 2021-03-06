// +build ignore

package main

import (
	"log"
	"os"
	. "text/template"
)

var idKeys = []struct {
	ID       int
	IDPrefix string
	SKPrefix string
	IDStr    string
	SKStr    string
}{{
	ID:       1,
	IDPrefix: "0x3f, 0xbe, 0xba",
	SKPrefix: "0x4d, 0xb6, 0xc9",
	IDStr:    "id12K4tCXKcJJYxJmZ1UY9EuKPvtGVAjo32xySMKNUahbmRcsqFgW",
	SKStr:    "sk13iLKJfxNQg8vpSmjacEgEQAnXkn7rbjd5ewexc1Un5wVPa7KTk",
}, {
	ID:       2,
	IDPrefix: "0x3f, 0xbe, 0xd8",
	SKPrefix: "0x4d, 0xb6, 0xe7",
	IDStr:    "id22pNvsaMWf9qxWFrmfQpwFJiKQoWfKmBwVgQtdvqVZuqzGmrFNY",
	SKStr:    "sk22UaDys2Mzg2pUCsToo9aKgxubJFnZN5Bc2LXfV59VxMvXXKwXa",
}, {
	ID:       3,
	IDPrefix: "0x3f, 0xbe, 0xf6",
	SKPrefix: "0x4d, 0xb7, 0x05",
	IDStr:    "id33pRgpm8ufXNGxtW7n5FgdGP6afXKjU4LfVmgfC8Yaq6LyYq2wA",
	SKStr:    "sk32Xyo9kmjtNqRUfRd3ZhU56NZd8M1nR61tdBaCLSQRdhUCk4yiM",
}, {
	ID:       4,
	IDPrefix: "0x3f, 0xbf, 0x14",
	SKPrefix: "0x4d, 0xb7, 0x23",
	IDStr:    "id42vYqBB63eoSz8DHozEwtCaLbEwvBTG9pWgD3D5CCaHWy1gCjF5",
	SKStr:    "sk43eMusQuvvChoGNn1VZZwbAH8BtKJSZNC7ZWoz1Vc4Y3greLA45",
}}

func main() {
	idKeyGoTmplt := Must(ParseFiles("./idkey.tmpl"))
	idKeyTestGoTmplt := Must(ParseFiles("./idkey_test.tmpl"))

	idKeyGoFile, err := os.OpenFile("./idkey_gen.go",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer idKeyGoFile.Close()

	idKeyTestGoFile, err := os.OpenFile("./idkey_gen_test.go",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer idKeyTestGoFile.Close()

	err = idKeyGoTmplt.Execute(idKeyGoFile, idKeys)
	if err != nil {
		log.Fatal(err)
	}

	err = idKeyTestGoTmplt.Execute(idKeyTestGoFile, idKeys)
	if err != nil {
		log.Fatal(err)
	}
}
