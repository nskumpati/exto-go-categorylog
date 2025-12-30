package file_utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gaeaglobal/exto/server/app"
	"github.com/gaeaglobal/exto/server/utils"
	"github.com/rs/xid"
)

func GetFilePath(appCtx *app.AppContext, reqCtx *app.RequestContext, dirPath string, fileName string, uniqueIDLength int, attempt int) string {
	if attempt > 5 {
		log.Printf("Failed to find a unique file path after %d attempts", attempt)
		return GetFilePath(appCtx, reqCtx, dirPath, fileName, uniqueIDLength+1, 0)
	}
	if uniqueIDLength > 10 {
		log.Printf("Unique ID length exceeded maximum limit, using original file name")
		return ""
	}
	orgIDString := reqCtx.User.OrganizationID.Hex()
	nanoID := xid.New()
	uniqueFileName := fmt.Sprintf("%s_%s", nanoID.String()[:uniqueIDLength], strings.ToLower(fileName))
	filePath := filepath.Join(dirPath, orgIDString, uniqueFileName)
	if utils.FileExists(filePath) {
		return GetFilePath(appCtx, reqCtx, dirPath, fileName, uniqueIDLength, attempt+1)
	}
	dirPath = filepath.Join(dirPath, orgIDString)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return ""
	}
	return filePath
}
