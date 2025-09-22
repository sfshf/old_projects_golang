package main

import (
	. "github.com/nextsurfer/word/scripts/book"
)

func main() {
	dbPath := "root:waf12KFkwo2@tcp(127.0.0.1:3306)/word?charset=utf8&interpolateParams=True&parseTime=true"

	// GenerateModels(dbPath)

	// TODO special pronunciation for some words to generate SSML speech.

	// csvPath := "/Users/lxm/Documents/Canada/WordDict/eva/A20721.csv"
	// AddBook(csvPath, dbPath, "A2", "CEFR A1, Cambridge version")

	// csvPath := "/Users/lxm/Documents/Canada/WordDict/eva/A20716.csv"
	// AddBook(csvPath, dbPath, "A2", "CEFR A2, Cambridge version")

	csvPath := "/Users/lxm/Documents/Canada/WordDict/eva/B1.csv"
	AddBook(csvPath, dbPath, "B1", "CEFR B1, Cambridge version")

	// bookID := 100000001
	// exportedFilePath := "/Users/lxm/Documents/Canada/WordDict/ExportedBook1.csv"
	// ExportBook(int64(bookID), exportedFilePath, dbPath)

	// bookID := int64(100000001)
	// encryptedFilePath := "/Users/lxm/Documents/Canada/WordDict/TestBook1"
	// ExportEncryptedBook(bookID, dbPath, encryptedFilePath)

	// csvPath := "/Users/lxm/Documents/Canada/WordDict/testA2.csv"
	// AddBook(csvPath, dbPath, "test A2", "bundled book")

	// bookID := 100000003
	// exportedFilePath := "/Users/lxm/Documents/Canada/WordDict/eva/Exported2.csv"
	// ExportBook(int64(bookID), exportedFilePath, dbPath)

	// bookID := int64(100000008)
	// encryptedFilePath := "/Users/lxm/Documents/Canada/WordDict/eva/A1"
	// ExportEncryptedBook(bookID, dbPath, encryptedFilePath)

	// bookID := int64(100000016)
	// encryptedFilePath := "/Users/lxm/Documents/Canada/WordDict/eva/A2"
	// ExportEncryptedBook(bookID, dbPath, encryptedFilePath)

	bookID := int64(100000019)
	encryptedFilePath := "/Users/lxm/Documents/Canada/WordDict/eva/B1"
	ExportEncryptedBook(bookID, dbPath, encryptedFilePath)

	// md5 := "2a0265f49571f91b2520d422f4417dfa"
	// downloadURL := "https://n1xt-test.s3.amazonaws.com/" + md5
	// UpdateDownloadURL(bookID, downloadURL, dbPath)

}

// import (
// 	"fmt"
// 	"regexp"
// 	"strings"
// )

// func main() {
// 	// test regex
// 	str := "a"
// 	exampleStr := "A cheetah can run faster than a lion."
// 	example := []byte(exampleStr)
// 	//
// 	re := regexp.MustCompile(`(?i)(^|\s|[^\w\s])` + str + `($|\s|[^\w\s])`)
// 	matched := re.FindAllIndex(example, -1)
// 	fmt.Println(matched)

// 	for i := 0; i < len(matched); i++ {
// 		positions := matched[i]
// 		positions[1] -= positions[0]
// 		if positions[1] == len(str)+1 {
// 			// remove space or punctuation
// 			matchedStr := exampleStr[positions[0]:(positions[0] + positions[1])]
// 			index := strings.Index(strings.ToLower(matchedStr), strings.ToLower(str))
// 			positions[1] -= 1
// 			if index == 1 {
// 				positions[0] += 1
// 			}
// 		} else if positions[1] == len(str)+2 {
// 			positions[0] += 1
// 			positions[1] -= 2
// 		}
// 		// else {
// 		// 	// positions[1] == len(str)
// 		// }
// 		fmt.Println(positions)
// 	}
// }
