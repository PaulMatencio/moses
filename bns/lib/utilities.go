package bns

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func ST33toFiles(inputFile string, outputusermdDir string, outputTiffDir string, outputContainerDir string, combine bool) error {

	//   EXTRACT ST33 Files
	//
	//  FOR EACH DOCUMENT in ST33 {
	// 		CREATE  page metadata			 ( ST33 HEADER  ++)
	// 		CREATE  page data  (Tiff)      ( ST33 TIFF RECORDS )
	//
	// 		FOR EACH PAGE of a document {
	// 			 if combine  {
	//      		  COMBINE  page data (TIFF) and metadata => PAGE Struct
	//      		  WRITE PAGE struct
	//       	 }
	//  		else {
	//     		  	  WRITE data ( TIFF)
	//	     		  WRITE Metadata  ( user metadata)
	//       	}
	//      }
	//      CREATE DOCUMENT metadata
	//      WRITE DOCUMENT  metadata
	//  }
	//

	//   REMOVED  =>  Check old bns for how to do
	error := errors.New("Function has been removed")
	return error
}

func ST33toFiles_p(inputFile string, outputusermdDir string, outputTiffDir string, outputContainerDir string, combine bool) error {

	//   EXTRACT ST33 Files
	//
	//  FOR EACH DOCUMENT in ST33 {
	// 		CREATE  page metadata			 ( ST33 HEADER  ++)
	// 		CREATE  page data  (Tiff)      ( ST33 TIFF RECORDS )
	//
	// 		FOR EACH PAGE of a document {
	// 			 if combine  {
	//      		  COMBINE  page data (TIFF) and metadata => PAGE Struct
	//      		  WRITE PAGE struct
	//       	 }
	//  		else {
	//     		  	  WRITE data ( TIFF)
	//	     		  WRITE Metadata  ( user metadata)
	//       	}
	//      }
	//      CREATE DOCUMENT metadata
	//      WRITE DOCUMENT  metadata
	//  }
	//

	//   REMOVED  =>  Check old bns for how to do
	error := errors.New("Function has been removed")
	return error
}

func ParseDate(str string) (dd Date, err error) {
	str = strings.TrimSpace(str)
	var (
		y, m, d int
	)
	if len(str) != 8 {
		goto invalid
	}
	if y, err = strconv.Atoi(str[0:4]); err != nil {
		return
	}
	if m, err = strconv.Atoi(str[4:6]); err != nil {
		return
	}
	if m < 1 || m > 12 {
		goto invalid
	}
	if d, err = strconv.Atoi(str[6:8]); err != nil {
		return
	}
	if d < 1 || d > 31 {
		goto invalid
	}
	dd.Year = int16(y)
	dd.Month = byte(m)
	dd.Day = byte(d)
	return
invalid:
	err = errors.New("Invalid metadata Date string: " + str)
	return
}

func (dd Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", dd.Year, dd.Month, dd.Day)
}

func noDate() Date {
	//return Date{int16(0),byte(0),byte(0)}
	return Date{}
}

func getuint16(in []byte) uint16 {
	out, _ := strconv.Atoi(string(in))
	return uint16(out)

}

func getuint32(in []byte) uint32 {
	out, _ := strconv.Atoi(string(in))
	return uint32(out)

}

func getConfig(configfile string) (Configuration, error) {
	cfile, err := os.Open(configfile)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(cfile)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	_ = cfile.Close()
	return configuration, err
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

/* image orientation */

func getOrientation(rotation_code []byte) uint16 {
	orientation, _ := strconv.Atoi(string(rotation_code))
	switch orientation {
	case 1:
		return uint16(1)
	case 2:
		return uint16(6)
	case 3:
		return uint16(3)
	case 4:
		return uint16(8)
	default:
		return uint16(1)
	}
}

func Tiff2Png(tiffile, pngfile string) error {
	// cmd := exec.Command("convert", "-resize", "950x", tiffile, pngfile)
	cmd := exec.Command("convert", tiffile, pngfile)
	return cmd.Run()

}

func RemoveSlash(input string) string {
	output := ""
	ar := strings.Split(input, "/")
	for _, word := range ar {
		output = output + word
	}
	return output
}
