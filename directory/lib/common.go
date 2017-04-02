package directory

import (
	"encoding/json"
	"fmt"
	sindexd "moses/sindexd/lib"

	files "moses/user/files/lib"
)

var (
	PubDate  string
	id       string
	cos      int
	volid    int
	specific int
	Action   string

	// Dispersion            Objectid                      Volid         service       speficic      COS
	// A4 8B EB      F7 11 D9 DB 02 FB B3 A8           00 00 00 05         03          00 00 2A       20
	// "CN", "CA", "DE", "EP", "FR", "GB", "JP", "KR", "US", "WO", "IT", "RU", "TW", "SU", "NL",
	// "NO", "PL", "AT", "MX", "IL", "ZA", "NZ", "FI", "ES", "DK", "DD", "CH", "BE", "AU", "BR", "OTHER"
	/*
			PnoidSpec = map[string][]string{
				"CN":    {"EC3FA646EC5EA65557A7AE000000050300002A20", "2", "5", "42"},
				"CA":    {"E73D08320D7F6B1A14A040000000060300002A20", "2", "6", "42"},
				"DE":    {"1868FEA35EAB74AC12221C000000070300002A20", "2", "7", "42"},
				"EP":    {"36027CB71DD48F34859573000000080300002A20", "2", "8", "42"},
				"FR":    {"A663650EE44EB7F11BD846000000090300002A20", "2", "9", "42"},
				"GB":    {"0480A9A5A0FA3E4D45E7A90000000A0300002A20", "2", "10", "42"},
				"JP":    {"D797836876A93BFB7243BB0000000B0300002A20", "2", "11", "42"},
				"KR":    {"9A849C03A7FBAD4890E18F0000000C0300002A20", "2", "12", "42"},
				"US":    {"84026C20C2F3F89099B3C30000000D0300002A20", "2", "13", "42"},
				"WO":    {"5919B9D8FF2C55070766270000000E0300002A20", "2", "14", "42"},
				"IT":    {"565BD93BE78AB007B01F210000000F0300002A20", "2", "15", "42"},
				"RU":    {"AAF5F58F96794400525B0D000000100300002A20", "2", "16", "42"},
				"TW":    {"C0AD8021C0C89818CD40B7000000110300002A20", "2", "17", "42"},
				"SU":    {"96F0345A311F974E84BE4F000000120300002A20", "2", "18", "42"},
				"NL":    {"48DA4C63319FD87C967001000000130300002A20", "2", "19", "42"},
				"NO":    {"078E233A81132001665F71000000140300002A20", "2", "20", "42"},
				"PL":    {"1905029B905D099E51F911000000150300002A20", "2", "21", "42"},
				"AT":    {"5D06CECA381929BB730E45000000160300002A20", "2", "22", "42"},
				"MX":    {"2D5A1E930C9B54B3385A51000000170300002A20", "2", "23", "42"},
				"IL":    {"6C8466B601CEE42CEF50EA000000180300002A20", "2", "24", "42"},
				"ZA":    {"54CC3D3FDC9206AC9059DA000000190300002A20", "2", "25", "42"},
				"NZ":    {"FB2F65D05F4155F57B68180000001A0300002A20", "2", "26", "42"},
				"FI":    {"84285869492FD2D715AE670000001B0300002A20", "2", "27", "42"},
				"ES":    {"15E12D43D753F6C2AB5C1F0000001C0300002A20", "2", "28", "42"},
				"DK":    {"648CF8B9E193EFC66A95470000001D0300002A20", "2", "29", "42"},
				"DD":    {"9F10FA0E0B60440ABE7BA00000001E0300002A20", "2", "30", "42"},
				"CH":    {"F014ACF9EE5B6E94D4A3B00000001F0300002A20", "2", "31", "42"},
				"BE":    {"54A95985A2567EFD80D203000000200300002A20", "2", "32", "42"},
				"AU":    {"AD4404E5A5EF52385A6835000000210300002A20", "2", "33", "42"},
				"BR":    {"AF881A336798CDDEE13CFC000000220300002A20", "2", "34", "42"},
				"OTHER": {"701EF3C6FC5DBBC5158975000000230300002A20", "2", "35", "42"},
			}
			PdoidSpec = map[string][]string{
				"CN":    {"EC3FA646EC5EA65557A7AE000000320300002A20", "2", "50", "42"},
				"CA":    {"E73D08320D7F6B1A14A040000000330300002A20", "2", "51", "42"},
				"DE":    {"1868FEA35EAB74AC12221C000000340300002A20", "2", "52", "42"},
				"EP":    {"36027CB71DD48F34859573000000350300002A20", "2", "53", "42"},
				"FR":    {"A663650EE44EB7F11BD846000000360300002A20", "2", "54", "42"},
				"GB":    {"0480A9A5A0FA3E4D45E7A9000000370300002A20", "2", "55", "42"},
				"JP":    {"D797836876A93BFB7243BB000000380300002A20", "2", "56", "42"},
				"KR":    {"9A849C03A7FBAD4890E18F000000390300002A20", "2", "57", "42"},
				"US":    {"84026C20C2F3F89099B3C30000003A0300002A20", "2", "58", "42"},
				"WO":    {"5919B9D8FF2C55070766270000003B0300002A20", "2", "59", "42"},
				"IT":    {"565BD93BE78AB007B01F210000003C0300002A20", "2", "60", "42"},
				"RU":    {"AAF5F58F96794400525B0D0000003D0300002A20", "2", "61", "42"},
				"TW":    {"C0AD8021C0C89818CD40B70000003E0300002A20", "2", "62", "42"},
				"SU":    {"96F0345A311F974E84BE4F0000003F0300002A20", "2", "63", "42"},
				"NL":    {"48DA4C63319FD87C967001000000400300002A20", "2", "64", "42"},
				"NO":    {"078E233A81132001665F71000000410300002A20", "2", "65", "42"},
				"PL":    {"1905029B905D099E51F911000000420300002A20", "2", "66", "42"},
				"AT":    {"5D06CECA381929BB730E45000000430300002A20", "2", "67", "42"},
				"MX":    {"2D5A1E930C9B54B3385A51000000440300002A20", "2", "68", "42"},
				"IL":    {"6C8466B601CEE42CEF50EA000000450300002A20", "2", "69", "42"},
				"ZA":    {"54CC3D3FDC9206AC9059DA000000460300002A20", "2", "70", "42"},
				"NZ":    {"FB2F65D05F4155F57B6818000000470300002A20", "2", "71", "42"},
				"FI":    {"84285869492FD2D715AE67000000480300002A20", "2", "72", "42"},
				"ES":    {"15E12D43D753F6C2AB5C1F000000490300002A20", "2", "73", "42"},
				"DK":    {"648CF8B9E193EFC66A95470000004A0300002A20", "2", "74", "42"},
				"DD":    {"9F10FA0E0B60440ABE7BA00000004B0300002A20", "2", "75", "42"},
				"CH":    {"F014ACF9EE5B6E94D4A3B00000004C0300002A20", "2", "76", "42"},
				"BE":    {"54A95985A2567EFD80D2030000004D0300002A20", "2", "77", "42"},
				"AU":    {"AD4404E5A5EF52385A68350000004E0300002A20", "2", "78", "42"},
				"BR":    {"AF881A336798CDDEE13CFC0000004F0300002A20", "2", "79", "42"},
				"OTHER": {"701EF3C6FC5DBBC5158975000000500300002A20", "2", "80", "42"},
			}
		)
	*/

	PnoidSpec = "/etc/moses/sindexd-prod-pn.json"
	PdoidSpec = "/etc/moses/sindexd-prod-pd.json"
)

type HttpResponse struct { // used for get prefix
	Pref     string
	Response *sindexd.Response
	Err      error
}

const (
	Limit = 500
)

func BuildIndexspec(file string) map[string]*sindexd.Index_spec {

	m := make(map[string]*sindexd.Index_spec)
	if scanner, err := files.Scanner(file); err != nil {
		fmt.Println(scanner, err)
	} else if linea, err := files.ScanLines(scanner, 100); err == nil {
		index := sindexd.IndexTab{}
		for _, v := range linea {
			if err = json.Unmarshal([]byte(v), &index); err == nil {
				m[index.Country] = &sindexd.Index_spec{
					Index_id: index.Index_id,
					Cos:      index.Cos,
					Vol_id:   int(index.Volid),
					Specific: int(index.Specific),
				}
			} else {
				fmt.Println(err)
			}

		}
	}
	return m
}

func GetIndexSpec(iIndex string) map[string]*sindexd.Index_spec {
	switch iIndex {
	case "PN":
		// return BuildIndexSpec(PnoidSpec)
		return BuildIndexspec(PnoidSpec)
	case "PD":
		// return BuildIndexSpec(PdoidSpec)
		return BuildIndexspec(PdoidSpec)
	default:
		return nil
	}
}
