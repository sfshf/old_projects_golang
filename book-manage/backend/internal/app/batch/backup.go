package batch

import (
	"os"

	"github.com/nextsurfer/book-manage-api/internal/app/dao"
	. "github.com/nextsurfer/book-manage-api/internal/app/model"
)

func RegainBackup(filepath string, book *Book, dbmanager *dao.Manager, _logger Logger) error {
	logger = _logger

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = readCSVToDatabase(true, f, book, dbmanager, false, true); err != nil {
		logger.InfoPrint(err)
		logger.Complete(err, "admin", "RegainBook", book.ID)
		return err
	}
	logger.InfoPrint("Book ID: ", book.ID)
	logger.Complete(nil, "admin", "RegainBook", book.ID)
	return err
}
